#!/bin/bash

set -e

# Function to check if a command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to install Docker
install_docker() {
    echo "Installing Docker..."
    curl -fsSL https://get.docker.com -o get-docker.sh
    sudo sh get-docker.sh
    sudo usermod -aG docker "$USER"
    rm get-docker.sh
    echo "Docker installed successfully. Please log out and log back in to apply group changes."
}

# Function to get public IP address
get_public_ip() {
    PUBLIC_IP=$(curl -s https://api.ipify.org)
    if [ -z "$PUBLIC_IP" ]; then
        echo "Error: Unable to retrieve public IP address."
        exit 1
    fi
    echo "Public IP address: $PUBLIC_IP"
}

# Function to read existing .env file
read_env_file() {
    if [ -f ".env" ]; then
        source .env
    fi
}

# Function to create or update .env file
create_env_file() {
    local env_file=".env"
    echo "Checking .env file..."
    
    read_env_file

    # Check if OPENAI_API_KEY is set
    if [ -z "$OPENAI_API_KEY" ]; then
        read -p "Enter your OpenAI API Key: " OPENAI_API_KEY
    else
        echo "OPENAI_API_KEY is already set."
    fi

    # Check if TELEGRAM_BOT_TOKEN is set
    if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
        read -p "Enter your Telegram Bot Token: " TELEGRAM_BOT_TOKEN
    else
        echo "TELEGRAM_BOT_TOKEN is already set."
    fi


    # Get public IP
    get_public_ip

    # Write variables to .env file
    cat > "$env_file" << EOF
OPENAI_API_KEY=$OPENAI_API_KEY
TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN
PUBLIC_IP=$PUBLIC_IP
EOF

    echo ".env file created/updated successfully."
}

# Main script execution

# Check and install Docker if not present
if ! command_exists docker; then
    install_docker
    echo "Please restart your terminal or log out and log back in, then run this script again."
    exit 0
else
    echo "Docker is already installed."
fi

# Check if docker compose is installed
if ! command_exists docker compose; then
    echo "Error: Docker Compose is not installed or not in PATH. Please install Docker Compose and run this script again."
    exit 1
fi

# Create or update .env file
create_env_file

# Check if docker-compose.yml exists
if [ ! -f "docker-compose.yml" ]; then
    echo "Error: docker-compose.yml not found in the current directory."
    exit 1
fi

# Start the project with docker compose
echo "Starting the project with docker compose..."
sudo docker compose up --build -d

echo "Project started successfully in the background."
echo "To view logs, use: docker compose logs -f"
echo "To stop the project, use: docker compose stop"
echo "Open http://$PUBLIC_IP:3456 in your browser to view the project."
