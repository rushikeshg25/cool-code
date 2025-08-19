#!/bin/bash

# Release script for cool-code
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting release process for cool-code${NC}"

# Check if we're on main branch
CURRENT_BRANCH=$(git branch --show-current)
if [ "$CURRENT_BRANCH" != "main" ]; then
    echo -e "${RED}❌ Please switch to main branch before releasing${NC}"
    exit 1
fi

# Check if working directory is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${RED}❌ Working directory is not clean. Please commit or stash changes.${NC}"
    exit 1
fi

# Pull latest changes
echo -e "${YELLOW}📥 Pulling latest changes...${NC}"
git pull origin main

# Install dependencies
echo -e "${YELLOW}📦 Installing dependencies...${NC}"
npm ci

# Build the project
echo -e "${YELLOW}🔨 Building project...${NC}"
npm run build

# Check if dist/index.js exists
if [ ! -f "dist/index.js" ]; then
    echo -e "${RED}❌ Build failed: dist/index.js not found${NC}"
    exit 1
fi

# Get version type from user
echo -e "${YELLOW}📋 What type of release is this?${NC}"
echo "1) patch (bug fixes)"
echo "2) minor (new features)"
echo "3) major (breaking changes)"
read -p "Enter choice (1-3): " choice

case $choice in
    1) VERSION_TYPE="patch";;
    2) VERSION_TYPE="minor";;
    3) VERSION_TYPE="major";;
    *) echo -e "${RED}❌ Invalid choice${NC}"; exit 1;;
esac

# Bump version
echo -e "${YELLOW}📈 Bumping $VERSION_TYPE version...${NC}"
npm version $VERSION_TYPE

# Get the new version
NEW_VERSION=$(node -p "require('./package.json').version")
echo -e "${GREEN}✅ New version: $NEW_VERSION${NC}"

# Push changes and tags
echo -e "${YELLOW}📤 Pushing changes and tags...${NC}"
git push origin main --tags

echo -e "${GREEN}🎉 Release $NEW_VERSION initiated!${NC}"
echo -e "${GREEN}📦 GitHub Actions will automatically publish to npm when the tag is pushed.${NC}"
echo -e "${YELLOW}🔗 Check the progress at: https://github.com/rushikeshg25/cool-code/actions${NC}"