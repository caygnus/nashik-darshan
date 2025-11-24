.PHONY: migrate-ent
migrate-ent:
	@echo "Running Ent migrations..."
	@go run cmd/migrate/main.go
	@echo "✅ Ent migrations complete"

.PHONY: generate-ent
generate-ent:
	@echo "Generating Ent schema..."
	@go run scripts/main.go -cmd generate-ent
	@echo "✅ Ent schema generated"

.PHONY: generate-keys
generate-keys:
	@echo "Generating encrypted RSA key pair for production..."
	@./scripts/generate-keys.sh
	@echo "✅ Encrypted key pair generated"

.PHONY: generate-dev-keys
generate-dev-keys:
	@echo "Generating unencrypted RSA key pair for development..."
	@./scripts/generate-dev-keys.sh
	@echo "✅ Development key pair generated"

.PHONY: install-hooks
install-hooks:
	@echo "Installing Git hooks..."
	@git config core.hooksPath .githooks
	@chmod +x .githooks/pre-commit
	@echo "✅ Git hooks installed"

.PHONY: run-hooks
run-hooks:
	@echo "Running Git hooks..."
	@.githooks/pre-commit

.PHONY: lint-fix
lint-fix:
	@echo "🧼 Running gofmt to auto-format..."
	gofmt -s -w .
	@echo "🧹 Running golangci-lint with --fix (autofix where possible)..."
	golangci-lint run --fix || true
	@echo "✅ Lint fixes applied (where possible)."


.PHONY: swagger-clean
swagger-clean:
	rm -rf docs/swagger

.PHONY: install-swag
install-swag:
	@which swag > /dev/null || (go install github.com/swaggo/swag/cmd/swag@latest)

.PHONY: swagger
swagger: swagger-2-0 swagger-3-0

.PHONY: swagger-2-0
swagger-2-0: install-swag
	$(shell go env GOPATH)/bin/swag init \
		--generalInfo cmd/server/main.go \
		--dir . \
		--parseDependency \
		--parseInternal \
		--output docs/swagger \
		--generatedTime=false \
		--parseDepth 1 \
		--instanceName swagger \
		--parseVendor \
		--outputTypes go,json,yaml
	@make swagger-fix-refs

.PHONY: swagger-3-0
swagger-3-0: install-swag
	@echo "Converting Swagger 2.0 to OpenAPI 3.0..."
	@curl -X 'POST' \
		'https://converter.swagger.io/api/convert' \
		-H 'accept: application/json' \
		-H 'Content-Type: application/json' \
		-d @docs/swagger/swagger.json > docs/swagger/swagger-3-0.json
	@echo "Conversion complete. Output saved to docs/swagger/swagger-3-0.json"

.PHONY: swagger-fix-refs
swagger-fix-refs:
	@./scripts/fix_swagger_refs.sh


.PHONY: run
run:
	@echo "Running development server..."
	@go run cmd/server/main.go
	@echo "✅ Development server running"

.PHONY: build
build:
	@echo "Running production server..."
	@go build cmd/server/main.go
	@echo "✅ Production server running"

# ============================================================================
# SDK Generation Targets
# ============================================================================
# Directory structure:
#   docs/swagger/swagger.yaml  - OpenAPI specification file (required)
#   sdks/ts/                  - Generated TypeScript SDK
#   sdks/dart/                - Generated Dart SDK

# OpenAPI specification file path
OPENAPI_SPEC := docs/swagger/swagger.yaml

# SDK output directories
SDK_TS_DIR := sdks/ts
SDK_DART_DIR := sdks/dart

.PHONY: install-deps
install-deps:
	@echo "📦 Installing dependencies..."
	@bash -c 'set -e; \
	if ! which openapi-generator-cli &> /dev/null; then \
		echo "Installing openapi-generator-cli globally..."; \
		npm install -g @openapitools/openapi-generator-cli || (echo "❌ Failed to install openapi-generator-cli" && exit 1); \
	else \
		echo "✓ openapi-generator-cli already installed"; \
	fi'
	@echo "✅ Dependencies installed"

.PHONY: check-env
check-env:
	@echo "🔍 Checking environment and dependencies..."
	@bash -c 'set -e; \
	bash scripts/assert.sh command node 18; \
	bash scripts/assert.sh command npm; \
	bash scripts/assert.sh command java; \
	bash scripts/assert.sh command dart; \
	bash scripts/assert.sh command openapi-generator-cli; \
	bash scripts/assert.sh file $(OPENAPI_SPEC)'
	@echo "✅ Environment check passed"

.PHONY: setup-sdk-dirs
setup-sdk-dirs:
	@echo "📁 Setting up SDK directories..."
	@mkdir -p $(SDK_TS_DIR) $(SDK_DART_DIR)
	@if [ ! -f $(SDK_TS_DIR)/.openapi-generator-ignore ]; then \
		echo "Creating .openapi-generator-ignore for TypeScript SDK..."; \
		cp sdks/ts/.openapi-generator-ignore $(SDK_TS_DIR)/.openapi-generator-ignore 2>/dev/null || \
		echo -e "# OpenAPI Generator Ignore File for TypeScript SDK\npackage.json\ntsconfig.json\njest.config.js\n.eslintrc.js\n.gitignore\n.npmignore\nREADME.md\nexamples/\ngit_push.sh" > $(SDK_TS_DIR)/.openapi-generator-ignore; \
	fi
	@if [ ! -f $(SDK_DART_DIR)/.openapi-generator-ignore ]; then \
		echo "Creating .openapi-generator-ignore for Dart SDK..."; \
		cp sdks/dart/.openapi-generator-ignore $(SDK_DART_DIR)/.openapi-generator-ignore 2>/dev/null || \
		echo -e "# OpenAPI Generator Ignore File for Dart SDK\npubspec.yaml\nanalysis_options.yaml\nREADME.md\n.gitignore\nexample/\ngit_push.sh" > $(SDK_DART_DIR)/.openapi-generator-ignore; \
	fi
	@echo "✅ SDK directories ready"

.PHONY: generate-ts-sdk
generate-ts-sdk: check-env setup-sdk-dirs
	@echo "🔧 Generating TypeScript SDK..."
	@bash -c 'set -e; \
	mkdir -p $(SDK_TS_DIR); \
	openapi-generator-cli generate \
		-i $(OPENAPI_SPEC) \
		-g typescript-axios \
		-o $(SDK_TS_DIR) || (echo "❌ TypeScript SDK generation failed" && exit 1)'
	@echo "✅ TypeScript SDK generated at $(SDK_TS_DIR)"

.PHONY: generate-dart-sdk
generate-dart-sdk: check-env setup-sdk-dirs
	@echo "🔧 Generating Dart SDK..."
	@bash -c 'set -e; \
	mkdir -p $(SDK_DART_DIR); \
	openapi-generator-cli generate \
		-i $(OPENAPI_SPEC) \
		-g dart-dio \
		-o $(SDK_DART_DIR) || (echo "❌ Dart SDK generation failed" && exit 1)'
	@echo "✅ Dart SDK generated at $(SDK_DART_DIR)"

.PHONY: generate-sdks
generate-sdks: install-deps check-env generate-ts-sdk generate-dart-sdk
	@echo "✅ All SDKs generated successfully!"
	@echo "📁 TypeScript SDK: $(SDK_TS_DIR)"
	@echo "📁 Dart SDK: $(SDK_DART_DIR)"

.PHONY: clean-sdks
clean-sdks:
	@echo "🧹 Cleaning generated SDKs..."
	@rm -rf $(SDK_TS_DIR) $(SDK_DART_DIR)
	@echo "✅ SDK directories cleaned"

.PHONY: verify-sdks
verify-sdks:
	@echo "🔍 Verifying generated SDKs..."
	@bash -c 'set -e; \
	if [ ! -d "$(SDK_TS_DIR)" ]; then \
		echo "❌ TypeScript SDK directory not found"; \
		exit 1; \
	fi; \
	if [ ! -f "$(SDK_TS_DIR)/package.json" ] && [ ! -f "$(SDK_TS_DIR)/index.ts" ]; then \
		echo "⚠️  TypeScript SDK appears incomplete"; \
	fi; \
	if [ ! -d "$(SDK_DART_DIR)" ]; then \
		echo "❌ Dart SDK directory not found"; \
		exit 1; \
	fi; \
	if [ ! -f "$(SDK_DART_DIR)/pubspec.yaml" ] && [ ! -f "$(SDK_DART_DIR)/lib/openapi.dart" ]; then \
		echo "⚠️  Dart SDK appears incomplete"; \
	fi; \
	echo "✅ SDK verification complete"'

.PHONY: version-ts-sdk
version-ts-sdk:
	@echo "📦 Updating TypeScript SDK version..."
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make version-ts-sdk VERSION=1.0.1"; \
		exit 1; \
	fi
	@bash -c 'set -e; \
	cd $(SDK_TS_DIR); \
	npm version $(VERSION) --no-git-tag-version || \
	(node -e "const fs=require(\"fs\"); const pkg=JSON.parse(fs.readFileSync(\"package.json\")); pkg.version=\"$(VERSION)\"; fs.writeFileSync(\"package.json\", JSON.stringify(pkg, null, 2));")'
	@echo "✅ TypeScript SDK version updated to $(VERSION)"

.PHONY: version-dart-sdk
version-dart-sdk:
	@echo "📦 Updating Dart SDK version..."
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make version-dart-sdk VERSION=1.0.1"; \
		exit 1; \
	fi
	@bash -c 'set -e; \
	cd $(SDK_DART_DIR); \
	sed -i.bak "s/^version: .*/version: $(VERSION)/" pubspec.yaml && rm -f pubspec.yaml.bak'
	@echo "✅ Dart SDK version updated to $(VERSION)"

.PHONY: version-sdks
version-sdks:
	@echo "📦 Updating SDK versions..."
	@if [ -z "$(VERSION)" ]; then \
		echo "❌ VERSION is required. Usage: make version-sdks VERSION=1.0.1"; \
		exit 1; \
	fi
	@$(MAKE) version-ts-sdk VERSION=$(VERSION)
	@$(MAKE) version-dart-sdk VERSION=$(VERSION)
	@echo "✅ All SDK versions updated to $(VERSION)"

.PHONY: publish-ts-sdk
publish-ts-sdk: verify-sdks
	@echo "📤 Publishing TypeScript SDK to npm..."
	@bash -c 'set -e; \
	cd $(SDK_TS_DIR); \
	if [ -z "$$NPM_TOKEN" ]; then \
		echo "⚠️  NPM_TOKEN not set. Run: npm login"; \
		npm publish --dry-run || npm publish; \
	else \
		echo "//registry.npmjs.org/:_authToken=$$NPM_TOKEN" > .npmrc; \
		npm publish; \
		rm -f .npmrc; \
	fi'
	@echo "✅ TypeScript SDK published"

.PHONY: publish-dart-sdk
publish-dart-sdk: verify-sdks
	@echo "📤 Publishing Dart SDK to pub.dev..."
	@bash -c 'set -e; \
	cd $(SDK_DART_DIR); \
	if ! pub token list 2>/dev/null | grep -q "pub.dev"; then \
		echo "⚠️  Not authenticated with pub.dev. Run: pub token add https://pub.dev"; \
		echo "   Or set PUB_CREDENTIALS environment variable"; \
	fi; \
	pub publish --dry-run || pub publish'
	@echo "✅ Dart SDK published"

.PHONY: publish-sdks
publish-sdks: verify-sdks
	@echo "📤 Publishing all SDKs..."
	@$(MAKE) publish-ts-sdk
	@$(MAKE) publish-dart-sdk
	@echo "✅ All SDKs published"

.PHONY: publish-ts-sdk-dry-run
publish-ts-sdk-dry-run: verify-sdks
	@echo "🔍 Dry-run: TypeScript SDK publish..."
	@bash -c 'set -e; cd $(SDK_TS_DIR); npm publish --dry-run'
	@echo "✅ Dry-run complete"

.PHONY: publish-dart-sdk-dry-run
publish-dart-sdk-dry-run: verify-sdks
	@echo "🔍 Dry-run: Dart SDK publish..."
	@bash -c 'set -e; cd $(SDK_DART_DIR); pub publish --dry-run'
	@echo "✅ Dry-run complete"