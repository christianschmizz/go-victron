
CGO_ENABLED ?= 0
GOOS ?= darwin
GOARCH ?= amd64

.PHONY: build-static-logger
build-static-logger:
	@echo Building MQTT logger...
	@go build -v -a -ldflags '-extldflags "-static"' -o mqtt-logger cmd/mqtt-logger/main.go

.PHONY: build-static-check
build-static-check:
	@echo Building VRMcheck...
	@go build -v -a -ldflags '-extldflags "-static"' -o vrmcheck cmd/vrmcheck/main.go

.PHONY: build
build: build-static-logger build-static-check
