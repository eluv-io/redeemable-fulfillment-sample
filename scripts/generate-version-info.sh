#!/bin/bash

# Generates go code with git version information to stdout

# PENDING(PTT) A simpler alternative might be to use build flags:
#   // where "who" is a string declared in package "main"
#   go run -ldflags "-X main.who=Universe" main.go
#
#   see https://lukeeckley.com/post/useful-go-build-flags/

branch=$(git rev-parse --abbrev-ref HEAD)
revision=$(git rev-parse HEAD)
commit_date=$(git show -s --format=%ct) # unix timestamp

# retrieve tagged version
# $version will be empty if the current revision is not tagged
version=$(git describe --tags --exact-match 2> /dev/null)

# last tag if otherwise missing
last_tag=$(git describe --tags 2> /dev/null)

code=$(cat <<EOM
package version

// DO NOT EDIT!
// file generated by scripts/generate-version-info.sh

const (
	commit_date int64 = ${commit_date}
	revision          = "${revision}"
	branch            = "${branch}"
	version           = "${version}"
	last_tag          = "${last_tag}"
)
EOM
)

echo "${code}"
