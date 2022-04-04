include test_backupmaker.mk
include test_generator.mk
include versions.mk

BM_BIN_PATH=$$(pwd)/.build/backup-maker
BMG_BIN_PATH=$$(pwd)/.build/bmg

build_bm:
	CGO_ENABLED=0 GO111MODULE=on go build -o ${BM_BIN_PATH} ./cmd/backupmaker/main.go

build_bmg:
	CGO_ENABLED=0 GO111MODULE=on go build -o ${BMG_BIN_PATH} ./cmd/bmg/main.go

build_docker: ## Builds docker image. Uses already built artifacts
	docker build . --build-arg BR_PGBR_VERSION=${BR_PGBR_VERSION} --build-arg BR_PGBR_DEFAULT_PG=${POSTGRES_VERSION} -t ghcr.io/riotkit-org/backup-maker:${DOCKER_TAG}

push_docker: ## Release docker
	docker push ghcr.io/riotkit-org/backup-maker:${DOCKER_TAG}

test:
	go test -v ./...

coverage:
	go test -v ./... -covermode=count -coverprofile=coverage.out
