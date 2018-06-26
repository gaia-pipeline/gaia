default: dev

dev:
	go run ./cmd/gaia/main.go -homepath=${PWD}/tmp -dev true

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
	rice embed-go

release: compile_frontend static_assets
	env GOOS=linux GOARCH=amd64 go build -o gaia-linux-amd64 ./cmd/gaia/main.go