#!/bin/bash
fail=0
for file in "$@"; do
  if grep -q "argocd.argoproj.io/sync-wave" "$file"; then
    wave=$(grep "argocd.argoproj.io/sync-wave" "$file" | awk -F': ' '{print $2}' | tr -d "'\"")
    if ! [[ "$wave" =~ ^-?[0-9]+$ ]]; then
      echo "❌ $file: Invalid sync-wave: $wave"
      fail=1
    fi
  fi
done
exit $fail
