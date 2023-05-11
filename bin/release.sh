#!/usr/bin/env bash

set -eo pipefail

version=$1

# git tag "$version"
# git push --tags

gh release create "$version" --generate-notes
