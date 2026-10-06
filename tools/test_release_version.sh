#!/usr/bin/env bash
set -euo pipefail

validator="$(cd "$(dirname "$0")" && pwd)/resolve_release_version.sh"
workspace=$(mktemp -d)
trap 'rm -rf -- "$workspace"' EXIT
export GIT_CONFIG_NOSYSTEM=1 GIT_CONFIG_GLOBAL=/dev/null
export GIT_AUTHOR_NAME="Release test" GIT_COMMITTER_NAME="Release test"
export GIT_AUTHOR_EMAIL="release-test@example.invalid" GIT_COMMITTER_EMAIL="release-test@example.invalid"

git init -q "$workspace"
cd "$workspace"
git -c commit.gpgsign=false commit -q --allow-empty -m initial
commit=$(git rev-parse HEAD)
git update-ref refs/remotes/origin/main "$commit"

expect_valid() {
    local tag=$1 prerelease=$2 actual expected
    git -c tag.gpgsign=false tag -a "$tag" -m "$tag"
    actual=$(bash "$validator" "$tag")
    expected=$(printf 'version=%s\nprerelease=%s\ncommit=%s' "$tag" "$prerelease" "$commit")
    if [[ "$actual" != "$expected" ]]; then
        echo "Unexpected validator output for $tag: $actual" >&2
        exit 1
    fi
}

expect_invalid() {
    local tag=$1 message=$2
    if bash "$validator" "$tag" >output 2>error; then
        echo "Validator accepted invalid release $tag." >&2
        exit 1
    fi
    if [[ -s output ]] || ! grep -qF "$message" error; then
        echo "Unexpected rejection for $tag:" >&2
        cat output error >&2
        exit 1
    fi
}

expect_valid v1.2.3 false
expect_valid v0.2.0 false
expect_valid v1.2.3-rc.1 true
expect_valid v1.2.3-0 true
expect_valid v1.2.3+build-1 false
expect_valid v1.2.3-rc.2+build.1 true
for tag in v01.2.3 v1.2 v1.2.3-01 v1.2.3-rc..1 v1.2.3+ v1.2.3+build..1 1.2.3; do
    expect_invalid "$tag" "Invalid release tag"
done
git -c tag.gpgsign=false tag v2.0.0
expect_invalid v2.0.0 "must be annotated"

git -c commit.gpgsign=false commit -q --allow-empty -m outside-main
git -c tag.gpgsign=false tag -a v3.0.0 -m v3.0.0
expect_invalid v3.0.0 "must belong to origin/main"
expect_invalid v1.2.3 "checked-out commit must match"
echo "Release version validation tests passed."
