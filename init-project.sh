#!/usr/bin/env bash
set -e
OLD=$(go list -m)
NEW=$1
go mod edit -module "$NEW"
find . -name '*.go' -exec sed -i "s|$OLD|$NEW|g" {} +
echo "Done: $OLD → $NEW"