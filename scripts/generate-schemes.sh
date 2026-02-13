#!/bin/bash
# Generates internal/community/schemes.json from iTerm2-Color-Schemes repository.
# Source: https://github.com/mbadolato/iTerm2-Color-Schemes (MIT)
#
# Usage: ./scripts/generate-schemes.sh [path-to-iterm2-color-schemes]
# Default path: ~/code/iTerm2-Color-Schemes

set -euo pipefail

REPO_DIR="${1:-$HOME/code/iTerm2-Color-Schemes}"
SOURCE_DIR="$REPO_DIR/windowsterminal"
OUTPUT="$(dirname "$0")/../internal/community/schemes.json"

if [ ! -d "$SOURCE_DIR" ]; then
    echo "Error: $SOURCE_DIR not found"
    echo "Clone https://github.com/mbadolato/iTerm2-Color-Schemes to $REPO_DIR"
    exit 1
fi

python3 -c "
import json, os, sys

source = '$SOURCE_DIR'
schemes = []
for f in sorted(os.listdir(source)):
    if not f.endswith('.json'):
        continue
    with open(os.path.join(source, f)) as fh:
        schemes.append(json.load(fh))

schemes.sort(key=lambda s: s.get('name', '').lower())

with open('$OUTPUT', 'w') as out:
    json.dump(schemes, out, separators=(',', ':'))

print(f'{len(schemes)} schemes -> $OUTPUT ({os.path.getsize(\"$OUTPUT\") / 1024:.1f} KB)')
"
