NAME=gaia
GO_LDFLAGS_STATIC="-s -w -extldflags -static"
NAMESPACE=${NAME}
RELEASE_NAME=${NAME}
HELM_DIR=$(shell pwd)/helm
TEST=$$(go list ./... | grep -v /vendor/ | grep /testacc)
TEST_TIMEOUT_ACC?=20m
TEST_TIMEOUT?=50s
# Set the build dir, where built cross-compiled binaries will be output
BUILDDIR := bin

default: dev

dev:
	go run ./cmd/gaia/main.go -home-path=${PWD}/tmp -dev=true

compile_frontend:
	cd ./frontend && \
	rm -rf dist && \
	npm install && \
	npm run build

static_assets:
	go get github.com/GeertJohan/go.rice && \
	go get github.com/GeertJohan/go.rice/rice && \
	cd ./handlers && \
	rm -f rice-box.go && \
	rice embed-go && \
	cd ../helper/assethelper && \
	rm -f rice-box.go && \
	rice embed-go

compile_backend:
	env GOOS=linux GOARCH=amd64 go build -ldflags $(GO_LDFLAGS_STATIC) -o $(NAME)-linux-amd64 ./cmd/gaia/main.go

binaries:
	CGO_ENABLED=0 gox \
		-osarch="linux/amd64 linux/arm darwin/amd64 windows/amd64" \
		-ldflags=${GO_LDFLAGS_STATIC} \
		-output="$(BUILDDIR)/{{.OS}}/{{.Arch}}/$(NAME)" \
		-tags="netgo" \
		./cmd/gaia/.

download:
	go mod download

get:
	go get ./...

test:
	go test -v -race -timeout=$(TEST_TIMEOUT) ./...

test-cover:
	go test -v -timeout=$(TEST_TIMEOUT) ./... --coverprofile=cover.out

test-acc:
	GAIA_RUN_ACC=true GAIA_DEV=true go test -v $(TEST) -timeout=$(TEST_TIMEOUT_ACC)

release: compile_frontend static_assets compile_backend

deploy-kube:
	helm upgrade --install ${RELEASE_NAME} ${HELM_DIR} --namespace ${NAMESPACE}

kube-ingress-lb:
	kubectl apply -R -f ${HELM_DIR}/_system

lint:
	golint -set_exit_status ./...
