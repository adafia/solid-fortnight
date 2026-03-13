#!/bin/bash

# Function to unset environment variables defined in a .env file
unset_env() {
    local env_file=".env"
    
    if [ -f "$env_file" ]; then
        # Extract keys from .env file (ignoring comments and empty lines)
        # 1. grep -v '^#' removes comments
        # 2. grep '=' ensures we only get lines with assignments
        # 3. cut -d '=' -f 1 gets the variable name before the equals sign
        # 4. xargs trims any whitespace
        local keys=$(grep -v '^#' "$env_file" | grep '=' | cut -d '=' -f 1 | xargs)
        
        for key in $keys; do
            unset "$key"
        done
        echo "Environment variables from $env_file have been unset."
    else
        echo "Warning: $env_file not found. Nothing to unset."
    fi
}

unset_env
