.PHONY: help build run test lint docker-build kind-up kind-down kind-load \
        deploy argocd-install observability-install tf-init tf-apply \
        fault-errors fault-cpu clean

IMAGE        ?= ghcr.io/han-sang-min/deploy-pilot
TAG          ?= dev
CLUSTER      ?= deploy-pilot
VERSION      ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT       ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo none)

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
	  awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-22s\033[0m %s\n",$$1,$$2}'

## --- Week 1: build + CI ---
build: ## Build the server binary
	go build -trimpath -ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT)" -o bin/server ./cmd/server

run: build ## Run locally on :8080
	./bin/server

test: ## Run unit tests
	go test -race ./...

lint: ## gofmt + go vet
	gofmt -l .
	go vet ./...

docker-build: ## Build the container image
	docker build -f deploy/docker/Dockerfile \
	  --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) \
	  -t $(IMAGE):$(TAG) .

## --- Week 2: kind + GitOps ---
kind-up: ## Create local kind cluster
	kind create cluster --name $(CLUSTER) --config kind/kind-cluster.yaml

kind-down: ## Delete local kind cluster
	kind delete cluster --name $(CLUSTER)

kind-load: docker-build ## Load the image into kind
	kind load docker-image $(IMAGE):$(TAG) --name $(CLUSTER)

argocd-install: ## Install Argo CD into the cluster
	kubectl create namespace argocd --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

deploy: ## Register the Argo CD project + application (GitOps)
	kubectl apply -f gitops/argocd/project.yaml
	kubectl apply -f gitops/argocd/application.yaml

## --- Week 3: observability ---
observability-install: ## Install kube-prometheus-stack via Helm
	helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
	helm upgrade --install kube-prom-stack prometheus-community/kube-prometheus-stack \
	  --namespace monitoring --create-namespace

## --- Week 4: IaC + failure scenarios ---
tf-init: ## terraform init
	cd infra/terraform && terraform init

tf-apply: ## terraform apply (cluster add-ons)
	cd infra/terraform && terraform apply -var-file=dev.tfvars

fault-errors: ## Drive the high-error-rate scenario
	./tools/faultinject/faultinject.sh errors

fault-cpu: ## Drive the cpu-spike scenario
	./tools/faultinject/faultinject.sh cpu

clean: ## Remove build artifacts
	rm -rf bin
