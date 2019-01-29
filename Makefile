GO      ?= GO111MODULE=on go
REPO    ?= contrib.go.opencensus.io/kubernetes-operator
IMAGE   ?= opencensus-operator
GIT_REV := $(shell git log -n1 --pretty='%h')
VERSION ?= $(GIT_REV)

build:
	$(GO) build $(REPO)/...

container:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build $(REPO)/...
	docker build -t $(IMAGE):$(VERSION) .
