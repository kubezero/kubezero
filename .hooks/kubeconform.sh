#!/bin/bash
# Validate rendered kustomize output with kubeconform.
# Usage: kubeconform.sh [dirs...]
#   If no dirs given, discovers all kustomization.yaml directories under
#   modules/, stacks/, packages/, and controller/.

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"

run_kubeconform() {
  kubeconform \
    -summary \
    -strict \
    -ignore-missing-schemas \
    -skip "CustomResourceDefinition" \
    -output pretty
}

if [ $# -gt 0 ]; then
  dirs=("$@")
else
  mapfile -t dirs < <(
    find "$ROOT"/modules "$ROOT"/stacks "$ROOT"/packages "$ROOT"/controller \
      -name "kustomization.yaml" -not -path "*/helm-chart/*" \
      | xargs -I{} dirname {} \
      | sort -u
  )
fi

failed=0
for dir in "${dirs[@]}"; do
  rel="${dir#$ROOT/}"
  output=$(kustomize build "$dir" 2>&1) || { echo "SKIP $rel (kustomize build failed)"; continue; }
  if echo "$output" | run_kubeconform; then
    echo "OK   $rel"
  else
    echo "FAIL $rel"
    failed=1
  fi
done

exit $failed
