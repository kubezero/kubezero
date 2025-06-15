#!/bin/bash
for file in "$@"; do
  kube-score score "$file"
done
