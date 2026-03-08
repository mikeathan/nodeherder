#!/usr/bin/env bash

set -e

AUTO_INCREMENT=false

for arg in "$@"; do
    if [ "$arg" == "--auto-increment-patch" ]; then
        AUTO_INCREMENT=true
        break
    fi
done

if [ "$AUTO_INCREMENT" = true ]; then
    echo "Determining next patch version from git tags..."
    # Get the latest tag matching v* pattern, sort by versions to get the highest one
    LATEST_TAG=$(git tag -l "v*" | sort -V | tail -n 1)

    if [ -z "$LATEST_TAG" ]; then
        # Fallback if no tags exist
        VERSION="0.0.1"
    else
        # Remove 'v' prefix
        CLEAN_TAG="${LATEST_TAG#v}"

        # Split into major, minor, patch
        IFS='.' read -r -a VERSION_PARTS <<< "$CLEAN_TAG"
        MAJOR="${VERSION_PARTS[0]}"
        MINOR="${VERSION_PARTS[1]}"
        PATCH="${VERSION_PARTS[2]}"

        if [ -z "$PATCH" ]; then
            PATCH=0
        fi

        # Increment patch
        PATCH=$((PATCH + 1))

        VERSION="${MAJOR}.${MINOR}.${PATCH}"
    fi
else
    # extract the version from frontend/package.json
    VERSION=$(node -p "require('./frontend/package.json').version")

    if [ -z "$VERSION" ]; then
        echo "Error: Could not extract version from frontend/package.json"
        exit 1
    fi
fi

echo "Deploying version: ${VERSION}"

echo "Creating tag..."
git tag -a "v${VERSION}" -m "Version ${VERSION}"

echo "Pushing tag..."
git push origin "v${VERSION}"
