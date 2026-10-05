#!/usr/bin/env bash
set -euo pipefail

ref_type=${1:?Expected ref type}
ref_name=${2:?Expected ref name}

if [[ "$ref_type" == tag ]]; then
    number='(0|[1-9][0-9]*)'
    identifier='(0|[1-9][0-9]*|[0-9]*[A-Za-z-][0-9A-Za-z-]*)'
    release_pattern="^v${number}\.${number}\.${number}(-${identifier}(\.${identifier})*)?$"
    if [[ ! "$ref_name" =~ $release_pattern ]] || (( ${#ref_name} > 128 )); then
        echo "Invalid release tag: expected vMAJOR.MINOR.PATCH with an optional SemVer prerelease (no build metadata)." >&2
        exit 1
    fi
    major=${BASH_REMATCH[1]}
    minor=${BASH_REMATCH[2]}
    if [[ "$(git cat-file -t "refs/tags/$ref_name")" != tag ]]; then
        echo "Release tags must be annotated." >&2
        exit 1
    fi
    if ! git merge-base --is-ancestor HEAD origin/main; then
        echo "Release commit must belong to origin/main." >&2
        exit 1
    fi
    version=$ref_name
    stable=true
    if [[ "$version" == *-* ]]; then
        stable=false
    fi
    major_alias=false
    if [[ "$stable" == true && "$major" != 0 ]]; then
        major_alias=true
    fi
    printf 'version=%s\nstable=%s\nmajor_alias=%s\nmajor=v%s\nminor=v%s.%s\n' \
        "$version" "$stable" "$major_alias" "$major" "$major" "$minor"
elif [[ "$ref_type" == branch && "$ref_name" == main ]]; then
    version=$(git describe --tags --match 'v[0-9]*' --always --dirty)
    printf 'version=%s\nstable=false\nmajor_alias=false\nmajor=\nminor=\n' "$version"
else
    echo "Only main builds and validated release tags can publish images." >&2
    exit 1
fi
