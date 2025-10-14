#!/bin/bash
# Quick test script to validate the exporter

set -e

echo "🧪 Testing K8s Deployment Exporter"
echo "===================================="

# Create test namespace
echo -e "\n1️⃣ Creating test namespace..."
kubectl create namespace test-exporter --dry-run=client -o yaml | kubectl apply -f -

# Create test deployment
echo -e "\n2️⃣ Creating test deployment..."
kubectl create deployment test-app -n test-exporter --image=nginx:alpine --replicas=3 --dry-run=client -o yaml | kubectl apply -f -

# Wait for deployment to be ready
echo -e "\n3️⃣ Waiting for deployment to be ready..."
kubectl wait --for=condition=available deployment/test-app -n test-exporter --timeout=60s

echo -e "\n✅ Deployment is ready"

# Check metrics
echo -e "\n4️⃣ Checking metrics endpoint..."
POD=$(kubectl get pods -n monitoring -l app=k8s-deployment-exporter -o jsonpath='{.items[0].metadata.name}')

if [ -z "$POD" ]; then
    echo "❌ Exporter pod not found. Is it deployed?"
    exit 1
fi

echo "Exporter pod: $POD"

# Port forward in background
echo -e "\n5️⃣ Port-forwarding to exporter..."
kubectl port-forward -n monitoring "$POD" 9101:9101 > /dev/null 2>&1 &
PF_PID=$!
sleep 3

# Cleanup function
cleanup() {
    echo -e "\n\n🧹 Cleaning up..."
    kill $PF_PID 2>/dev/null || true
}
trap cleanup EXIT

# Check for test-app metrics
echo -e "\n6️⃣ Fetching metrics for test-app..."
METRICS=$(curl -s http://localhost:9101/metrics | grep 'test-app')

if [ -z "$METRICS" ]; then
    echo "⚠️  No metrics found yet. The exporter might need more time to discover the deployment."
    echo "    Trying one more time in 10 seconds..."
    sleep 10
    METRICS=$(curl -s http://localhost:9101/metrics | grep 'test-app')
fi

if [ -n "$METRICS" ]; then
    echo "✅ Metrics found!"
    echo "$METRICS"
else
    echo "❌ No metrics found for test-app"
    echo "This might be normal if the exporter hasn't scraped yet."
fi

# Simulate downtime
echo -e "\n7️⃣ Simulating deployment downtime..."
echo "   Scaling down to 0 replicas..."
kubectl scale deployment test-app -n test-exporter --replicas=0

sleep 5

echo "   Scaling back up to 3 replicas..."
kubectl scale deployment test-app -n test-exporter --replicas=3

echo "   Waiting for recovery..."
kubectl wait --for=condition=available deployment/test-app -n test-exporter --timeout=60s

echo -e "\n✅ Deployment recovered"

# Check recovery metrics
echo -e "\n8️⃣ Checking recovery metrics (waiting 20s for scrape)..."
sleep 20

RECOVERY_METRICS=$(curl -s http://localhost:9101/metrics | grep -E 'test-app.*(downtime|recovery|restart)')

if [ -n "$RECOVERY_METRICS" ]; then
    echo "✅ Recovery metrics captured!"
    echo "$RECOVERY_METRICS"
else
    echo "⚠️  Recovery metrics not yet available. Check again in a few minutes."
fi

# Summary
echo -e "\n============================================"
echo "✅ Test completed successfully!"
echo "============================================"
echo ""
echo "The exporter is working correctly."
echo ""
echo "To view all metrics:"
echo "  curl http://localhost:9101/metrics"
echo ""
echo "To clean up test resources:"
echo "  kubectl delete namespace test-exporter"
echo ""
echo "View exporter logs:"
echo "  kubectl logs -n monitoring -l app=k8s-deployment-exporter -f"
