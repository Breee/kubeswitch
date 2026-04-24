#!/usr/bin/env bash
# End-to-end test for kubeswitch using two kind clusters.
# Prerequisites: kind, kubectl, go
# Usage: ./e2e_test.sh
set -euo pipefail

CLUSTER_A="kubeswitch-e2e-a"
CLUSTER_B="kubeswitch-e2e-b"
BINARY="./kubeswitch"
KUBECONFIG_FILE="$(mktemp)"
export KUBECONFIG="$KUBECONFIG_FILE"

cleanup() {
  echo "--- Cleaning up ---"
  kind delete cluster --name "$CLUSTER_A" 2>/dev/null || true
  kind delete cluster --name "$CLUSTER_B" 2>/dev/null || true
  rm -f "$KUBECONFIG_FILE" "$BINARY"
}
trap cleanup EXIT

echo "=== Building kubeswitch ==="
CGO_ENABLED=0 go build -ldflags="-s -w" -o "$BINARY"

echo "=== Creating kind clusters ==="
kind create cluster --name "$CLUSTER_A" --wait 60s
kind create cluster --name "$CLUSTER_B" --wait 60s

echo "=== Creating test namespaces ==="
kubectl --context "kind-${CLUSTER_A}" create namespace test-ns-a
kubectl --context "kind-${CLUSTER_B}" create namespace test-ns-b

echo "=== Test 1: Quick-switch namespace in current context ==="
kubectl config use-context "kind-${CLUSTER_A}"
$BINARY test-ns-a
CURRENT_NS=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.namespace}')
if [[ "$CURRENT_NS" != "test-ns-a" ]]; then
  echo "FAIL: expected namespace test-ns-a, got $CURRENT_NS"
  exit 1
fi
echo "PASS: switched to test-ns-a in current context"

echo "=== Test 2: Quick-switch context + namespace ==="
$BINARY "kind-${CLUSTER_B}" test-ns-b
CURRENT_CTX=$(kubectl config current-context)
CURRENT_NS=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.namespace}')
if [[ "$CURRENT_CTX" != "kind-${CLUSTER_B}" ]]; then
  echo "FAIL: expected context kind-${CLUSTER_B}, got $CURRENT_CTX"
  exit 1
fi
if [[ "$CURRENT_NS" != "test-ns-b" ]]; then
  echo "FAIL: expected namespace test-ns-b, got $CURRENT_NS"
  exit 1
fi
echo "PASS: switched to kind-${CLUSTER_B}/test-ns-b"

echo "=== Test 3: Quick-switch context with default namespace (.) ==="
$BINARY "kind-${CLUSTER_A}" .
CURRENT_CTX=$(kubectl config current-context)
CURRENT_NS=$(kubectl config view --minify -o jsonpath='{.contexts[0].context.namespace}')
if [[ "$CURRENT_CTX" != "kind-${CLUSTER_A}" ]]; then
  echo "FAIL: expected context kind-${CLUSTER_A}, got $CURRENT_CTX"
  exit 1
fi
if [[ "$CURRENT_NS" != "default" ]]; then
  echo "FAIL: expected namespace default, got $CURRENT_NS"
  exit 1
fi
echo "PASS: switched to kind-${CLUSTER_A}/default"

echo "=== Test 4: Fail on non-existent namespace ==="
if $BINARY nonexistent-ns 2>/dev/null; then
  echo "FAIL: expected error for non-existent namespace"
  exit 1
fi
echo "PASS: correctly rejected non-existent namespace"

echo ""
echo "=== All e2e tests passed ==="
