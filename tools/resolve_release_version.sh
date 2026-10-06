#!/usr/bin/env bash
set -euo pipefail

version=${1:?Expected release tag}
number='(0|[1-9][0-9]*)'
identifier='(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
release_pattern="^v${number}\.${number}\.${number}(-${identifier}(\.${identifier})*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$"
if [[ ! "$version" =~ $release_pattern ]]; then
    echo "Invalid release tag: expected vMAJOR.MINOR.PATCH with optional SemVer prerelease and build metadata." >&2
    exit 1
fi
if [[ "$(git cat-file -t "refs/tags/$version")" != tag ]]; then
    echo "New release tags must be annotated." >&2
    exit 1
fi
commit=$(git rev-parse "refs/tags/$version^{commit}")
if [[ "$(git rev-parse HEAD)" != "$commit" ]]; then
    echo "The checked-out commit must match the release tag." >&2
    exit 1
fi
if ! git merge-base --is-ancestor "$commit" origin/main; then
    echo "Release commit must belong to origin/main." >&2
    exit 1
fi
prerelease=false
if [[ "${version%%+*}" == *-* ]]; then
    prerelease=true
fi
printf 'version=%s\nprerelease=%s\ncommit=%s\n' "$version" "$prerelease" "$commit"
