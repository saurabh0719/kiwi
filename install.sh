#!/bin/bash

set -e

# Set colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}Installing Kiwi CLI...${NC}"

# Determine OS and architecture
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

if [ "$ARCH" == "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" == "aarch64" ] || [ "$ARCH" == "arm64" ]; then
  ARCH="arm64"
else
  echo "Unsupported architecture: $ARCH"
  exit 1
fi

# Get the latest release URL from GitHub API
echo -e "${BLUE}Fetching latest release...${NC}"
API_URL="https://api.github.com/repos/saurabh0719/kiwi/releases/latest"
ASSET_PATTERN="kiwi-${OS}-${ARCH}.tar.gz"

# Use GitHub API to find the download URL
DOWNLOAD_URL=$(curl -s $API_URL | 
  grep "browser_download_url.*$ASSET_PATTERN" | 
  cut -d : -f 2,3 | 
  tr -d \")

if [ -z "$DOWNLOAD_URL" ]; then
  echo "Failed to find a release for your platform: ${OS}-${ARCH}"
  exit 1
fi

# Create temporary directory
TMP_DIR=$(mktemp -d)
cd $TMP_DIR

# Download and extract
echo -e "${BLUE}Downloading Kiwi binary...${NC}"
curl -sL "$DOWNLOAD_URL" -o kiwi.tar.gz
tar -xzf kiwi.tar.gz

# Create install directory if it doesn't exist
INSTALL_DIR="/usr/local/bin"
if [ ! -d "$INSTALL_DIR" ]; then
  sudo mkdir -p "$INSTALL_DIR"
fi

echo -e "${BLUE}Moving kiwi to $INSTALL_DIR${NC}"
sudo mv kiwi "$INSTALL_DIR"
sudo chmod +x "$INSTALL_DIR/kiwi"

# Clean up
cd - > /dev/null
rm -rf $TMP_DIR

# Verify installation
if command -v kiwi >/dev/null 2>&1; then
  echo -e "${GREEN}Kiwi installed successfully! Try running: kiwi --help${NC}"
else
  echo "Installation failed. Please check errors above."
  exit 1
fi

# Instructions for configuration
echo -e "${BLUE}Next steps:${NC}"
echo -e "1. Set your API key: ${GREEN}kiwi config set llm.api_key your-api-key-here${NC}"
echo -e "2. Set your preferred model: ${GREEN}kiwi config set llm.model gpt-4o${NC}"
echo -e "3. Try a sample query: ${GREEN}kiwi \"Explain what Kiwi can do\"${NC}" 