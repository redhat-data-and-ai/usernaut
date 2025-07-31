#!/bin/bash

# Pre-commit hook to detect sensitive information
# This script scans for common patterns of sensitive data that should not be committed

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

echo -e "${GREEN}🔍 Scanning for sensitive information...${NC}"

# Get list of files to be committed
FILES=$(git diff --cached --name-only --diff-filter=ACM)

if [ -z "$FILES" ]; then
    echo -e "${GREEN}✅ No files to check${NC}"
    exit 0
fi

# Flag to track if any sensitive data is found
SECRETS_FOUND=0

# Common sensitive patterns - using arrays for better compatibility
PATTERN_NAMES=(
    "API Keys"
    "API Secrets"
    "JWT Tokens"
    "Personal Access Tokens"
    "Passwords"
    "Private Keys"
    "SSH Keys"
    "Database URLs"
    "Generic Secrets"
    "AWS Keys"
    "Google API Keys"
    "GitHub Tokens"
)

PATTERN_REGEXES=(
    "(api[_-]?key|apikey)[[:space:]]*[:=][[:space:]]*['\"]?[a-zA-Z0-9]{20,}['\"]?"
    "(api[_-]?secret|apisecret)[[:space:]]*[:=][[:space:]]*['\"]?[a-zA-Z0-9]{20,}['\"]?"
    "eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+"
    "(pat|token)[[:space:]]*[:=][[:space:]]*['\"]?[a-zA-Z0-9_-]{20,}['\"]?"
    "(password|passwd|pwd)[[:space:]]*[:=][[:space:]]*['\"]?[a-zA-Z0-9!@#$%^&*()_+-=]{8,}['\"]?"
    "\-\-\-\-\-BEGIN [A-Z]+ PRIVATE KEY\-\-\-\-\-"
    "ssh-rsa [A-Za-z0-9+/]+"
    "(postgres|mysql|mongodb)://[a-zA-Z0-9_.-]+:[a-zA-Z0-9_.-]+@[a-zA-Z0-9_.-]+[:/]"
    "(secret|credential)[[:space:]]*[:=][[:space:]]*['\"]?[a-zA-Z0-9]{10,}['\"]?"
    "AKIA[0-9A-Z]{16}"
    "AIza[0-9A-Za-z_-]{35}"
    "gh[ps]_[A-Za-z0-9_]{36,251}"
)

# Specific patterns for appconfig directory
APPCONFIG_PATTERN_NAMES=(
    "Fivetran API Key"
    "Fivetran API Secret"
    "Snowflake PAT"
    "Redis Password"
    "Certificate Paths"
)

APPCONFIG_PATTERN_REGEXES=(
    "apiKey:[[:space:]]*[a-zA-Z0-9]{12,30}"
    "apiSecret:[[:space:]]*[a-zA-Z0-9]{20,50}"
    "pat:[[:space:]]*eyJ[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+"
    "password:[[:space:]]*['\"]?[a-zA-Z0-9!@#$%^&*()_+-=]{1,}['\"]?"
    "(cert_path|private_key_path):[[:space:]]*['\"]?/[a-zA-Z0-9/_.-]+['\"]?"
)

# Function to check a file for patterns
check_file() {
    local file="$1"
    local file_content
    
    if [ ! -f "$file" ]; then
        return 0
    fi
    
    # Get the staged content of the file
    file_content=$(git show ":$file" 2>/dev/null || cat "$file")
    
    echo -e "${YELLOW}📄 Checking: $file${NC}"
    
    # Check general patterns
    for i in "${!PATTERN_NAMES[@]}"; do
        pattern_name="${PATTERN_NAMES[$i]}"
        pattern="${PATTERN_REGEXES[$i]}"
        if echo "$file_content" | grep -qiE "$pattern"; then
            echo -e "${RED}❌ FOUND $pattern_name in $file${NC}"
            echo "$file_content" | grep -niE "$pattern" | head -5
            SECRETS_FOUND=1
        fi
    done
    
    # Check appconfig-specific patterns if file is in appconfig directory
    if [[ "$file" == appconfig/* ]]; then
        echo -e "${YELLOW}🔧 Extra checks for appconfig file${NC}"
        for i in "${!APPCONFIG_PATTERN_NAMES[@]}"; do
            pattern_name="${APPCONFIG_PATTERN_NAMES[$i]}"
            pattern="${APPCONFIG_PATTERN_REGEXES[$i]}"
            if echo "$file_content" | grep -qE "$pattern"; then
                echo -e "${RED}❌ FOUND $pattern_name in $file${NC}"
                echo "$file_content" | grep -nE "$pattern" | head -5
                SECRETS_FOUND=1
            fi
        done
        
        # Special check for hardcoded values that look suspicious
        if echo "$file_content" | grep -qE "(apiKey|apiSecret|pat):[[:space:]]*[a-zA-Z0-9]{8,}"; then
            echo -e "${RED}❌ FOUND hardcoded credentials in $file${NC}"
            echo -e "${YELLOW}💡 Consider using environment variables or external secret files${NC}"
            SECRETS_FOUND=1
        fi
    fi
}

# Check each file
for file in $FILES; do
    check_file "$file"
done

# Check if any sensitive data was found
if [ $SECRETS_FOUND -eq 1 ]; then
    echo -e "${RED}🚨 COMMIT BLOCKED: Sensitive information detected!${NC}"
    echo -e "${YELLOW}💡 Recommendations:${NC}"
    echo "   • Use environment variables for sensitive values"
    echo "   • Store secrets in external files (not tracked by git)"
    echo "   • Use secret management tools like Vault or AWS Secrets Manager"
    echo "   • Add sensitive files to .gitignore"
    echo "   • Consider using 'file|path' references in config files"
    echo ""
    echo -e "${YELLOW}🔧 To bypass this check (not recommended):${NC}"
    echo "   git commit --no-verify"
    exit 1
else
    echo -e "${GREEN}✅ No sensitive information detected. Commit allowed!${NC}"
    exit 0
fi