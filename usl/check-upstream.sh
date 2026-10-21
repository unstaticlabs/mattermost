#!/usr/bin/env bash
# Shows what changed upstream since the base of the USL branch, and stops
# when a change can break a USL feature or enable a denied feature.
#
# The script only reads. It fetches the upstream tags and release branches,
# and changes no branch, no tag, and no file.
#
# Usage: usl/check-upstream.sh [--no-fetch] [--base <tag>] [<target-ref>]
#
#   <target-ref>  The upstream tag or branch to compare, for example v12.0.1
#                 or upstream/release-12.0. The default is the newest final
#                 tag after the base, else the release branch of the base.
#   --base <tag>  The upstream tag of the USL branch. The default comes from
#                 the branch name: usl/v12.0.0 gives v12.0.0.
#
# Exit status: 0 if nothing stops the update, 1 if something does, 2 on a
# usage error.

set -euo pipefail

here=$(cd "$(dirname "$0")" && pwd)
cd "$here/.."

fetch=1
base=""
target=""
while [ $# -gt 0 ]; do
	case "$1" in
	--no-fetch) fetch=0 ;;
	--base) base="$2"; shift ;;
	-h | --help) sed -n '2,20p' "$0"; exit 0 ;;
	-*) echo "Unknown option: $1" >&2; exit 2 ;;
	*) target="$1" ;;
	esac
	shift
done

if [ -z "$base" ]; then
	branch=$(git rev-parse --abbrev-ref HEAD)
	case "$branch" in
	usl/v*) base="${branch#usl/}" ;;
	*) echo "Not on a usl/v<version> branch. Use --base <tag>." >&2; exit 2 ;;
	esac
fi

if [ "$fetch" = 1 ]; then
	echo "Fetch the upstream tags and release branches ..."
	git fetch --quiet --no-recurse-submodules --tags upstream \
		'refs/heads/release-*:refs/remotes/upstream/release-*'
fi

git rev-parse --verify --quiet "$base^{commit}" >/dev/null ||
	{ echo "The base tag $base does not exist." >&2; exit 2; }

# v12.0.0 -> 12.0
minor=$(echo "${base#v}" | cut -d. -f1-2)
release_branch="upstream/release-$minor"

# The tags after the base, newest first. sort -V puts v12.1.0-rc1 after
# v12.1.0, so the final tags and the release candidates are separate lists.
newer_tags() {
	git tag -l 'v[0-9]*' | grep -E -e "$1" | grep -v -- '-usl\.' |
		while read -r t; do
			if [ "$(printf '%s\n%s\n' "$base" "$t" | sort -V | tail -1)" = "$t" ] &&
				[ "$t" != "$base" ] && [ "${t#"$base"-}" = "$t" ]; then
				echo "$t"
			fi
		done | sort -rV
}
finals=$(newer_tags '^v[0-9]+\.[0-9]+\.[0-9]+$')
rcs=$(newer_tags '-rc[0-9]+$')

echo
echo "== Upstream after $base"
echo "Final releases: ${finals:-none}" | tr '\n' ' '; echo
echo "Release candidates: ${rcs:-none}" | tr '\n' ' '; echo
if git rev-parse --verify --quiet "$release_branch" >/dev/null; then
	n=$(git rev-list --count "$base..$release_branch")
	echo "Commits on $release_branch after $base, not released yet: $n"
	git log --format='  %h %s' "$base..$release_branch" | head -20
fi

if [ -z "$target" ]; then
	target=$(echo "$finals" | head -1)
	[ -n "$target" ] || target="$release_branch"
fi
git rev-parse --verify --quiet "$target^{commit}" >/dev/null ||
	{ echo "The target $target does not exist." >&2; exit 2; }

echo
echo "== Compare $base with $target"
stops=0
stop() { echo "STOP: $*"; stops=$((stops + 1)); }

echo
echo "-- Watched files (usl/watched-files.txt)"
changed=0
while read -r level path; do
	case "$level" in '' | \#*) continue ;; esac
	if ! git diff --quiet "$base" "$target" -- "$path"; then
		changed=1
		stat=$(git diff --shortstat "$base" "$target" -- "$path")
		if [ "$level" = stop ]; then
			stop "$path changed:$stat"
		else
			echo "REVIEW: $path changed:$stat"
		fi
	fi
done <"$here/watched-files.txt"
[ "$changed" = 1 ] || echo "No change."

added() { git diff "$base" "$target" -- "$@" | grep -E '^\+[^+]' || true; }

echo
echo "-- Feature flags (server/public/model/feature_flags.go)"
flags=$(git diff "$base" "$target" -- server/public/model/feature_flags.go |
	grep -E '^[+-]\s*f\.[A-Za-z0-9]+ = ' || true)
if [ -n "$flags" ]; then
	echo "$flags"
	echo "$flags" | grep -E '^\+\s*f\.[A-Za-z0-9]+ = true' | while read -r l; do
		echo "STOP: a feature flag is on by default: ${l#+}"
	done || true
	n=$(echo "$flags" | grep -cE '^\+\s*f\.[A-Za-z0-9]+ = true' || true)
	stops=$((stops + n))
else
	echo "No change."
fi

echo
echo "-- New defaults that are true (server/public/model/config.go)"
defaults=$(added server/public/model/config.go |
	grep -E '= (new|model\.NewPointer|NewPointer)\(true\)' || true)
if [ -n "$defaults" ]; then
	echo "$defaults" | sed 's/^+[[:space:]]*/REVIEW: /'
else
	echo "No change."
fi

echo
echo "-- Prepackaged plugins (server/Makefile)"
plugins=$(git diff "$base" "$target" -- server/Makefile |
	grep -E '^[+-]PLUGIN_PACKAGES' || true)
echo "${plugins:-No change.}"

echo
echo "-- Database migrations (one way: you cannot go back without a backup)"
migrations=$(git diff --name-only --diff-filter=A "$base" "$target" -- \
	'server/channels/db/migrations/postgres/*.up.sql' || true)
echo "${migrations:-None.}"

echo
echo "-- Changed GitHub workflows (the fork runs the active ones)"
workflows=$(git diff --name-status "$base" "$target" -- .github/workflows || true)
echo "${workflows:-No change.}"
if echo "$workflows" | grep -q 'usl-'; then
	stop "upstream changed a file with a usl- name in .github/workflows"
fi

echo
echo "-- Denied features (usl/denylist.txt)"
denied=0
watch_added=$(added server/public/model/feature_flags.go \
	server/public/model/config.go server/Makefile)
while IFS= read -r line; do
	pattern=$(echo "$line" | sed -E 's/[[:space:]]+#.*$//')
	case "$pattern" in '' | \#*) continue ;; esac
	hit=$(echo "$watch_added" | grep -E -- "$pattern" || true)
	if [ -n "$hit" ]; then
		denied=1
		stop "denied pattern '$pattern' matches: $(echo "$hit" | head -1 | sed 's/^+[[:space:]]*//')"
	fi
done <"$here/denylist.txt"
[ "$denied" = 1 ] || echo "No match."

echo
echo "-- Conflict test: the USL commits ($base..HEAD) on $target, in memory only"
if git merge-tree --write-tree --name-only --merge-base="$base" "$target" HEAD >/dev/null 2>&1; then
	echo "No conflict."
else
	stop "the USL commits conflict with $target:"
	git merge-tree --write-tree --name-only --merge-base="$base" "$target" HEAD 2>/dev/null |
		sed -n '2,/^$/p' | sed 's/^/  /' || true
fi

echo
if [ "$stops" -gt 0 ]; then
	echo "== Result: $stops STOP item(s). Read usl/update.md, step 2, before you update to $target."
	exit 1
fi
echo "== Result: nothing stops the update to $target. Read the REVIEW items."
