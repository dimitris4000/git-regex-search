#!/bin/bash

# Git Regex Search Release Script (low-tech version 🙂)
# Usage: ./release.sh <version>
# Example: ./release.sh v1.0.0

set -e  # Exit on any error

# Check if version parameter is provided
if [ $# -eq 0 ]; then
    echo "❌ Error: Version parameter is required"
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

VERSION=$1
BINARY_NAME="git-regex-search"
BUILD_DIR="build"

# Validate version format
if [[ ! $VERSION =~ ^[0-9]+\.[0-9]+\.[0-9]+.*$ ]]; then
    echo "❌ Error: Version should follow semantic versioning format (e.g., 1.0.0)"
    exit 1
fi

echo "🚀 Starting release process for version: $VERSION"
echo ""

# Check if gh CLI is installed and authenticated
if ! command -v gh &> /dev/null; then
    echo "❌ Error: GitHub CLI (gh) is not installed"
    echo "Install it from: https://cli.github.com/"
    exit 1
fi

# Check if authenticated with GitHub
if ! gh auth status &> /dev/null; then
    echo "❌ Error: Not authenticated with GitHub CLI"
    echo "Run: gh auth login"
    exit 1
fi

# Check if we're in a git repository
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    echo "❌ Error: Not in a git repository"
    exit 1
fi

# Check if there are uncommitted changes
if ! git diff-index --quiet HEAD --; then
    echo "⚠️  Warning: There are uncommitted changes"
    read -p "Do you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "❌ Release cancelled"
        exit 1
    fi
fi

# Create build directory
echo "📁 Creating build directory..."
rm -rf $BUILD_DIR
mkdir -p $BUILD_DIR

# Build for multiple platforms
echo "🔨 Building binaries for multiple platforms..."

platforms=(
    "linux/amd64"
    "linux/arm64"
    "darwin/amd64"
    "darwin/arm64"
    "windows/amd64"
)

for platform in "${platforms[@]}"; do
    IFS='/' read -r -a array <<< "$platform"
    GOOS="${array[0]}"
    GOARCH="${array[1]}"
    
    output_name="$BUILD_DIR/${BINARY_NAME}-${GOOS}-${GOARCH}"
    if [ $GOOS = "windows" ]; then
        output_name+='.exe'
    fi
    
    echo "  📦 Building static binary for $GOOS/$GOARCH..."
    if [ $GOOS = "linux" ]; then
        # Build static binary for Linux
        GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags="-s -w -extldflags '-static'" -o $output_name
    else
        # Build regular binary for other platforms
        GOOS=$GOOS GOARCH=$GOARCH CGO_ENABLED=0 go build -ldflags="-s -w" -o $output_name
    fi
    
    if [ $? -ne 0 ]; then
        echo "❌ Error: Failed to build for $GOOS/$GOARCH"
        exit 1
    fi
done

echo "✅ All binaries built successfully!"
echo ""

# List built files
echo "📋 Built files:"
ls -la $BUILD_DIR/
echo ""

# Check if tag already exists
if git tag -l | grep -q "^$VERSION$"; then
    echo "⚠️  Tag $VERSION already exists"
    read -p "Do you want to delete and recreate it? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🗑️  Deleting existing tag..."
        git tag -d $VERSION
        git push origin :refs/tags/$VERSION 2>/dev/null || true
    else
        echo "❌ Release cancelled"
        exit 1
    fi
fi

# Create git tag
echo "🏷️  Creating git tag: $VERSION"
git tag -a $VERSION -m "Release $VERSION"
git push origin $VERSION

# Generate release notes
RELEASE_NOTES="## 🎉 Release $VERSION

### Features
- 🔍 **Cross-branch regex search**: Search for patterns across all remote branches or specific branches
- 🚀 **Smart tool detection**: Uses \`ripgrep\` (rg) for faster searches when available, falls back to \`grep\`
- 💾 **Safe operations**: Automatically stashes uncommitted changes and restores them after search
- 🎨 **Colorized output**: Branch names and line numbers are highlighted for better readability
- 📊 **Verbose progress**: Real-time updates showing which branch is being searched and match counts
- 🌐 **Remote branch support**: Automatically fetches and pulls latest changes from all remote branches
- ⚡ **Selective branch search**: Option to search only specific branches

### Installation

Download the appropriate binary for your platform:

- **Linux (x64)**: \`git-regex-search-linux-amd64\`
- **Linux (ARM64)**: \`git-regex-search-linux-arm64\`
- **macOS (Intel)**: \`git-regex-search-darwin-amd64\`
- **macOS (Apple Silicon)**: \`git-regex-search-darwin-arm64\`
- **Windows (x64)**: \`git-regex-search-windows-amd64.exe\`

Make the binary executable and run:
\`\`\`bash
chmod +x git-regex-search-*
./git-regex-search-* --help
\`\`\`

### Usage
\`\`\`bash
# Search all branches
./git-regex-search --repo /path/to/repo --regex \"your-pattern\"

# Search specific branches
./git-regex-search --repo /path/to/repo --regex \"your-pattern\" --branches \"main,develop\"
\`\`\`

### Requirements
- Git
- \`ripgrep\` (optional, for faster searches)

---
Built with ❤️ using Go"

# Create GitHub release
echo "🚀 Creating GitHub release..."
gh release create $VERSION \
    $BUILD_DIR/* \
    --title "git-regex-search $VERSION" \
    --generate-notes \
    --latest
    # --notes "$RELEASE_NOTES" \

if [ $? -eq 0 ]; then
    echo ""
    echo "🎉 Release $VERSION created successfully!"
    echo "📦 Binaries uploaded to GitHub Releases"
    echo "🔗 View release: $(gh repo view --web)/releases/tag/$VERSION"
    echo ""
    echo "🧹 Cleaning up build directory..."
    rm -rf $BUILD_DIR
    echo "✅ Done!"
else
    echo "❌ Error: Failed to create GitHub release"
    exit 1
fi
