include test_backupmaker.mk
include test_generator.mk

BM_BIN_PATH=$$(pwd)/.build/backup-maker
BMG_BIN_PATH=$$(pwd)/.build/bmg

build_bm:
	CGO_ENABLED=0 GO111MODULE=on go build -o ${BM_BIN_PATH} ./cmd/backupmaker/main.go

build_bmg:
	CGO_ENABLED=0 GO111MODULE=on go build -o ${BMG_BIN_PATH} ./cmd/bmg/main.go

test:
	go test -v ./...

coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out
