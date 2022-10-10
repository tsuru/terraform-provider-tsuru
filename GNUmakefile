HOSTNAME=registry.terraform.io
NAMESPACE=tsuru
NAME=tsuru
BINARY=terraform-provider-${NAME}
VERSION=2.1.6

UNAME_S := $(shell uname -s)
UNAME_P := $(shell uname -p)
ifeq ($(UNAME_S),Linux)
	OS := linux
	UNAME_P := $(shell uname -m)
endif
ifeq ($(UNAME_S),Darwin)
	OS := darwin
	UNAME_P := $(shell uname -m)
endif

ifeq ($(UNAME_P),x86_64)
	ARCH := amd64
endif

ifneq ($(filter %86,$(UNAME_P)),)
	ARCH := 386
endif
ifneq ($(filter arm%,$(UNAME_P)),)
	ARCH := arm
endif
ifeq ($(UNAME_P),arm64)
	ARCH := arm64
endif

OS_ARCH=${OS}_${ARCH}

default: install

build:
	go build -o ${BINARY}

release:
	GOOS=darwin GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_darwin_amd64
	GOOS=freebsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_freebsd_386
	GOOS=freebsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_freebsd_amd64
	GOOS=freebsd GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_freebsd_arm
	GOOS=linux GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_linux_386
	GOOS=linux GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_linux_amd64
	GOOS=linux GOARCH=arm go build -o ./bin/${BINARY}_${VERSION}_linux_arm
	GOOS=openbsd GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_openbsd_386
	GOOS=openbsd GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_openbsd_amd64
	GOOS=solaris GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_solaris_amd64
	GOOS=windows GOARCH=386 go build -o ./bin/${BINARY}_${VERSION}_windows_386
	GOOS=windows GOARCH=amd64 go build -o ./bin/${BINARY}_${VERSION}_windows_amd64

install: build
	mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}

uninstall:
	rm -Rf ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}

lint:
	curl -sfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.39.0
	time golangci-lint run

test:
	TF_ACC=1 TF_ACC_TERRAFORM_VERSION=1.3.2 go test ./... -v

generate-docs:
	go install github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.13.0
	go generate
