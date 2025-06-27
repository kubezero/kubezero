#!/bin/bash

# KubeZero CLI Demo Script
# This script demonstrates the basic usage of the KubeZero CLI

set -e

echo "🚀 KubeZero CLI Demo"
echo "==================="
echo

# Build the CLI if it doesn't exist
if [ ! -f "./kubezero" ]; then
    echo "📦 Building KubeZero CLI..."
    make build
    echo
fi

# Show version
echo "📋 CLI Version:"
./kubezero version
echo

# Show help
echo "💡 Available Commands:"
./kubezero --help
echo

# Check if k3d is available
echo "🔍 Checking Prerequisites:"
if command -v k3d >/dev/null 2>&1; then
    echo "✅ k3d is installed: $(k3d version | head -1)"
else
    echo "❌ k3d is not installed"
    echo "   Install k3d from: https://k3d.io/"
    exit 1
fi

if command -v kubectl >/dev/null 2>&1; then
    echo "✅ kubectl is installed: $(kubectl version --client --short 2>/dev/null || kubectl version --client)"
else
    echo "❌ kubectl is not installed"
    echo "   Install kubectl from: https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi
echo

# Check configuration file
echo "📄 Configuration File:"
if [ -f "../bootstrap/k3d-bootstrap-cluster.yaml" ]; then
    echo "✅ Found: ../bootstrap/k3d-bootstrap-cluster.yaml"
    echo "   Cluster name: $(grep 'name:' ../bootstrap/k3d-bootstrap-cluster.yaml | awk '{print $2}')"
else
    echo "❌ Configuration file not found"
    echo "   Expected: ../bootstrap/k3d-bootstrap-cluster.yaml"
    exit 1
fi
echo

echo "🎯 Ready to bootstrap!"
echo "   Run: ./kubezero bootstrap"
echo "   Or:  make run-dev"
echo
echo "📊 After bootstrap, check status:"
echo "   Run: ./kubezero status"
echo
echo "🧹 To cleanup when done:"
echo "   Run: ./kubezero cleanup"
