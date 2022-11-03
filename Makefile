# Set the shell to bash always
SHELL := /bin/bash

# Look for a .env file, and if present, set make variables from it.
ifneq (,$(wildcard ./.env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

KIND_CLUSTER_NAME ?= local-dev
KUBECONFIG ?= $(HOME)/.kube/config

VERSION := $(shell git describe --dirty --always --tags | sed 's/-/./2' | sed 's/-/./2')
ifndef VERSION
VERSION := 0.0.0
endif

# Tools
KIND=$(shell which kind)
LINT=$(shell which golangci-lint)
KUBECTL=$(shell which kubectl)
SED=$(shell which sed)

.DEFAULT_GOAL := help

.PHONY: help
## help: Print this help
help: Makefile
	@echo
	@echo " Choose a command run in "$(PROJECT_NAME)":"
	@echo
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo


.PHONY: dev
## dev: Run the controller in debug mode
dev: generate
	$(KUBECTL) apply -f package/crds/ -R
	go run cmd/main.go -d

.PHONY: generate
## generate: Generate all CRDs
generate: tidy
	go generate ./...

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	$(LINT) run

.PHONY: kind.up
## kind.up: Starts a KinD cluster for local development
kind.up:
	@$(KIND) get kubeconfig --name $(KIND_CLUSTER_NAME) >/dev/null 2>&1 || $(KIND) create cluster --name=$(KIND_CLUSTER_NAME)

.PHONY: kind.down
## kind.down: Shuts down the KinD cluster
kind.down:
	@$(KIND) delete cluster --name=$(KIND_CLUSTER_NAME)


## install.crossplane: Install Crossplane into the local KinD cluster
install.crossplane:
	$(KUBECTL) create namespace crossplane-system || true
	helm repo add crossplane-stable https://charts.crossplane.io/stable
	helm repo update
	helm install crossplane --namespace crossplane-system crossplane-stable/crossplane


## install.argocd: Install ArgoCD into the local KinD cluster
install.argocd:
	$(KUBECTL) create namespace argo-system || true
	$(KUBECTL) apply -n argo-system -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

