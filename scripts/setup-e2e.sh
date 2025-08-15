#!/bin/bash

# Script to set up e2e test repositories from source directories
# This script creates bare Git repositories from the source files for testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
E2E_INPUT_DIR="$PROJECT_ROOT/e2e/_input"
TEST_REPOS_DIR="$E2E_INPUT_DIR/test-repos"

echo "Setting up e2e test repositories..."

# Clean up existing test repos
rm -rf "$TEST_REPOS_DIR"
mkdir -p "$TEST_REPOS_DIR"

# Function to create a bare repository from source directory
create_test_repo() {
    local repo_name="$1"
    local source_dir="$2"
    local branch_name="$3"
    
    echo "Creating $repo_name from $source_dir..."
    
    # Create bare repository
    git init --bare "$TEST_REPOS_DIR/$repo_name.git"
    
    # Create a temporary working directory
    local temp_dir=$(mktemp -d)
    cd "$temp_dir"
    
    # Initialize working repository
    git init
    git config user.name "E2E Test Setup"
    git config user.email "e2e@test.local"
    
    # Copy source files
    cp -r "$source_dir"/* .
    
    # Add and commit all files
    git add .
    git commit -m "Initial commit for e2e testing"
    
    # Push to bare repository
    git remote add origin "$TEST_REPOS_DIR/$repo_name.git"
    git push origin main:$branch_name
    
    # Clean up temporary directory
    cd "$PROJECT_ROOT"
    rm -rf "$temp_dir"
    
    echo "Created $repo_name.git with branch $branch_name"
}

# Create repo-1 with 'init' branch
if [ -d "$E2E_INPUT_DIR/repo-1-source" ]; then
    create_test_repo "repo-1" "$E2E_INPUT_DIR/repo-1-source" "init"
else
    echo "Error: $E2E_INPUT_DIR/repo-1-source does not exist"
    exit 1
fi

# Create repo-2 with 'main' branch  
if [ -d "$E2E_INPUT_DIR/repo-2-source" ]; then
    create_test_repo "repo-2" "$E2E_INPUT_DIR/repo-2-source" "main"
else
    echo "Error: $E2E_INPUT_DIR/repo-2-source does not exist"
    exit 1
fi

echo "E2E test repositories setup complete!"
echo "Created repositories:"
echo "  - $TEST_REPOS_DIR/repo-1.git (branch: init)"
echo "  - $TEST_REPOS_DIR/repo-2.git (branch: main)"