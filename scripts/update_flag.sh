#!/bin/bash
set -eu

if [ -z "$1" ] || [ -z "$2" ]; then
  echo "Usage: $0 <project_id> <flag_id>"
  exit 1
fi

PROJECT_ID=$1
FLAG_ID=$2

curl -X PUT -H "Content-Type: application/json" -d '{"project_id": "'"$PROJECT_ID"'", "key": "my-new-flag", "name": "Updated Feature Flag Name", "description": "This flag has been updated.", "type": "string", "created_by": "a-user-uuid"}' http://localhost:8083/flags/$FLAG_ID
