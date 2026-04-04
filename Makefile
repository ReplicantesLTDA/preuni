.PHONY: bump-version changelog tag help

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

help:
	@grep -E '^## ' Makefile | sed 's/## //'
