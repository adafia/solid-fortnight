#!/bin/bash

# Function to load environment variables from a .env file
load_env() {
    local env_file=".env"

    if [ -f "$env_file" ]; then
        # Load the .env file
        # Using export to make them available in the current shell (if sourced)
        # Using grep to skip comments and empty lines
        set -a
        source "$env_file"
        set +a
        echo "Environment variables loaded from $env_file"
    else
        echo "Error: $env_file not found."
        return 1
    fi
}

load_env
