#!/bin/bash

kubeconform \
  -summary \
  -strict \
  -ignore-missing-schemas \
  -skip "CustomResourceDefinition" \
  -ignore-filename-pattern "^\..+" \
  "$@"
