.PHONY: help bootstrap teardown validate registry-add registry-remove scaffold cli-build cli-install

BOOTSTRAP_CONFIG ?= bootstrap/k3d-bootstrap-cluster.yaml
CLUSTER_NAME     ?= kubezero

help: ## Show this help
	@awk 'BEGIN {FS = ":.*##"} /^[a-zA-Z_-]+:.*##/ { printf "  %-22s %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ── Bootstrap ──────────────────────────────────────────────────────────────────

bootstrap: ## Create the local k3d bootstrap cluster
	k3d cluster create --config $(BOOTSTRAP_CONFIG)

teardown: ## Delete the local k3d bootstrap cluster
	k3d cluster delete $(CLUSTER_NAME)

# ── Registry management ────────────────────────────────────────────────────────

registry-add: ## Copy a package into registry/ (usage: make registry-add PACKAGE=aws-management)
ifndef PACKAGE
	$(error PACKAGE is required. Example: make registry-add PACKAGE=aws-management)
endif
	@if [ -d registry/$(PACKAGE) ]; then \
	  echo "registry/$(PACKAGE) already exists — skipping"; \
	else \
	  cp -r packages/$(PACKAGE) registry/$(PACKAGE); \
	  echo "Added registry/$(PACKAGE)"; \
	fi

registry-remove: ## Remove a package from registry/ (usage: make registry-remove PACKAGE=aws-management)
ifndef PACKAGE
	$(error PACKAGE is required. Example: make registry-remove PACKAGE=aws-management)
endif
	@if [ ! -d registry/$(PACKAGE) ]; then \
	  echo "registry/$(PACKAGE) does not exist — nothing to remove"; \
	else \
	  rm -rf registry/$(PACKAGE); \
	  echo "Removed registry/$(PACKAGE)"; \
	fi

registry-list: ## List all packages and their activation status in registry/
	@echo "Available packages (packages/):"
	@ls packages/ | sed 's/^/  /'
	@echo ""
	@echo "Active registry entries (registry/):"
	@for d in registry/*/; do \
	  name=$$(basename "$$d"); \
	  if [ -f "$$d/gitops.yaml" ]; then \
	    echo "  $$name  (active)"; \
	  elif [ -d "$$d" ]; then \
	    echo "  $$name  (template)"; \
	  fi; \
	done || echo "  none"

# ── Scaffold ───────────────────────────────────────────────────────────────────

scaffold: ## Scaffold a new package (usage: make scaffold CLOUD=aws ROLE=worker)
ifndef CLOUD
	$(error CLOUD is required. Example: make scaffold CLOUD=aws ROLE=worker)
endif
ifndef ROLE
	$(error ROLE is required. Example: make scaffold CLOUD=aws ROLE=worker)
endif
	cd cli && go run main.go scaffold --cloud $(CLOUD) --role $(ROLE) --repo ..

# ── Validation ─────────────────────────────────────────────────────────────────

validate: ## Run kubeconform + kube-score on all kustomize builds
	chmod +x .hooks/kubeconform.sh .hooks/kube-score.sh
	.hooks/kubeconform.sh
	.hooks/kube-score.sh

validate-kubeconform: ## Run kubeconform only
	chmod +x .hooks/kubeconform.sh
	.hooks/kubeconform.sh

validate-kube-score: ## Run kube-score only
	chmod +x .hooks/kube-score.sh
	.hooks/kube-score.sh

# ── CLI ────────────────────────────────────────────────────────────────────────

cli-build: ## Build the kubezero CLI binary
	cd cli && go build -o ../bin/kubezero .

cli-install: ## Install the kubezero CLI to /usr/local/bin
	cd cli && go build -o /usr/local/bin/kubezero .

cli-vet: ## Run go vet on the CLI
	cd cli && go vet ./...
