#!/bin/bash
set -eu

if [ -z "$1" ]; then
  echo "Usage: $0 <flag_id>"
  exit 1
fi

FLAG_ID=$1

curl -X DELETE http://localhost:8083/flags/$FLAG_ID
