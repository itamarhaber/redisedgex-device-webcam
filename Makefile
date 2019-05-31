.PHONY: build test clean docker

GO=CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=cmd/device-webcam/device-webcam
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/redislabs/edgex-device-webcam.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) install -tags=safe

cmd/device-webcam/device-webcam:
	$(GO) build $(GOFLAGS) -o $@ ./cmd/device-webcam

docker:
	docker build \
		-f cmd/device-webcam/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t redislabs/edgex-device-webcam:$(GIT_SHA) \
		-t redislabs/edgex-device-webcam:$(VERSION)-dev \
		.

test:
	$(GO) vet ./...
	gofmt -l .
	$(GO) test -coverprofile=coverage.out ./...

clean:
	rm -f $(MICROSERVICES)
