#!/usr/bin/env bash

set -e

AUTO_INCREMENT=false

for arg in "$@"; do
    if [ "$arg" == "--auto-increment-patch" ]; then
        AUTO_INCREMENT=true
        break
    fi
done

echo "Reading base version from frontend/package.json..."
PKG_VERSION=$(node -p "require('./frontend/package.json').version")

if [ -z "$PKG_VERSION" ]; then
    echo "Error: Could not extract version from frontend/package.json"
    exit 1
fi

# Split PKG_VERSION into major, minor, patch
IFS='.' read -r -a PKG_VERSION_PARTS <<< "$PKG_VERSION"
MAJOR="${PKG_VERSION_PARTS[0]}"
MINOR="${PKG_VERSION_PARTS[1]}"
BASE_PATCH="${PKG_VERSION_PARTS[2]}"

if [ -z "$BASE_PATCH" ]; then
    BASE_PATCH=0
fi

if [ "$AUTO_INCREMENT" = true ]; then
    echo "Determining next patch version from git tags matching v${MAJOR}.${MINOR}.*..."
    # Get the latest tag matching the current major and minor
    LATEST_TAG=$(git tag -l "v${MAJOR}.${MINOR}.*" | sort -V | tail -n 1)

    if [ -z "$LATEST_TAG" ]; then
        # No existing tags for this major/minor version, start at the patch defined in package.json (usually 0)
        PATCH="$BASE_PATCH"
        echo "No existing tags found for v${MAJOR}.${MINOR}.*. Starting at patch ${PATCH}."
    else
        # Remove 'v' prefix
        CLEAN_TAG="${LATEST_TAG#v}"

        # Split into major, minor, patch
        IFS='.' read -r -a VERSION_PARTS <<< "$CLEAN_TAG"
        LATEST_PATCH="${VERSION_PARTS[2]}"

        if [ -z "$LATEST_PATCH" ]; then
            LATEST_PATCH=0
        fi

        # Increment patch
        PATCH=$((LATEST_PATCH + 1))
        echo "Highest existing tag is ${LATEST_TAG}. Incrementing patch to ${PATCH}."
    fi

    VERSION="${MAJOR}.${MINOR}.${PATCH}"
else
    # Do not auto-increment; use exact version from package.json
    VERSION="$PKG_VERSION"
fi

echo "Deploying version: ${VERSION}"

# Prevent Git failure if tag already exists for some reason
if git show-ref --tags "v${VERSION}" --quiet; then
    echo "Error: Tag v${VERSION} already exists. Please manually increment the version in package.json."
    exit 1
fi

echo "Creating tag..."
git tag -a "v${VERSION}" -m "Version ${VERSION}"

echo "Pushing tag..."
git push origin "v${VERSION}"
