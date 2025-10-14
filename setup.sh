#!/bin/bash
set -e

echo "========================================="
echo "K8s Deployment Exporter Setup"
echo "========================================="

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check prerequisites
echo -e "\n${YELLOW}Checking prerequisites...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}Error: Go is not installed${NC}"
    echo "Please install Go 1.21 or later from https://golang.org/dl/"
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    echo "Please install Docker from https://docs.docker.com/get-docker/"
    exit 1
fi

if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}Error: kubectl is not installed${NC}"
    echo "Please install kubectl from https://kubernetes.io/docs/tasks/tools/"
    exit 1
fi

echo -e "${GREEN}✓ All prerequisites found${NC}"

# Change to exporter directory
cd "$(dirname "$0")"

# Initialize Go module
echo -e "\n${YELLOW}Downloading Go dependencies...${NC}"
go mod download
echo -e "${GREEN}✓ Dependencies downloaded${NC}"

# Build the binary
echo -e "\n${YELLOW}Building the exporter binary...${NC}"
make build
echo -e "${GREEN}✓ Binary built successfully${NC}"

# Build Docker image
echo -e "\n${YELLOW}Building Docker image...${NC}"
make docker-build
echo -e "${GREEN}✓ Docker image built successfully${NC}"

# Check Kubernetes connection
echo -e "\n${YELLOW}Checking Kubernetes connection...${NC}"
if kubectl cluster-info &> /dev/null; then
    echo -e "${GREEN}✓ Connected to Kubernetes cluster${NC}"
    kubectl cluster-info | head -n 1
else
    echo -e "${RED}Error: Cannot connect to Kubernetes cluster${NC}"
    echo "Please configure your kubeconfig or start a local cluster (minikube, kind, etc.)"
    exit 1
fi

# Prompt for deployment
echo -e "\n${YELLOW}Ready to deploy to Kubernetes${NC}"
read -p "Do you want to deploy now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo -e "\n${YELLOW}Deploying to Kubernetes...${NC}"
    make deploy
    
    echo -e "\n${GREEN}✓ Deployment complete!${NC}"
    echo -e "\n${YELLOW}Waiting for pod to be ready...${NC}"
    kubectl wait --for=condition=ready pod -l app=k8s-deployment-exporter -n monitoring --timeout=60s
    
    echo -e "\n${GREEN}=========================================${NC}"
    echo -e "${GREEN}Setup Complete!${NC}"
    echo -e "${GREEN}=========================================${NC}"
    
    echo -e "\nTo view metrics, run:"
    echo -e "  ${YELLOW}kubectl port-forward -n monitoring svc/k8s-deployment-exporter 9101:9101${NC}"
    echo -e "  ${YELLOW}curl http://localhost:9101/metrics${NC}"
    
    echo -e "\nTo view logs:"
    echo -e "  ${YELLOW}kubectl logs -n monitoring -l app=k8s-deployment-exporter -f${NC}"
    
    echo -e "\nTo test with a sample deployment:"
    echo -e "  ${YELLOW}kubectl create deployment test-app --image=nginx --replicas=3${NC}"
    echo -e "  ${YELLOW}kubectl scale deployment test-app --replicas=0  # Trigger downtime${NC}"
    echo -e "  ${YELLOW}kubectl scale deployment test-app --replicas=3  # Trigger recovery${NC}"
else
    echo -e "\n${YELLOW}Skipping deployment${NC}"
    echo -e "\nYou can deploy later with:"
    echo -e "  ${YELLOW}make deploy${NC}"
fi

echo -e "\n${GREEN}For more information, see README.md${NC}"
