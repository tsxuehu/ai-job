#!/usr/bin/env bash

set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$repo_root"

broken=0

while IFS= read -r -d '' file; do
  file_dir="$(dirname "$file")"

  while IFS= read -r target; do
    [[ -z "$target" ]] && continue
    [[ "$target" == http://* ]] && continue
    [[ "$target" == https://* ]] && continue
    [[ "$target" == mailto:* ]] && continue
    [[ "$target" == \#* ]] && continue

    target="${target%%#*}"
    target="${target%%\?*}"
    target="${target#<}"
    target="${target%>}"

    if [[ "$target" == /* ]]; then
      resolved=".$target"
    else
      resolved="$file_dir/$target"
    fi

    if [[ ! -e "$resolved" ]]; then
      printf 'Broken link: %s -> %s\n' "$file" "$target"
      broken=1
    fi
  done < <(perl -ne '
    if (/^\s*```/) {
      $in_fence = !$in_fence;
      next;
    }
    next if $in_fence;
    s/`[^`]*`//g;
    while (/\[[^]]*\]\(([^)]+)\)/g) {
      print "$1\n";
    }
  ' "$file")
done < <(find . -type f -name '*.md' -not -path './.git/*' -print0)

if [[ "$broken" -ne 0 ]]; then
  exit 1
fi

printf 'All local Markdown links are valid.\n'
