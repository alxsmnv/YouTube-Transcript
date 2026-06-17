package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strings"
)

const (
	IOSClientVersion = "20.10.38"
	IOSUserAgent     = "com.google.ios.youtube/20.10.38 (iPhone16,2; U; CPU iOS 17_5_1 like Mac OS X; en_US)"
	WebUserAgent     = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: youtube-transcript <youtube-url>")
		fmt.Println("Example: youtube-transcript https://www.youtube.com/watch?v=dQw4w9WgXcQ")
		os.Exit(1)
	}

	videoID := extractVideoID(os.Args[1])
	if videoID == "" {
		fmt.Println("Error: Could not extract video ID from URL")
		os.Exit(1)
	}

	transcript, err := getTranscript(videoID, "en")
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(transcript)
}

func extractVideoID(ytURL string) string {
	re := regexp.MustCompile(`(?:v=|/)([a-zA-Z0-9_-]{11})`)
	matches := re.FindStringSubmatch(ytURL)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

type TranscriptLine struct {
	Text  string
	Start float64
	Dur   float64
}

func getTranscript(videoID, lang string) (string, error) {
	// Step 1: POST to /youtubei/v1/player with IOS client
	playerPayload := map[string]interface{}{
		"context": map[string]interface{}{
			"client": map[string]interface{}{
				"clientName":    "IOS",
				"clientVersion": IOSClientVersion,
				"hl":            "en",
				"gl":            "US",
			},
		},
		"videoId": videoID,
	}

	jsonData, err := json.Marshal(playerPayload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal payload: %v", err)
	}

	req, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/player", strings.NewReader(string(jsonData)))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", IOSUserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch player API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("player API returned %s: %s", resp.Status, string(body))
	}

	var playerResponse map[string]interface{}
	if err := json.Unmarshal(body, &playerResponse); err != nil {
		return "", fmt.Errorf("failed to parse player response: %v", err)
	}

	// Check for API errors
	if errVal, ok := playerResponse["error"]; ok {
		errMap, ok := errVal.(map[string]interface{})
		if ok {
			return "", fmt.Errorf("YouTube player API error: %v", errMap["message"])
		}
		return "", fmt.Errorf("YouTube player API error")
	}

	// Step 2: Extract caption track URL from player response
	captions, ok := playerResponse["captions"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no captions in response")
	}

	playerCaptions, ok := captions["playerCaptionsTracklistRenderer"].(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("no playerCaptionsTracklistRenderer in response")
	}

	captionTracksRaw, ok := playerCaptions["captionTracks"].([]interface{})
	if !ok {
		return "", fmt.Errorf("no captionTracks in response")
	}

	if len(captionTracksRaw) == 0 {
		return "", fmt.Errorf("no transcript available for this video")
	}

	// Find the requested language or fallback to first available
	var selectedTrack map[string]interface{}
	for _, track := range captionTracksRaw {
		trackMap, ok := track.(map[string]interface{})
		if !ok {
			continue
		}
		if langCode, ok := trackMap["languageCode"].(string); ok && langCode == lang {
			selectedTrack = trackMap
			break
		}
	}

	// Fallback to first track if language not found
	if selectedTrack == nil {
		if firstTrack, ok := captionTracksRaw[0].(map[string]interface{}); ok {
			selectedTrack = firstTrack
			if langCode, ok := selectedTrack["languageCode"].(string); ok {
				fmt.Printf("Note: Language '%s' not available, using '%s'\n", lang, langCode)
			}
		}
	}

	if selectedTrack == nil {
		return "", fmt.Errorf("no valid caption track found")
	}

	// Step 3: Get the baseUrl and append &fmt=json3
	baseURL, ok := selectedTrack["baseUrl"].(string)
	if !ok {
		return "", fmt.Errorf("no baseUrl in caption track")
	}

	timedtextURL := baseURL + "&fmt=json3"

	// Step 4: GET the timedtext JSON
	req2, err := http.NewRequest("GET", timedtextURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create timedtext request: %v", err)
	}
	req2.Header.Set("User-Agent", WebUserAgent)

	resp2, err := client.Do(req2)
	if err != nil {
		return "", fmt.Errorf("failed to fetch transcript: %v", err)
	}
	defer resp2.Body.Close()

	body2, err := io.ReadAll(resp2.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read transcript: %v", err)
	}

	if resp2.StatusCode != 200 {
		return "", fmt.Errorf("transcript API returned %s: %s", resp2.Status, string(body2))
	}

	// Step 5: Parse the timedtext JSON (JSON3 format)
	var timedtextResponse map[string]interface{}
	if err := json.Unmarshal(body2, &timedtextResponse); err != nil {
		return "", fmt.Errorf("failed to parse transcript: %v", err)
	}

	eventsRaw, ok := timedtextResponse["events"].([]interface{})
	if !ok {
		return "", fmt.Errorf("no events in transcript")
	}

	var lines []TranscriptLine
	for _, event := range eventsRaw {
		eventMap, ok := event.(map[string]interface{})
		if !ok {
			continue
		}

		// Skip non-text events (window/pen definitions)
		segsRaw, ok := eventMap["segs"].([]interface{})
		if !ok || len(segsRaw) == 0 {
			continue
		}

		// Build text from segs[].utf8
		var text strings.Builder
		for _, seg := range segsRaw {
			segMap, ok := seg.(map[string]interface{})
			if !ok {
				continue
			}
			if utf8, ok := segMap["utf8"].(string); ok {
				text.WriteString(utf8)
			}
		}

		lineText := text.String()
		lineText = strings.ReplaceAll(lineText, "\n", " ")
		lineText = strings.TrimSpace(lineText)

		if lineText == "" {
			continue
		}

		// Extract timing
		startMs, _ := eventMap["tStartMs"].(float64)
		durationMs, _ := eventMap["dDurationMs"].(float64)

		lines = append(lines, TranscriptLine{
			Text:  lineText,
			Start: startMs / 1000,
			Dur:   durationMs / 1000,
		})
	}

	if len(lines) == 0 {
		return "", fmt.Errorf("no transcript text found")
	}

	// Step 6: Format output with timestamps
	var sb strings.Builder
	for _, line := range lines {
		minutes := int(line.Start) / 60
		seconds := int(line.Start) % 60
		sb.WriteString(fmt.Sprintf("[%02d:%02d] %s\n", minutes, seconds, line.Text))
	}

	return sb.String(), nil
}

