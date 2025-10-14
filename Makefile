.PHONY: build docker-build docker-push deploy clean test run

# Variables
IMAGE_NAME ?= k8s-deployment-exporter
IMAGE_TAG ?= latest
REGISTRY ?= docker.io/yourorg

# Build binary
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o k8s-deployment-exporter .

# Build Docker image
docker-build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# Push Docker image
docker-push:
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)
	docker push $(REGISTRY)/$(IMAGE_NAME):$(IMAGE_TAG)

# Deploy to Kubernetes
deploy:
	kubectl apply -f deployment.yaml

# Remove from Kubernetes
undeploy:
	kubectl delete -f deployment.yaml

# Run locally
run:
	go run main.go --kubeconfig ~/.kube/config

# Run with specific namespace
run-namespace:
	go run main.go --kubeconfig ~/.kube/config --namespace=default

# Clean build artifacts
clean:
	rm -f k8s-deployment-exporter
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Test build
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code (requires golangci-lint)
lint:
	golangci-lint run

# View metrics locally
metrics:
	@echo "Port-forwarding to view metrics..."
	kubectl port-forward -n monitoring svc/k8s-deployment-exporter 9101:9101

# View logs
logs:
	kubectl logs -n monitoring -l app=k8s-deployment-exporter -f

# Complete build and deploy
all: build docker-build deploy

# Help
help:
	@echo "Available targets:"
	@echo "  build          - Build Go binary"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-push    - Push Docker image to registry"
	@echo "  deploy         - Deploy to Kubernetes"
	@echo "  undeploy       - Remove from Kubernetes"
	@echo "  run            - Run locally with kubeconfig"
	@echo "  run-namespace  - Run locally monitoring specific namespace"
	@echo "  clean          - Clean build artifacts"
	@echo "  deps           - Install dependencies"
	@echo "  test           - Run tests"
	@echo "  fmt            - Format code"
	@echo "  lint           - Lint code"
	@echo "  metrics        - Port-forward and view metrics"
	@echo "  logs           - View exporter logs"
	@echo "  all            - Build and deploy everything"
