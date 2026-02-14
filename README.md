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

3. Install the binary and prepare directories:

```bash
# Download (replace URL with the latest release and your platform)
wget https://github.com/thinkscotty/kibble/releases/latest/download/kibble-linux-amd64

# Install to a standard location
sudo cp kibble-linux-amd64 /usr/local/bin/kibble
sudo chmod +x /usr/local/bin/kibble

# Create the data directory (Kibble stores its database here)
sudo mkdir -p /var/lib/kibble
```

> **Important:** Kibble creates its SQLite database (`kibble.db`) in the **current working directory** by default. You can also set `database.path` in `config.yaml` to either a file path (e.g., `/var/lib/kibble/kibble.db`) or a directory (e.g., `/var/lib/kibble`) — if you point to a directory, Kibble will automatically use `kibble.db` inside it. See [Configuration](#configuration-optional) below.

4. Run it:

```bash
cd /var/lib/kibble
kibble
```

If you're running as a non-root user, make sure you own the data directory:

```bash
sudo chown $USER:$USER /var/lib/kibble
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

Run the binary from the data directory where you want the database stored:

```bash
cd /var/lib/kibble
kibble
```

Kibble will:
1. Create a `kibble.db` SQLite database in the current working directory
2. Start a web server on port 8080
3. Look for an optional `config.yaml` in the current directory

Open your browser and go to: **http://localhost:8080** (or your server's IP/domain).

> **Note:** The current working directory matters. Kibble writes `kibble.db` relative to wherever you run it from. If you start Kibble from `/root` instead of `/var/lib/kibble`, the database will be created at `/root/kibble.db`. When using systemd, the `WorkingDirectory=` setting controls this (see [Production Deployment](#running-as-a-systemd-service)).

### Configuration (Optional)

You can create a `config.yaml` file in the data directory (`/var/lib/kibble/`) to customize settings:

```yaml
server:
  host: "0.0.0.0"     # Listen on all interfaces (default)
  port: 8080           # Web server port

database:
  path: "/var/lib/kibble"  # Directory or file path — if a directory, Kibble uses kibble.db inside it

logging:
  level: "info"        # debug, info, warn, error

similarity:
  threshold: 0.6       # How similar facts must be to be considered duplicates (0.0-1.0)
  ngram_size: 3        # Trigram size for similarity comparison
```

All of these have sensible defaults, so the config file is entirely optional. You can set `database.path` to either a directory (`/var/lib/kibble`) or a full file path (`/var/lib/kibble/kibble.db`) — both work. Kibble will also create any missing parent directories automatically.

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

Kibble provides a JSON API for external devices like LED matrix displays, smart screens, and custom clients.

### Authentication

All API endpoints require your Gemini API key. You can provide it in two ways:

**Query parameter** (simplest for devices):
```
GET /api/v1/facts/random?api_key=YOUR_API_KEY
```

**Authorization header**:
```
Authorization: Bearer YOUR_API_KEY
```

The API key is the same Gemini API key you entered on the Settings page.

### Endpoints

#### Get Active Topics
```
GET /api/v1/topics
```
Returns all active topics with their fact counts.

**Response:**
```json
{
  "topics": [
    { "id": 1, "name": "Space", "fact_count": 25 },
    { "id": 2, "name": "Marine Biology", "fact_count": 18 }
  ]
}
```

#### Get Facts for a Topic
```
GET /api/v1/facts?topic_id=1&limit=5
```
Returns facts for a specific topic. The `limit` parameter is optional (default: 10).

**Response:**
```json
{
  "topic": "Space",
  "facts": [
    { "id": 42, "content": "The Voyager 1 spacecraft..." },
    { "id": 41, "content": "A neutron star can spin..." }
  ]
}
```

#### Get All Facts
```
GET /api/v1/facts/all
```
Returns every non-archived fact from all active topics, grouped by topic. Use this to sync a client device with the complete fact library.

**Response:**
```json
{
  "topics": [
    {
      "topic_id": 1,
      "topic_name": "Space",
      "facts": [
        { "id": 42, "content": "The Voyager 1 spacecraft..." },
        { "id": 41, "content": "A neutron star can spin..." }
      ]
    },
    {
      "topic_id": 2,
      "topic_name": "Marine Biology",
      "facts": [ ... ]
    }
  ]
}
```

#### Get Recent Facts
```
GET /api/v1/facts/recent
```
Returns the 100 most recently created facts from each active topic, grouped by topic. Useful for periodic syncs where you only need the latest content.

**Response:** Same structure as `/api/v1/facts/all`, but capped at 100 facts per topic.

#### Get a Random Fact
```
GET /api/v1/facts/random
```
Returns a single random fact from any active topic — perfect for scrolling tickers and displays.

**Response:**
```json
{
  "fact": {
    "id": 42,
    "topic": "Space",
    "content": "The Voyager 1 spacecraft..."
  }
}
```

### Example: Client Device Sync

To populate a client device with all current facts in a single request:
```bash
curl "https://your-domain.com/api/v1/facts/all?api_key=YOUR_API_KEY"
```

To periodically refresh with the latest facts:
```bash
curl "https://your-domain.com/api/v1/facts/recent?api_key=YOUR_API_KEY"
```

To fetch one random fact at a time (e.g., for an LED ticker):
```bash
curl "https://your-domain.com/api/v1/facts/random?api_key=YOUR_API_KEY"
```

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
Restart=always
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
Restart=always
RestartSec=5s

NoNewPrivileges=true
PrivateTmp=true

StandardOutput=journal
StandardError=journal
SyslogIdentifier=kibble

[Install]
WantedBy=multi-user.target
```

> **Critical:** The `WorkingDirectory=` must point to a directory that **exists** and is **writable** by the `User=` specified above. This is where Kibble creates its database. If the directory is missing or the user doesn't have write permission, Kibble will fail with "unable to open database file". Create the directory first:
> ```bash
> # For VPS (running as root):
> sudo mkdir -p /var/lib/kibble
>
> # For Raspberry Pi (running as pi):
> mkdir -p /home/pi/kibble
> ```

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

- **"unable to open database file"** or **"out of memory (14)"**: This is SQLite error code 14 (`SQLITE_CANTOPEN`) — it means Kibble can't create or open the database file. Check that:
  1. The running user has write permission to the data directory: `sudo chown $USER:$USER /var/lib/kibble`
  2. If using systemd, `WorkingDirectory=` in the service file points to a writable directory
  3. Set `database.path` in `config.yaml` to point to your data directory or file: `database: { path: "/var/lib/kibble" }`
- **"API key not configured"**: Go to Settings and enter your Gemini API key
- **Facts not generating**: Click "Test Key" on the Settings page to verify your API key works
- **Page not loading**: Make sure nothing else is using port 8080, or change the port in `config.yaml`
- **Can't access from another device**: Make sure you're using the server's IP address (not `localhost`) and that both devices are on the same network

## Uninstalling Kibble

To completely remove Kibble from your system, use the provided uninstall script:

```bash
# Download the uninstall script
wget https://raw.githubusercontent.com/thinkscotty/kibble/main/uninstall.sh

# Make it executable
chmod +x uninstall.sh

# Run with sudo
sudo ./uninstall.sh
```

The script will:
1. Stop and remove the systemd service
2. Remove Kibble binaries from common locations
3. Ask if you want to remove data directories (database, config, themes)
4. Search for any remaining Kibble files

**Note:** If you choose to keep your data directory when uninstalling, you can reinstall Kibble later and your topics, facts, and settings will be preserved.

### Manual Uninstall

If you prefer to uninstall manually:

```bash
# Stop and disable the service
sudo systemctl stop kibble
sudo systemctl disable kibble

# Remove the service file
sudo rm /etc/systemd/system/kibble.service
sudo systemctl daemon-reload

# Remove the binary (VPS)
sudo rm /usr/local/bin/kibble

# Or remove the binary (Raspberry Pi)
rm ~/kibble/kibble-linux-arm64

# Remove data directory (optional - contains your topics and facts)
sudo rm -rf /var/lib/kibble  # VPS
# or
rm -rf ~/kibble  # Raspberry Pi
```

## License

MIT
