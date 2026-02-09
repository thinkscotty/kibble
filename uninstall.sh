#!/bin/bash
#
# Kibble Uninstall Script
#
# This script removes Kibble binaries, systemd service, and optionally user data.
# Run with sudo: sudo ./uninstall.sh
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo -e "${RED}Error: This script must be run as root (use sudo)${NC}"
    exit 1
fi

echo -e "${GREEN}Kibble Uninstall Script${NC}"
echo "========================================"
echo ""

# Detect common installation locations
BINARY_LOCATIONS=(
    "/usr/local/bin/kibble"
    "/home/pi/kibble/kibble-linux-arm64"
    "/home/pi/kibble/kibble-linux-arm"
    "$HOME/kibble/kibble"
)

DATA_LOCATIONS=(
    "/var/lib/kibble"
    "/home/pi/kibble"
)

SERVICE_FILE="/etc/systemd/system/kibble.service"

# Function to remove systemd service
remove_service() {
    if [ -f "$SERVICE_FILE" ]; then
        echo -e "${YELLOW}Stopping and removing systemd service...${NC}"

        # Stop the service if it's running
        if systemctl is-active --quiet kibble; then
            echo "  Stopping kibble service..."
            systemctl stop kibble
        fi

        # Disable the service if it's enabled
        if systemctl is-enabled --quiet kibble 2>/dev/null; then
            echo "  Disabling kibble service..."
            systemctl disable kibble
        fi

        # Remove the service file
        echo "  Removing service file: $SERVICE_FILE"
        rm -f "$SERVICE_FILE"

        # Reload systemd
        echo "  Reloading systemd daemon..."
        systemctl daemon-reload
        systemctl reset-failed 2>/dev/null || true

        echo -e "${GREEN}✓ Service removed successfully${NC}"
    else
        echo "  No systemd service found (not installed as service)"
    fi
    echo ""
}

# Function to remove binaries
remove_binaries() {
    echo -e "${YELLOW}Removing Kibble binaries...${NC}"

    local found=false
    for binary in "${BINARY_LOCATIONS[@]}"; do
        if [ -f "$binary" ]; then
            echo "  Removing: $binary"
            rm -f "$binary"
            found=true
        fi
    done

    if [ "$found" = true ]; then
        echo -e "${GREEN}✓ Binaries removed successfully${NC}"
    else
        echo "  No binaries found in common locations"
    fi
    echo ""
}

# Function to remove data directories
remove_data() {
    echo -e "${YELLOW}Checking for data directories...${NC}"

    local found_dirs=()
    for dir in "${DATA_LOCATIONS[@]}"; do
        if [ -d "$dir" ] && [ -f "$dir/kibble.db" -o -f "$dir/config.yaml" -o -f "$dir/themes.yaml" ]; then
            found_dirs+=("$dir")
        fi
    done

    if [ ${#found_dirs[@]} -eq 0 ]; then
        echo "  No data directories found"
        echo ""
        return
    fi

    echo "  Found data directories:"
    for dir in "${found_dirs[@]}"; do
        echo "    - $dir"
        if [ -f "$dir/kibble.db" ]; then
            local db_size=$(du -h "$dir/kibble.db" 2>/dev/null | cut -f1)
            echo "      (contains database: $db_size)"
        fi
    done
    echo ""

    echo -e "${YELLOW}WARNING: This will delete all Kibble data including:${NC}"
    echo "  - Database (topics, facts, settings, usage logs)"
    echo "  - Configuration files"
    echo "  - Theme files"
    echo ""

    read -p "Do you want to remove data directories? [y/N] " -n 1 -r
    echo ""

    if [[ $REPLY =~ ^[Yy]$ ]]; then
        for dir in "${found_dirs[@]}"; do
            echo "  Removing: $dir"
            rm -rf "$dir"
        done
        echo -e "${GREEN}✓ Data directories removed${NC}"
    else
        echo "  Keeping data directories (you can manually remove them later)"
    fi
    echo ""
}

# Function to search for any remaining Kibble files
search_remaining() {
    echo -e "${YELLOW}Searching for remaining Kibble files...${NC}"

    local remaining=()

    # Search common locations for kibble executables
    while IFS= read -r -d '' file; do
        # Skip files in bin/ or go build cache
        if [[ ! "$file" =~ "/bin/" ]] && [[ ! "$file" =~ "/.cache/" ]] && [[ ! "$file" =~ "/go/" ]]; then
            remaining+=("$file")
        fi
    done < <(find /usr /opt /home 2>/dev/null -name "kibble*" -type f -executable -print0 2>/dev/null || true)

    if [ ${#remaining[@]} -gt 0 ]; then
        echo -e "${YELLOW}  Found additional files:${NC}"
        for file in "${remaining[@]}"; do
            echo "    - $file"
        done
        echo ""
        echo "  You may want to remove these manually if they're related to Kibble."
    else
        echo "  No additional Kibble files found"
    fi
    echo ""
}

# Main uninstall process
main() {
    echo "This script will remove Kibble from your system."
    echo ""
    read -p "Continue with uninstall? [y/N] " -n 1 -r
    echo ""

    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Uninstall cancelled."
        exit 0
    fi

    echo ""

    # Remove systemd service
    remove_service

    # Remove binaries
    remove_binaries

    # Ask about data removal
    remove_data

    # Search for any remaining files
    search_remaining

    echo -e "${GREEN}========================================"
    echo -e "Kibble uninstall complete!${NC}"
    echo ""
    echo "If you reinstall Kibble later and kept your data directory,"
    echo "your previous topics, facts, and settings will be preserved."
    echo ""
}

# Run main function
main
