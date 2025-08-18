#!/usr/bin/env bash

set -e

# Build and run tests
echo "Building and running tests..."
make build
make test
make e2e

# Make sure we are on the main branch
CURRENT_BRANCH=`git rev-parse --abbrev-ref HEAD`
if [ "$CURRENT_BRANCH" != "main" ]; then
  echo "Must be on main branch to release."
  exit 1
fi

# Calculate tag YYYY-MM-DD-HASH
DATE=`git log -n1 --pretty='%cd' --date=format:'%Y-%m-%d'`
HASH=`git rev-parse --short HEAD`
TAG="$DATE-$HASH"
IMAGE_NAME="slash-prompt"

# Build Docker image
echo "Building Docker image..."
docker build --build-arg TAG="$TAG" -t "$IMAGE_NAME:$TAG" .
docker tag "$IMAGE_NAME:$TAG" "$IMAGE_NAME:latest"

# Tag images for ghcr.io
GHCR_REPO="ghcr.io/appgardenstudios/slash-prompt"
docker tag "$IMAGE_NAME:$TAG" "$GHCR_REPO:$TAG"
docker tag "$IMAGE_NAME:$TAG" "$GHCR_REPO:latest"

# Save Docker image as artifact
echo "Saving Docker image as artifact..."
mkdir -p ./dist/
docker save "$IMAGE_NAME:$TAG" | gzip > "./dist/$IMAGE_NAME-$TAG.tar.gz"

# Create Tag
echo "Creating Git tag $TAG..."
git tag $TAG

# Push tag to GitHub
git push origin $TAG

# Push Docker images to ghcr.io
echo "Pushing Docker images to ghcr.io..."
docker push "$GHCR_REPO:$TAG"
docker push "$GHCR_REPO:latest"

# Create Draft Release
gh release create $TAG --draft --verify-tag --generate-notes --latest ./dist/$IMAGE_NAME-$TAG.tar.gz