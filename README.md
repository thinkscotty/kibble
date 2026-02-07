# Kibble

Kibble is a lightweight, AI-powered facts generator that runs on minimal hardware like a Raspberry Pi. It uses Google's Gemini Flash 2.5 AI to generate interesting facts about topics you choose, and displays them through a clean web interface.

## What Does Kibble Do?

1. **You add topics** (e.g., "Space", "Ancient History", "Marine Biology") with brief descriptions
2. **Kibble asks an AI** to generate interesting facts about each topic
3. **Facts are displayed** on a dashboard you can view from any device on your network
4. **External devices** (LED tickers, smart displays) can fetch facts via a simple API

Kibble automatically refreshes facts on a schedule you set, and it's smart enough to avoid generating duplicate facts.

## Requirements

- A computer to run Kibble on (works great on a Raspberry Pi 3B+ or newer)
- A free Google Gemini API key (instructions below)

## Getting Your Gemini API Key (Free)

1. Go to [Google AI Studio](https://aistudio.google.com/apikey)
2. Sign in with your Google account
3. Click "Create API Key"
4. Copy the key — you'll paste it into Kibble's Settings page

The free tier is generous and more than enough for personal use.

## Installation

### Option 1: Download a Pre-Built Binary (Recommended)

1. Go to the [Releases page](https://github.com/thinkscotty/kibble/releases/latest)
2. Download the file that matches your system:

| System | File to Download |
|--------|------------------|
| **Raspberry Pi 3B+/4/5** (64-bit OS) | `kibble-linux-arm64` |
| **Raspberry Pi** (32-bit OS) | `kibble-linux-arm` |
| **Linux PC** | `kibble-linux-amd64` |
| **Mac (Apple Silicon)** | `kibble-darwin-arm64` |

3. Make it executable and run it:

```bash
chmod +x kibble-linux-arm64
./kibble-linux-arm64
```

> **Tip for Raspberry Pi users:** If you're not sure whether you're running 32-bit or 64-bit, run `uname -m` in a terminal. If it says `aarch64`, download the `arm64` version. If it says `armv7l`, download the `arm` version.

### Option 2: Build From Source

You'll need [Go 1.23+](https://go.dev/dl/) installed.

```bash
# Clone the repository
git clone https://github.com/thinkscotty/kibble.git
cd kibble

# Build for your current system
make build

# Or cross-compile for Raspberry Pi
make build-arm64

# The binary will be in the bin/ directory
./bin/kibble
```

## Running Kibble

Simply run the binary:

```bash
./kibble
```

Kibble will:
1. Create a `kibble.db` database file in the current directory
2. Start a web server on port 8080

Open your browser and go to: **http://localhost:8080** (or replace `localhost` with your Pi's IP address if accessing from another device).

### Configuration (Optional)

You can create a `config.yaml` file next to the binary to customize settings:

```yaml
server:
  host: "0.0.0.0"     # Listen on all interfaces (default)
  port: 8080           # Web server port

database:
  path: "./kibble.db"  # Database file location

logging:
  level: "info"        # debug, info, warn, error

similarity:
  threshold: 0.6       # How similar facts must be to be considered duplicates (0.0-1.0)
  ngram_size: 3        # Trigram size for similarity comparison
```

All of these have sensible defaults, so the config file is entirely optional.

## Using Kibble

### First-Time Setup

1. Go to the **Settings** page
2. Paste your Gemini API key and click "Test Key" to verify it works
3. Click "Save Settings"

### Adding Topics

1. Go to the **Topics** page
2. Enter a topic name (e.g., "Space Exploration")
3. Optionally add a description to guide the AI (e.g., "Focus on recent discoveries and missions")
4. Set how many facts to generate per refresh (default: 5)
5. Set the refresh interval in minutes (default: 1440 = 24 hours)
6. Click "Add Topic"

### Viewing Facts

- The **Dashboard** shows cards for each active topic with their latest facts
- Click "Refresh" on any card to generate new facts immediately

### Managing Facts

- On the **Topics** page, use the search bar to find specific facts
- Click "Edit" to modify any fact, or "Delete" to remove it
- Add your own custom facts using the "Add Custom Fact" form

### Customizing Appearance

On the **Settings** page you can:
- Switch between **dark mode** and **light mode**
- Adjust text size (small, medium, large)
- Set the number of card columns on the dashboard
- Set how many facts to display per topic

### AI Instructions

On the **Settings** page you can give the AI custom instructions:
- **Custom Instructions**: Guide what kind of facts to generate (e.g., "Focus on lesser-known facts", "Include recent discoveries")
- **Tone & Style**: Control how facts are written (e.g., "Keep facts concise, under 2 sentences", "Use a casual, friendly tone")

## External Device API

Kibble provides a simple JSON API for external devices like LED matrix displays:

### Get Active Topics
```
GET /api/v1/topics
```

### Get Facts for a Topic
```
GET /api/v1/facts?topic_id=1&limit=5
```

### Get a Random Fact
```
GET /api/v1/facts/random
```

This returns a single random fact from any active topic — perfect for scrolling tickers.

## Running as a Service (Raspberry Pi)

To run Kibble automatically on boot, create a systemd service:

```bash
sudo nano /etc/systemd/system/kibble.service
```

Paste the following (adjust paths as needed):

```ini
[Unit]
Description=Kibble Facts Generator
After=network.target

[Service]
Type=simple
User=pi
WorkingDirectory=/home/pi/kibble
ExecStart=/home/pi/kibble/kibble
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Then enable and start it:

```bash
sudo systemctl enable kibble
sudo systemctl start kibble
```

Check status with: `sudo systemctl status kibble`

## Troubleshooting

- **"API key not configured"**: Go to Settings and enter your Gemini API key
- **Facts not generating**: Click "Test Key" on the Settings page to verify your API key works
- **Page not loading**: Make sure nothing else is using port 8080, or change the port in `config.yaml`
- **Can't access from another device**: Make sure you're using the Pi's IP address (not `localhost`) and that both devices are on the same network

## License

MIT
