# Kibble

Kibble is a lightweight, AI-powered facts generator web application. It uses Google's Gemini Flash 2.5 AI to generate interesting facts about topics you choose, caches them in SQLite, and displays them through a clean, themeable web interface. Facts are also served to external client devices (LED matrices, smart displays) via a JSON API.

Deploys as a single binary with no external dependencies.

## What Does Kibble Do?

1. **You add topics** (e.g., "Space", "Ancient History", "Marine Biology") with brief descriptions
2. **Kibble asks an AI** to generate interesting facts about each topic
3. **Facts are displayed** on a themeable dashboard you can view from any device
4. **External devices** (LED tickers, smart displays) can fetch facts via a simple JSON API
5. **Updates itself** from the Settings page with one click

Kibble automatically refreshes facts on a schedule you set, and it's smart enough to avoid generating duplicate facts.

## Requirements

- A server or computer to run Kibble on (VPS, Raspberry Pi 3B+, or any Linux/macOS machine)
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
| **Linux VPS / Cloud Server** (x86_64) | `kibble-linux-amd64` |
| **Raspberry Pi 3B+/4/5** (64-bit OS) | `kibble-linux-arm64` |
| **Raspberry Pi** (32-bit OS) | `kibble-linux-arm` |
| **Mac (Apple Silicon)** | `kibble-darwin-arm64` |

3. Install the binary:

```bash
# Download (replace URL with the latest release and your platform)
wget https://github.com/thinkscotty/kibble/releases/latest/download/kibble-linux-amd64

# Install to a standard location
sudo cp kibble-linux-amd64 /usr/local/bin/kibble
sudo chmod +x /usr/local/bin/kibble

# Create a data directory
sudo mkdir -p /var/lib/kibble
```

4. Run it:

```bash
cd /var/lib/kibble
kibble
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

Simply run the binary from the directory where you want the database stored:

```bash
cd /var/lib/kibble
kibble
```

Kibble will:
1. Create a `kibble.db` database file in the current directory
2. Start a web server on port 8080

Open your browser and go to: **http://localhost:8080** (or your server's IP/domain).

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

1. Open Kibble in your browser — you'll be directed to create an admin account
2. Set a username and password (minimum 8 characters)
3. Log in with your new credentials
4. Go to the **Settings** page
5. Paste your Gemini API key and click "Test Key" to verify it works
6. Click "Save Settings"

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

## Production Deployment

### Running as a Systemd Service

Create a systemd service so Kibble starts on boot and stays running:

```bash
sudo nano /etc/systemd/system/kibble.service
```

#### For a VPS (Alma Linux / Rocky Linux / Ubuntu):

```ini
[Unit]
Description=Kibble Facts Dashboard
After=network.target

[Service]
Type=simple
User=root
Group=root
WorkingDirectory=/var/lib/kibble
ExecStart=/usr/local/bin/kibble
Restart=on-failure
RestartSec=5s

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal
SyslogIdentifier=kibble

[Install]
WantedBy=multi-user.target
```

#### For a Raspberry Pi:

```ini
[Unit]
Description=Kibble Facts Dashboard
After=network.target

[Service]
Type=simple
User=pi
Group=pi
WorkingDirectory=/home/pi/kibble
ExecStart=/home/pi/kibble/kibble-linux-arm64
Restart=on-failure
RestartSec=5s

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal
SyslogIdentifier=kibble

[Install]
WantedBy=multi-user.target
```

Then enable and start it:

```bash
sudo systemctl daemon-reload
sudo systemctl enable kibble
sudo systemctl start kibble
```

Check status with: `sudo systemctl status kibble`

View logs with: `sudo journalctl -u kibble -f`

### SELinux (RHEL / Alma Linux / Rocky Linux)

If you're running a RHEL-based distribution with SELinux enforcing, set the correct file context:

```bash
sudo semanage fcontext -a -t var_lib_t "/var/lib/kibble(/.*)?"
sudo restorecon -Rv /var/lib/kibble
```

### Reverse Proxy with Caddy (for HTTPS)

If you're exposing Kibble to the internet (e.g., via Cloudflare Tunnel), use Caddy as a reverse proxy:

```bash
sudo dnf install caddy   # or apt install caddy
```

Create `/etc/caddy/Caddyfile`:

```
:80 {
    reverse_proxy localhost:8080 {
        header_up X-Forwarded-Proto https
    }
}
```

```bash
sudo systemctl enable --now caddy
```

Kibble automatically detects the `X-Forwarded-Proto` header and sets secure cookies when behind HTTPS.

### Directory Structure (Recommended)

```
/usr/local/bin/kibble          # Binary
/var/lib/kibble/               # Working directory
/var/lib/kibble/kibble.db      # SQLite database (auto-created)
/var/lib/kibble/config.yaml    # Optional config file
/etc/systemd/system/kibble.service  # Systemd service
```

## Updating Kibble

### Browser-Based Update (Recommended)

Kibble can update itself from the Settings page:

1. Go to **Settings** and scroll to the **Update Kibble** card
2. Click **Check for Updates** to see if a new version is available
3. If an update is available, click **Install Update**
4. Kibble will download the correct binary for your platform, replace itself, and restart automatically
5. The page will reload with the new version

Your database, settings, topics, facts, and password are never affected by updates.

> **Note:** The self-update feature requires that the Kibble process has write permission to its own binary location. If running as a systemd service with `User=root`, this works automatically. If running as a non-root user, ensure the user has write access to the binary directory.

### Manual Update

```bash
# Download the new binary
wget https://github.com/thinkscotty/kibble/releases/latest/download/kibble-linux-amd64

# Stop the service, replace the binary, and restart
sudo systemctl stop kibble
sudo cp kibble-linux-amd64 /usr/local/bin/kibble
sudo chmod +x /usr/local/bin/kibble
sudo systemctl start kibble
```

## Troubleshooting

- **"API key not configured"**: Go to Settings and enter your Gemini API key
- **Facts not generating**: Click "Test Key" on the Settings page to verify your API key works
- **Page not loading**: Make sure nothing else is using port 8080, or change the port in `config.yaml`
- **Can't access from another device**: Make sure you're using the Pi's IP address (not `localhost`) and that both devices are on the same network

## License

MIT
