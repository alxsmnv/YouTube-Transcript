# YouTube Transcript

A blazing-fast, single-binary Go tool to extract transcripts from YouTube videos. No bloat, no dependencies, just works.

## Features

- **Single Binary**: 10MB standalone executable, no runtime dependencies
- **Cross-Platform**: Compiles on Linux, macOS, Windows, and more
- **Blazing Fast**: Pure Go implementation, no Node.js, no Python, no bloat
- **Open Source**: Licensed under GNU GPL v3 - freedom for everyone
- **Smart Extraction**: Uses YouTube's hidden iOS API endpoint for reliable transcript access
- **Timestamped Output**: Each line includes `[MM:SS]` timestamps
- **Language Support**: Auto-detects available languages, falls back gracefully

## Why This Tool?

Most YouTube transcript tools are bloated with dependencies:
- Node.js + npm packages (500MB+)
- Python with virtualenvs and pip
- Docker containers when you just need a simple binary

This tool is different:
- **10MB binary** vs 500MB+ of dependencies
- **Works everywhere** - single binary, no setup
- **Instant execution** - no package installation, no environment setup
- **Fully open source** - you own your tools

## Installation

### From Source (Recommended)

```bash
git clone <repository-url>
cd youtube-transcript
go build -o youtube-transcript main.go
```

### Pre-built Binaries

Download the latest release for your platform from the [Releases page](https://github.com/yourusername/youtube-transcript/releases).

## Usage

```bash
./youtube-transcript <youtube-url>
```

### Examples

```bash
# Extract transcript from a video
./youtube-transcript https://www.youtube.com/watch?v=dQw4w9WgXcQ

# Works with short URLs too
./youtube-transcript https://youtu.be/dQw4w9WgXcQ
```

### Output Format

```
[00:00] Welcome to this video about something interesting.
[00:05] Today we're going to explore the fundamentals of the topic.
[00:12] Let's dive right in and see what we can learn together.
```

Each line includes:
- `[MM:SS]` - Timestamp in minutes and seconds
- Text content from the transcript

## How It Works

1. **Extracts Video ID** from the provided YouTube URL
2. **Queries YouTube's iOS API** - presents as an iPhone client to access the hidden endpoint
3. **Retrieves Caption Track** - finds available transcript languages
4. **Downloads Transcript** - fetches the timedtext JSON data
5. **Formats Output** - parses and displays with timestamps

## Building from Source

### Prerequisites

- Go 1.20 or later
- Git (for cloning the repository)

### Build Commands

```bash
# Build for current platform
go build -o youtube-transcript main.go

# Build for specific platforms (cross-compilation)
GOOS=linux GOARCH=amd64 go build -o youtube-transcript-linux-amd64 main.go
GOOS=darwin GOARCH=amd64 go build -o youtube-transcript-darwin-amd64 main.go
GOOS=windows GOARCH=amd64 go build -o youtube-transcript-windows-amd64.exe main.go

# Build for ARM (Apple Silicon, Raspberry Pi)
GOOS=darwin GOARCH=arm64 go build -o youtube-transcript-darwin-arm64 main.go
GOOS=linux GOARCH=arm64 go build -o youtube-transcript-linux-arm64 main.go
```

## Error Handling

The tool provides clear error messages for common issues:

- **Invalid URL**: "Error: Could not extract video ID from URL"
- **No Transcript**: "Error: no transcript available for this video"
- **Language Not Available**: "Note: Language 'xx' not available, using 'yy'"
- **Network Errors**: Detailed error messages with HTTP status codes

## License

**GNU General Public License v3.0 (GPL-3.0)**

This is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

**Key points:**
- You are free to use, modify, and distribute this software
- Any modifications must be released under the same license (copyleft)
- Source code of any modifications must be made publicly available
- You cannot use this code in proprietary/closed-source projects

See [LICENSE](LICENSE) for the full license text.

## Contributing

Contributions are welcome! This project is community-driven and licensed under GPL-3.0, which means:

1. All contributions must be made under the GPL-3.0 license
2. Pull requests should include clear descriptions of changes
3. Code should follow Go best practices and formatting (`go fmt`)
4. Tests should be added for new features

## Disclaimer

This tool is for educational and personal use. YouTube's Terms of Service may restrict automated access to their platform. Use responsibly and respect content creators' rights.

This tool is not affiliated with, endorsed by, or connected to Google/YouTube in any way.

## Credits

Built with ❤️ using Go, because sometimes you just need a simple tool that works.
