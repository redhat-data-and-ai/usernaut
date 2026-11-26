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
echo -e "${GREEN}💡 Safe patterns: empty values, 'env|VAR_NAME', 'file|/path/to/file'${NC}"

# Files and directories to exclude from checks
EXCLUDED_PATTERNS=(
    "vendor/"
    "*.pb.go"
    "*_test.go"
    "test/"
)

# Function to check if a file should be excluded
should_exclude_file() {
    local file="$1"
    for pattern in "${EXCLUDED_PATTERNS[@]}"; do
        # Check for path prefix or filename glob match
        if [[ "$pattern" == */ && "$file" == "$pattern"* ]]; then
            return 0 # Path prefix match
        elif [[ "$pattern" != */ && "$(basename -- "$file")" == $pattern ]]; then
            return 0 # Filename glob match
        fi
    done
    return 1  # Should not exclude
}

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

# Function to check if a value is safe (env/file reference or empty)
is_safe_value() {
    local value="$1"
    # Remove quotes and whitespace
    value=$(echo "$value" | sed 's/^[[:space:]]*["\x27]//;s/["\x27][[:space:]]*$//')
    
    # Check if empty
    if [ -z "$value" ]; then
        return 0
    fi
    
    # Check if starts with env| or file|
    if [[ "$value" =~ ^(env\||file\|) ]]; then
        return 0
    fi
    
    return 1
}

# Function to check if a value looks like code reference (not a hardcoded secret)
is_code_reference() {
    local value="$1"
    
    # Remove quotes and whitespace
    value=$(echo "$value" | sed 's/^[[:space:]]*["\x27]//;s/["\x27][[:space:]]*$//')
    
    # Check if value contains dots (like config.Password, settings.apiKey)
    if [[ "$value" =~ \. ]]; then
        return 0
    fi
    
    # Check if value starts with $ (like $PASSWORD, ${PASSWORD})
    if [[ "$value" =~ ^\$ ]]; then
        return 0
    fi
    
    # Check if value is a common variable pattern (camelCase or snake_case identifiers without special chars)
    # This catches things like: myPassword, user_password, configPassword
    # But not actual passwords like: MyP@ssw0rd!, password123!
    if [[ "$value" =~ ^[a-zA-Z_][a-zA-Z0-9_]*$ ]] && [[ ! "$value" =~ [0-9]{3,} ]]; then
        return 0
    fi
    
    return 1
}

# Function to check a file for patterns
check_file() {
    local file="$1"
    local file_content
    
    if [ ! -f "$file" ]; then
        return 0
    fi
    
    # Check if file should be excluded
    if should_exclude_file "$file"; then
        echo -e "${YELLOW}⏭️  Skipping excluded file: $file${NC}"
        return 0
    fi
    
    # Get the staged content of the file
    file_content=$(git show ":$file" 2>/dev/null || cat "$file")
    
    echo -e "${YELLOW}📄 Checking: $file${NC}"
    
    # Check general patterns
    for i in "${!PATTERN_NAMES[@]}"; do
        pattern_name="${PATTERN_NAMES[$i]}"
        pattern="${PATTERN_REGEXES[$i]}"
        
        # Get matching lines
        matching_lines=$(echo "$file_content" | grep -niE "$pattern" || true)
        
        if [ -n "$matching_lines" ]; then
            # For password-related patterns, filter out code references
            if [[ "$pattern_name" == *"Password"* ]] || [[ "$pattern_name" == *"Token"* ]] || [[ "$pattern_name" == *"Key"* ]] || [[ "$pattern_name" == *"Secret"* ]]; then
                # Check each match to see if it's a code reference
                has_real_secret=false
                filtered_matches=""
                
                while IFS= read -r line; do
                    # Extract the value part after the colon or equals
                    value=$(echo "$line" | sed 's/^[0-9]*://; s/.*[:=][[:space:]]*//')
                    
                    # Skip if it's a safe value or code reference
                    if ! is_safe_value "$value" && ! is_code_reference "$value"; then
                        has_real_secret=true
                        filtered_matches="${filtered_matches}${line}\n"
                    fi
                done <<< "$matching_lines"
                
                if [ "$has_real_secret" = true ]; then
                    echo -e "${RED}❌ FOUND $pattern_name in $file${NC}"
                    echo -e "$filtered_matches" | head -5
                    SECRETS_FOUND=1
                fi
            else
                # For other patterns (SSH keys, AWS keys, etc.), show all matches
                echo -e "${RED}❌ FOUND $pattern_name in $file${NC}"
                echo "$matching_lines" | head -5
                SECRETS_FOUND=1
            fi
        fi
    done
    
    # Check appconfig-specific patterns if file is in appconfig directory
    if [[ "$file" == appconfig/* ]]; then
        echo -e "${YELLOW}🔧 Extra checks for appconfig file${NC}"
        for i in "${!APPCONFIG_PATTERN_NAMES[@]}"; do
            pattern_name="${APPCONFIG_PATTERN_NAMES[$i]}"
            pattern="${APPCONFIG_PATTERN_REGEXES[$i]}"
            
            # Get matching lines
            matching_lines=$(echo "$file_content" | grep -nE "$pattern" || true)
            
            if [ -n "$matching_lines" ]; then
                # Check each matching line to see if it's a safe value
                while IFS= read -r line; do
                    # Extract the value part after the colon
                    value=$(echo "$line" | sed 's/^[0-9]*://; s/.*:[[:space:]]*//')
                    
                    if ! is_safe_value "$value"; then
                        echo -e "${RED}❌ FOUND $pattern_name in $file${NC}"
                        echo "$line"
                        SECRETS_FOUND=1
                    fi
                done <<< "$matching_lines"
            fi
        done
        
        # Special check for hardcoded values that look suspicious
        hardcoded_lines=$(echo "$file_content" | grep -nE "(apiKey|apiSecret|pat|password):[[:space:]]*['\"]?[a-zA-Z0-9!@#$%^&*()_+-=]{8,}" || true)
        
        if [ -n "$hardcoded_lines" ]; then
            while IFS= read -r line; do
                # Extract the value part after the colon
                value=$(echo "$line" | sed 's/^[0-9]*://; s/.*:[[:space:]]*//')
                
                if ! is_safe_value "$value"; then
                    echo -e "${RED}❌ FOUND hardcoded credentials in $file${NC}"
                    echo "$line"
                    echo -e "${YELLOW}💡 Consider using 'env|VAR_NAME' or 'file|/path/to/secret'${NC}"
                    SECRETS_FOUND=1
                    break
                fi
            done <<< "$hardcoded_lines"
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
    echo "   • Use 'env|VAR_NAME' to reference environment variables"
    echo "   • Use 'file|/path/to/secret' to reference external files"
    echo "   • Use empty values (\"\") for optional/unset configuration"
    echo "   • Store secrets in external files (not tracked by git)"
    echo "   • Use secret management tools like Vault or AWS Secrets Manager"
    echo "   • Add sensitive files to .gitignore"
    echo ""
    echo -e "${YELLOW}🔧 To bypass this check (not recommended):${NC}"
    echo "   git commit --no-verify"
    exit 1
else
    echo -e "${GREEN}✅ No sensitive information detected. Commit allowed!${NC}"
    exit 0
fi