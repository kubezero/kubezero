#!/usr/bin/env python3
import sys, yaml, os

def validate(file_path):
    try:
        with open(file_path, "r") as f:
            data = yaml.safe_load(f)
        if not isinstance(data, dict):
            print(f"{file_path}: top-level YAML is not a dict")
            return 1
        required_keys = ["replicaCount", "image", "resources"]
        for key in required_keys:
            if key not in data:
                print(f"{file_path}: missing key '{key}'")
        return 0
    except Exception as e:
        print(f"{file_path}: failed to parse YAML - {e}")
        return 1

sys.exit(any(validate(path) for path in sys.argv[1:]))
