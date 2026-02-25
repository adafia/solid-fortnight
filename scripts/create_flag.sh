#!/bin/bash
set -eu

if [ -z "$1" ]; then
  echo "Usage: $0 <project_id>"
  exit 1
fi

PROJECT_ID=$1

curl -X POST -H "Content-Type: application/json" -d '{"project_id": "'"$PROJECT_ID"'", "key": "my-new-flag", "name": "My New Feature Flag", "description": "This is a test flag.", "type": "boolean", "created_by": "a-user-uuid"}' http://localhost:8083/flags/
