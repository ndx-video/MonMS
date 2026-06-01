#!/usr/bin/env bash
# Rebuild offline Tailwind v4 CSS for the MonMS dashboard.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
STATIC="$ROOT/internal/monmsdash/ui/static"
cd "$STATIC"
if [[ ! -d node_modules ]]; then
  npm install tailwindcss@4 @tailwindcss/cli@4
fi
npx @tailwindcss/cli -i src/input.css -o monms-dash.css --minify
echo "Wrote $STATIC/monms-dash.css"
