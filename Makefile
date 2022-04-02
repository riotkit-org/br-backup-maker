include test_backupmaker.mk

BIN_PATH=$$(pwd)/.build/backup-maker

build_bm:
	go build -o ${BIN_PATH} ./cmd/backupmaker/main.go

test:
	cd context && go test
	cd client && go test
	go test

