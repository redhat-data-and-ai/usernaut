#!/bin/bash

# DCO (Developer Certificate of Origin) check script
# This script validates that commits are properly signed off

# Ensure we're running with bash
if [ -z "$BASH_VERSION" ]; then
    echo "This script requires bash. Please run with bash."
    exit 1
fi

set -e

RED='\033[0;31m'
YELLOW='\033[1;33m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

echo -e "${GREEN}📝 Checking DCO (Developer Certificate of Origin)...${NC}"

# Get the commit message file (passed as first argument to commit-msg hook)
COMMIT_MSG_FILE="$1"

if [ -z "$COMMIT_MSG_FILE" ] || [ ! -f "$COMMIT_MSG_FILE" ]; then
    echo -e "${RED}❌ DCO check failed - no commit message file found${NC}"
    exit 1
fi

# Read the commit message
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Check if commit message contains Signed-off-by line
if echo "$COMMIT_MSG" | grep -qE "^Signed-off-by: .+ <.+@.+>$"; then
    echo -e "${GREEN}✅ DCO check passed - commit is signed off${NC}"
    exit 0
else
    echo -e "${RED}❌ DCO check failed - commit must be signed off${NC}"
    echo -e "${YELLOW}💡 To sign off your commit:${NC}"
    echo "   git commit --amend --signoff"
    echo "   # Or for new commits:"
    echo "   git commit --signoff -m \"your commit message\""
    echo "   # Or add -s flag:"
    echo "   git commit -s"
    echo ""
    echo -e "${YELLOW}🔧 What is DCO?${NC}"
    echo "   Developer Certificate of Origin ensures you have the right to"
    echo "   contribute the code and agree to the project's license terms."
    echo "   Format: Signed-off-by: Your Name <your.email@example.com>"
    echo ""
    echo -e "${YELLOW}📋 Your current commit message:${NC}"
    echo "---"
    cat "$COMMIT_MSG_FILE"
    echo "---"
    echo ""
    echo -e "${YELLOW}🔧 To bypass this check (not recommended it will fail PR build):${NC}"
    echo "   git commit --no-verify"
    exit 1
fi