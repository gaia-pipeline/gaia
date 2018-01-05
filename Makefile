default: run

build: ./cmd/gaia/main.go
	go install ./cmd/gaia/

run: build
	gaia