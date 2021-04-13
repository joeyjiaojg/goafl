export GOBIN=$(shell pwd -P)/bin
TARGETOS ?= linux
TARGETARCH ?= amd64
GO ?= go

executor:
	GOOS=$(TARGETOS) GOARCH=$(TARGETARCH) $(GO) build $(GOTARGETFLAGS) -o ./bin/$(TARGETOS)_$(TARGETARCH)/executor github.com/joeyjiaojg/goafl/src/executor
