.PHONY: bump-version changelog tag help \
        run-infra stop-infra run-backend stop-backend \
        run-web stop-web migrate \
        run-android build-android run-ios open-ios \
        run-monolith test-monolith dev test-backend doctor

## bump-version patch|minor|major
##   Increments the version, tags the repo, and updates version.txt.
##   Usage: make bump-version patch
bump-version:
	@if [ -z "$(word 2, $(MAKECMDGOALS))" ]; then \
		echo "Usage: make bump-version <patch|minor|major>"; exit 1; \
	fi
	@BUMP=$(word 2, $(MAKECMDGOALS)); \
	CURRENT=$$(cat version.txt | tr -d '[:space:]'); \
	MAJOR=$$(echo $$CURRENT | cut -d. -f1); \
	MINOR=$$(echo $$CURRENT | cut -d. -f2); \
	PATCH=$$(echo $$CURRENT | cut -d. -f3); \
	case "$$BUMP" in \
		patch) PATCH=$$((PATCH+1)) ;; \
		minor) MINOR=$$((MINOR+1)); PATCH=0 ;; \
		major) MAJOR=$$((MAJOR+1)); MINOR=0; PATCH=0 ;; \
		*) echo "Unknown bump type: $$BUMP. Use patch, minor, or major."; exit 1 ;; \
	esac; \
	NEW="$$MAJOR.$$MINOR.$$PATCH"; \
	echo "Bumping $$CURRENT → $$NEW"; \
	echo "$$NEW" > version.txt; \
	if grep -q 'appVersion' mobile/gradle/libs.versions.toml 2>/dev/null; then \
		sed -i.bak "s/appVersion = \".*\"/appVersion = \"$$NEW\"/" mobile/gradle/libs.versions.toml && rm -f mobile/gradle/libs.versions.toml.bak; \
	fi; \
	git add version.txt mobile/gradle/libs.versions.toml 2>/dev/null || git add version.txt; \
	git commit -m "chore(release): bump version to $$NEW"; \
	git tag -a "v$$NEW" -m "Release v$$NEW"

## changelog
##   Generates CHANGELOG.md using git-cliff (must be installed).
changelog:
	@if command -v git-cliff >/dev/null 2>&1; then \
		git-cliff --output CHANGELOG.md; \
		echo "CHANGELOG.md updated."; \
	else \
		echo "git-cliff not found. Install: cargo install git-cliff"; exit 1; \
	fi

## tag
##   Shows all existing version tags.
tag:
	@git tag --sort=-version:refname | head -20

patch minor major:
	@: # Absorb sub-target so make doesn't complain

# ── Local Development ─────────────────────────────────────────────────────────

## run-infra
##   Start PostgreSQL and Redis via Docker Compose.
run-infra:
	docker compose -f infra/docker-compose.yml up -d postgres redis

## stop-infra
##   Stop PostgreSQL and Redis.
stop-infra:
	docker compose -f infra/docker-compose.yml stop postgres redis

## run-backend
##   Start infra, run all migrations, then build and start all backend services.
run-backend:
	docker compose -f infra/docker-compose.yml up -d postgres redis
	@echo "Waiting for postgres to be healthy..."
	@until docker compose -f infra/docker-compose.yml exec -T postgres pg_isready -U preuni -d preuni >/dev/null 2>&1; do sleep 1; done
	@$(MAKE) migrate
	docker compose -f infra/docker-compose.yml up --build -d \
		monolith gateway

## stop-backend
##   Stop all backend services.
stop-backend:
	docker compose -f infra/docker-compose.yml stop \
		monolith gateway

## migrate
##   Run all SQL migrations against the local postgres instance.
##   Requires 'make run-infra' to be running first.
migrate:
	@find infra/migrations -name "*.sql" | sort | while read f; do \
		echo "→ $$f"; \
		docker compose -f infra/docker-compose.yml exec -T postgres \
			psql -U preuni -d preuni < "$$f" || exit 1; \
	done
	@echo "Migrations complete."

## run-web
##   Start the Kotlin/Wasm web app in the browser (requires JDK 17+ and Node.js 20+).
##   The app opens at http://localhost:3000 and proxies /v1/* to the NGINX gateway on :8080.
run-web:
	@JAVA_HOME_CANDIDATE=$$(/usr/libexec/java_home -v 23 2>/dev/null || /usr/libexec/java_home -v 21 2>/dev/null || /usr/libexec/java_home -v 17 2>/dev/null || true); \
	if [ -z "$$JAVA_HOME_CANDIDATE" ]; then \
		echo "No compatible JDK found. Install JDK 17, 21, or 23."; exit 1; \
	fi; \
	cd mobile && JAVA_HOME="$$JAVA_HOME_CANDIDATE" PATH="$$JAVA_HOME_CANDIDATE/bin:$$PATH" ./gradlew :webApp:wasmJsBrowserDevelopmentRun

## stop-web
##   Kill the webpack dev server and Gradle daemon started by run-web.
stop-web:
	@PIDS=$$(lsof -ti TCP:3000 2>/dev/null); \
	if [ -n "$$PIDS" ]; then \
		echo "$$PIDS" | xargs kill -9 2>/dev/null; \
		echo "Web server stopped."; \
	else \
		echo "No web server running."; \
	fi

## run-android
##   Install and launch the debug APK on a connected device or running emulator.
##   Start an emulator first: emulator -avd <avd_name> &
run-android:
	@JAVA_HOME_CANDIDATE=$$(/usr/libexec/java_home -v 23 2>/dev/null || /usr/libexec/java_home -v 21 2>/dev/null || /usr/libexec/java_home -v 17 2>/dev/null || true); \
	if [ -z "$$JAVA_HOME_CANDIDATE" ]; then \
		echo "No compatible JDK found. Install JDK 17, 21, or 23."; exit 1; \
	fi; \
	cd mobile && JAVA_HOME="$$JAVA_HOME_CANDIDATE" PATH="$$JAVA_HOME_CANDIDATE/bin:$$PATH" ./gradlew :androidApp:installDebug

## build-android
##   Build the debug APK without installing (outputs to androidApp/build/outputs/apk/).
build-android:
	@JAVA_HOME_CANDIDATE=$$(/usr/libexec/java_home -v 23 2>/dev/null || /usr/libexec/java_home -v 21 2>/dev/null || /usr/libexec/java_home -v 17 2>/dev/null || true); \
	if [ -z "$$JAVA_HOME_CANDIDATE" ]; then \
		echo "No compatible JDK found. Install JDK 17, 21, or 23."; exit 1; \
	fi; \
	cd mobile && JAVA_HOME="$$JAVA_HOME_CANDIDATE" PATH="$$JAVA_HOME_CANDIDATE/bin:$$PATH" ./gradlew :androidApp:assembleDebug

## run-ios
##   Build and launch the iOS app on the iPhone 17 simulator.
##   Xcode's build phases handle the KMP framework compilation automatically.
##   Requires Xcode 16+ and the iOS 18 simulator runtime.
run-ios:
	@JAVA_HOME_CANDIDATE=$$(/usr/libexec/java_home -v 23 2>/dev/null || /usr/libexec/java_home -v 21 2>/dev/null || /usr/libexec/java_home -v 17 2>/dev/null || true); \
	if [ -z "$$JAVA_HOME_CANDIDATE" ]; then \
		echo "No compatible JDK found. Install JDK 17, 21, or 23."; exit 1; \
	fi; \
	IOS_SIMULATOR="iPhone 17"; \
	IOS_BUNDLE_ID="com.preuni.app"; \
	DERIVED_DATA_PATH="$$PWD/mobile/iosApp/.build/DerivedData"; \
	IOS_APP_PATH="$$DERIVED_DATA_PATH/Build/Products/Debug-iphonesimulator/iosApp.app"; \
	open -a Simulator; \
	xcrun simctl boot "$$IOS_SIMULATOR" >/dev/null 2>&1 || true; \
	xcrun simctl bootstatus "$$IOS_SIMULATOR" -b; \
	cd mobile/iosApp && JAVA_HOME="$$JAVA_HOME_CANDIDATE" PATH="$$JAVA_HOME_CANDIDATE/bin:$$PATH" \
		xcodebuild \
		-scheme iosApp \
		-destination "platform=iOS Simulator,name=$$IOS_SIMULATOR" \
		-allowProvisioningUpdates \
		-derivedDataPath "$$DERIVED_DATA_PATH" \
		build || exit 1; \
	if [ ! -d "$$IOS_APP_PATH" ]; then \
		echo "Built app not found at $$IOS_APP_PATH"; exit 1; \
	fi; \
	xcrun simctl install booted "$$IOS_APP_PATH"; \
	xcrun simctl terminate booted "$$IOS_BUNDLE_ID" >/dev/null 2>&1 || true; \
	xcrun simctl launch booted "$$IOS_BUNDLE_ID"

## open-ios
##   Open the iosApp Xcode project in Xcode (requires Xcode 16+).
open-ios:
	open mobile/iosApp/iosApp.xcodeproj

## run-monolith
##   Run the unified backend monolith locally (requires infra + migrations).
run-monolith:
	cd backend/app && go run ./cmd/server

## test-monolith
##   Run monolith tests (unit + contract + integration).
test-monolith:
	cd backend/app && go test ./...

## test-backend
##   Run all backend Go tests via the workspace.
test-backend:
	cd backend && go test ./...

## dev
##   Bring up local infra + monolith + gateway for development.
dev:
	docker compose -f infra/docker-compose.yml up -d postgres redis
	@echo "Waiting for postgres to be healthy..."
	@until docker compose -f infra/docker-compose.yml exec -T postgres pg_isready -U preuni -d preuni >/dev/null 2>&1; do sleep 1; done
	@$(MAKE) migrate
	docker compose -f infra/docker-compose.yml up --build -d monolith gateway

## doctor
##   Verify local dev prerequisites (Docker, Go 1.24, Postgres).
doctor:
	@command -v docker >/dev/null 2>&1 || { echo "✗ docker not installed"; exit 1; }
	@docker info >/dev/null 2>&1 || { echo "✗ docker daemon not running"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo "✗ go not installed"; exit 1; }
	@go version | grep -qE 'go1\.(2[4-9]|[3-9][0-9])' || { echo "✗ go 1.24+ required"; exit 1; }
	@echo "✓ docker running, go 1.24+ available"

help:
	@grep -E '^## ' Makefile | sed 's/## //'
