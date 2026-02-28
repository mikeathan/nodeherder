#!/usr/bin/env bash

set -e

# extract the version from frontend/package.json
VERSION=$(node -p "require('./frontend/package.json').version")

if [ -z "$VERSION" ]; then
    echo "Error: Could not extract version from frontend/package.json"
    exit 1
fi

echo "Deploying version: ${VERSION}"

echo "Creating tag..."
git tag -a "v${VERSION}" -m "Version ${VERSION}"

echo "Pushing tag..."
git push origin "v${VERSION}"
