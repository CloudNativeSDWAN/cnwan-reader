# Image URL to use all building/pushing image targets
IMG ?= <repository>

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Update gomod
update-gomod:
	go mod download
	go mod tidy 
	go mod verify

# Build this
build: test
	go build -a -o cnwan-reader *.go

test:
	go test ./...

# Build the docker image
docker-build: test
	docker build . -t ${IMG}

# Push the docker image
docker-push:
	docker push ${IMG}