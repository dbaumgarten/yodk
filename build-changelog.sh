#!/bin/bash
set -e
echo "## Changes since $(git describe --tags --abbrev=0 @^) (auto-generated)"  > CHANGELOG.md
git log '--format=format: - %s' $(git describe --tags --abbrev=0 @^)..@ >> CHANGELOG.md
echo "" >> CHANGELOG.md
cat CHANGELOG.md