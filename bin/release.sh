#!/usr/bin/env bash

set -eo pipefail

version=$1

gh release create "$version" --generate-notes
gh release upload "$version" param.go --clobber