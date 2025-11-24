#!/bin/bash
# publish-dart-sdk.sh - Publish Dart SDK to pub.dev
# This script handles authentication and publishing of the Dart SDK

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Script directory and project root
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
SDK_DIR="$PROJECT_ROOT/sdks/dart"
ENV_FILE="$PROJECT_ROOT/.env"
PUB_CACHE="$HOME/.pub-cache"

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1" >&2
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

log_step() {
    echo -e "${CYAN}[STEP]${NC} $(date '+%Y-%m-%d %H:%M:%S') - $1"
}

# Error handler
error_exit() {
    log_error "$1"
    exit 1
}

# Load PUB_CREDENTIALS from .env file
load_env_credentials() {
    if [ -f "$ENV_FILE" ]; then
        # Extract PUB_CREDENTIALS from .env, handling comments and empty lines
        CREDS=$(grep -v "^#" "$ENV_FILE" | grep "^PUB_CREDENTIALS=" | head -1 | cut -d= -f2- | xargs)
        if [ -n "$CREDS" ]; then
            echo "$CREDS"
            return 0
        fi
    fi
    return 1
}

# Main execution
main() {
    log_info "Starting Dart SDK publishing"
    log_info "Project root: $PROJECT_ROOT"
    log_info "SDK directory: $SDK_DIR"
    echo ""

    # Step 1: Verify SDK directory exists
    log_step "Step 1/5: Verifying SDK directory"
    if [ ! -d "$SDK_DIR" ]; then
        error_exit "SDK directory not found: $SDK_DIR. Run 'make generate-dart-sdk' first."
    fi
    log_success "SDK directory found: $SDK_DIR"
    echo ""

    # Step 2: Verify pubspec.yaml exists
    log_step "Step 2/5: Verifying pubspec.yaml"
    if [ ! -f "$SDK_DIR/pubspec.yaml" ]; then
        error_exit "pubspec.yaml not found in SDK directory"
    fi
    
    PACKAGE_NAME=$(grep "^name:" "$SDK_DIR/pubspec.yaml" | cut -d: -f2 | xargs || echo "unknown")
    PACKAGE_VERSION=$(grep "^version:" "$SDK_DIR/pubspec.yaml" | cut -d: -f2 | xargs || echo "unknown")
    log_success "Package: $PACKAGE_NAME@$PACKAGE_VERSION"
    echo ""

    # Step 3: Check Dart and pub authentication
    log_step "Step 3/5: Checking Dart and pub.dev authentication"
    
    if ! command -v dart &> /dev/null; then
        error_exit "Dart not found. Install Dart SDK first."
    fi
    DART_VERSION=$(dart --version 2>&1 | head -n 1)
    log_success "Dart found: $DART_VERSION"
    
    PUB_CREDENTIALS=""
    AUTH_METHOD=""
    
    # Try loading from .env file
    if CREDS=$(load_env_credentials); then
        PUB_CREDENTIALS="$CREDS"
        AUTH_METHOD=".env file"
        log_success "Found PUB_CREDENTIALS in .env file"
    # Try environment variable
    elif [ -n "$PUB_CREDENTIALS" ]; then
        AUTH_METHOD="environment variable"
        log_success "Found PUB_CREDENTIALS in environment"
    # Check pub token
    elif dart pub token list 2>/dev/null | grep -q "pub.dev"; then
        AUTH_METHOD="pub token"
        log_success "Found existing pub.dev token"
    else
        error_exit "No pub.dev authentication found. Options:\n  1. Add PUB_CREDENTIALS='{\"accessToken\":\"...\"}' to .env file\n  2. Set PUB_CREDENTIALS environment variable\n  3. Run: dart pub token add https://pub.dev"
    fi
    echo ""

    # Step 4: Configure pub.dev authentication
    log_step "Step 4/5: Configuring pub.dev authentication"
    
    if [ -n "$PUB_CREDENTIALS" ]; then
        mkdir -p "$PUB_CACHE"
        echo "$PUB_CREDENTIALS" > "$PUB_CACHE/credentials.json"
        log_success "Configured pub.dev authentication using $AUTH_METHOD"
    else
        log_info "Using existing pub.dev token"
    fi
    echo ""

    # Step 5: Publish package
    log_step "Step 5/5: Publishing package to pub.dev"
    log_info "Package: $PACKAGE_NAME@$PACKAGE_VERSION"
    log_info "Registry: https://pub.dev"
    echo ""

    cd "$SDK_DIR"
    
    # Verify package first
    log_info "Verifying package..."
    if ! dart pub publish --dry-run 2>&1 | tee /tmp/pub-publish-dart-dry.log; then
        log_warn "Package verification had warnings, but continuing..."
    fi
    echo ""

    # Actual publish
    log_info "Publishing package (this may take a moment)..."
    if dart pub publish --force 2>&1 | tee /tmp/pub-publish-dart.log; then
        echo ""
        log_success "Package published successfully!"
        log_info "View package at: https://pub.dev/packages/$PACKAGE_NAME"
    else
        EXIT_CODE=$?
        echo ""
        log_error "Publishing failed with exit code: $EXIT_CODE"
        
        # Check for common errors
        if grep -qi "unauthorized\|401\|403" /tmp/pub-publish-dart.log 2>/dev/null; then
            log_error "Authentication error detected"
            log_error "Check your PUB_CREDENTIALS or pub.dev token"
        elif grep -qi "already exists\|409" /tmp/pub-publish-dart.log 2>/dev/null; then
            log_error "Package version already exists"
            log_error "Update version in sdks/version.json and run 'make version-dart-sdk'"
        fi
        
        if [ -f /tmp/pub-publish-dart.log ]; then
            log_error "Full log saved to: /tmp/pub-publish-dart.log"
        fi
        
        # Cleanup credentials if we created them
        if [ -n "$PUB_CREDENTIALS" ] && [ -f "$PUB_CACHE/credentials.json" ]; then
            # Only remove if it's the one we created (check if it matches)
            if [ "$(cat "$PUB_CACHE/credentials.json")" = "$PUB_CREDENTIALS" ]; then
                rm -f "$PUB_CACHE/credentials.json"
            fi
        fi
        
        exit $EXIT_CODE
    fi

    # Cleanup credentials if we created them
    if [ -n "$PUB_CREDENTIALS" ] && [ -f "$PUB_CACHE/credentials.json" ]; then
        if [ "$(cat "$PUB_CACHE/credentials.json")" = "$PUB_CREDENTIALS" ]; then
            rm -f "$PUB_CACHE/credentials.json"
            log_info "Cleaned up temporary credentials"
        fi
    fi
    echo ""

    log_success "Dart SDK publishing complete!"
    log_info "Package: $PACKAGE_NAME@$PACKAGE_VERSION"
    log_info "URL: https://pub.dev/packages/$PACKAGE_NAME"
    echo ""
}

# Run main function
main "$@"

