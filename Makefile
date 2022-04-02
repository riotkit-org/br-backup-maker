include test_backupmaker.mk

BM_BIN_PATH=$$(pwd)/.build/backup-maker
BMG_BIN_PATH=$$(pwd)/.build/bmg

build_bm:
	go build -o ${BM_BIN_PATH} ./cmd/backupmaker/main.go

build_bmg:
	go build -o ${BMG_BIN_PATH} ./cmd/bmg/main.go

test:
	go test -v ./...

coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out
