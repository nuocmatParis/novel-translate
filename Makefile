.PHONY: build run test vet install

build:
	go build -trimpath -o novel-translate-cli ./cmd/novel-translate-cli

run: build
	./novel-translate-cli tui

test:
	go test -buildvcs=false -count=1 ./...

vet:
	go vet ./...

install:
	sh scripts/install.sh
