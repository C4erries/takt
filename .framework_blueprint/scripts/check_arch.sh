#!/usr/bin/env bash
set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

ERRORS=0

echo -e "${BLUE}=== [ARCH LINTER] Starting Architecture Integrity Check ===${NC}"

# -----------------------------------------------------------------------------
# Check 1: File Line Limit (< 400 lines)
# -----------------------------------------------------------------------------
echo -e "\n${BLUE}--> [Check 1/5] Checking file line limit (< 400 lines)...${NC}"
MAX_LINES=400

# Scan all relevant code files excluding generated files, vendor, node_modules, lockfiles
while IFS= read -r file; do
    # Skip generated files and assets
    if [[ "$file" =~ \.gen\. || "$file" =~ _gen\. || "$file" =~ package-lock\.json || "$file" =~ \.min\. || "$file" =~ \.svg$ ]]; then
        continue
    fi

    lines=$(wc -l < "$file" || echo 0)
    if [ "$lines" -ge "$MAX_LINES" ]; then
        echo -e "${RED}[FAIL]${NC} File exceeds $MAX_LINES lines ($lines lines): $file"
        ERRORS=$((ERRORS + 1))
    fi
done < <(find backend frontend api scripts tests -type f \( -name "*.go" -o -name "*.ts" -o -name "*.tsx" -o -name "*.js" -o -name "*.jsx" -o -name "*.sh" -o -name "*.py" -o -name "*.yaml" -o -name "*.yml" \) \
    ! -path "*/node_modules/*" \
    ! -path "*/.git/*" \
    ! -path "*/dist/*" \
    ! -path "*/build/*" \
    ! -path "*/.venv/*" \
    ! -path "*/vendor/*" 2>/dev/null || true)

# -----------------------------------------------------------------------------
# Check 2: Package File Count Limit (<= 7 files per package)
# -----------------------------------------------------------------------------
echo -e "\n${BLUE}--> [Check 2/5] Checking package file count limit (<= 7 working files)...${NC}"
MAX_FILES=7

check_dir_file_count() {
    local target_root="$1"
    if [ ! -d "$target_root" ]; then
        return
    fi

    # Find directories inside target_root
    while IFS= read -r dir; do
        # Count non-test, non-hidden, non-generated working files
        count=$(find "$dir" -maxdepth 1 -type f \
            ! -name "*_test.go" \
            ! -name "*.test.ts" \
            ! -name "*.test.tsx" \
            ! -name "*.spec.ts" \
            ! -name "*.spec.tsx" \
            ! -name "*.gen.*" \
            ! -name "*_gen.*" \
            ! -name ".*" 2>/dev/null | wc -l || echo 0)

        if [ "$count" -gt "$MAX_FILES" ]; then
            echo -e "${RED}[FAIL]${NC} Package exceeds $MAX_FILES working files ($count files): $dir"
            ERRORS=$((ERRORS + 1))
        fi
    done < <(find "$target_root" -mindepth 1 -type d \
        ! -path "*/node_modules*" \
        ! -path "*/.git*" \
        ! -path "*/dist*" \
        ! -path "*/shared/components/ui*" 2>/dev/null || true)
}

# Check backend/internal (domain, application, infra, delivery) and frontend/src/features
if [ -d "backend/internal" ]; then
    check_dir_file_count "backend/internal"
fi
if [ -d "frontend/src/features" ]; then
    check_dir_file_count "frontend/src/features"
fi

# -----------------------------------------------------------------------------
# Check 3: Clean Domain Layer (no external framework/SQL/HTTP imports)
# -----------------------------------------------------------------------------
echo -e "\n${BLUE}--> [Check 3/5] Checking Domain layer purity (internal/domain)...${NC}"
if [ -d "backend/internal/domain" ]; then
    MODULE_NAME=""
    if [ -f "backend/go.mod" ]; then
        MODULE_NAME=$(grep -m 1 "^module " backend/go.mod | awk '{print $2}')
    fi

    if [ -n "$MODULE_NAME" ]; then
        FORBIDDEN_IMPORTS_PATTERN="github.com/labstack/echo|net/http|github.com/jackc/pgx|database/sql|${MODULE_NAME}/internal/application|${MODULE_NAME}/internal/infrastructure"
    else
        FORBIDDEN_IMPORTS_PATTERN="github.com/labstack/echo|net/http|github.com/jackc/pgx|database/sql|/internal/application|/internal/infrastructure"
    fi
    
    while IFS= read -r file; do
        # Exclude _test.go files from strict domain checks if needed, but domain shouldn't need them either
        if grep -E "$FORBIDDEN_IMPORTS_PATTERN" "$file" >/dev/null 2>&1; then
            matched=$(grep -E "$FORBIDDEN_IMPORTS_PATTERN" "$file")
            echo -e "${RED}[FAIL]${NC} Domain layer violation in $file: forbidden import found:"
            echo "       $matched"
            ERRORS=$((ERRORS + 1))
        fi
    done < <(find backend/internal/domain -type f -name "*.go" ! -path "*/shared/*" 2>/dev/null || true)
fi

# -----------------------------------------------------------------------------
# Check 4: No Raw Tailwind Colors in features and widgets
# -----------------------------------------------------------------------------
echo -e "\n${BLUE}--> [Check 4/5] Checking for raw Tailwind colors in features/ and widgets/...${NC}"
RAW_COLOR_PATTERN="(bg|text|border|fill|stroke)-(red|blue|green|yellow|purple|pink|indigo|emerald|teal|cyan|sky|violet|fuchsia|rose|amber|lime|orange|zinc|slate|neutral|stone|gray)-[0-9]+"

for dir in "frontend/src/features" "frontend/src/widgets"; do
    if [ -d "$dir" ]; then
        while IFS= read -r file; do
            if grep -E -n "$RAW_COLOR_PATTERN" "$file" >/dev/null 2>&1; then
                matches=$(grep -E -n "$RAW_COLOR_PATTERN" "$file" | head -n 3)
                echo -e "${RED}[FAIL]${NC} Raw Tailwind color detected in $file (use semantic tokens instead):"
                echo "$matches"
                ERRORS=$((ERRORS + 1))
            fi
        done < <(find "$dir" -type f \( -name "*.tsx" -o -name "*.jsx" -o -name "*.ts" -o -name "*.js" \) 2>/dev/null || true)
    fi
done

# -----------------------------------------------------------------------------
# Check 5: Strict 3 subpackages in backend/internal (domain, application, infrastructure)
# -----------------------------------------------------------------------------
echo -e "\n${BLUE}--> [Check 5/5] Checking backend/internal root subpackages (strictly: domain, application, infrastructure)...${NC}"
if [ -d "backend/internal" ]; then
    ALLOWED_ROOT_DIRS="domain application infrastructure"
    for dir in backend/internal/*/; do
        [ -d "$dir" ] || continue
        basename_dir=$(basename "$dir")
        if [[ ! " $ALLOWED_ROOT_DIRS " =~ " $basename_dir " ]]; then
            echo -e "${RED}[FAIL]${NC} Forbidden directory in backend/internal: '$basename_dir'."
            echo "       Allowed strictly 3 subpackages: domain, application, infrastructure."
            echo "       Move container/di, http, config, db, etc. inside infrastructure/ (e.g. infrastructure/container, infrastructure/http, infrastructure/config)!"
            ERRORS=$((ERRORS + 1))
        fi
    done
fi

# -----------------------------------------------------------------------------
# Summary
# -----------------------------------------------------------------------------
echo -e "\n${BLUE}=== [ARCH LINTER] Summary ===${NC}"
if [ "$ERRORS" -gt 0 ]; then
    echo -e "${RED}Architecture check FAILED with $ERRORS error(s).${NC}"
    exit 1
else
    echo -e "${GREEN}Architecture check PASSED with 0 errors! All constraints satisfied.${NC}"
    exit 0
fi

