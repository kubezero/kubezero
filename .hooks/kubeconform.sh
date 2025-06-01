#!/bin/bash

for file in "$@"; do
  # skip CRDs if needed
  if grep -q "kind: CustomResourceDefinition" "$file"; then
    echo "🔄 Skipping CRD: $file"
    continue
  fi

  kubeconform -strict -summary -ignore-missing-schemas "$file"
done
