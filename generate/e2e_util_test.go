package generate_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type ServiceStack struct {
	ServerHost string
	ServerPort int

	StorageHost string
	StoragePort int

	PostgresHost string
	PostgresPort int

	AdminJwt string
}

func WithBackupRepositoryDockerStack(test func(ServiceStack)) {
	ctx := context.Background()
	storage, storageHost, storagePort := createMinioContainer(ctx)
	postgres, postgresHost, postgresPort := createPostgresContainer(ctx)
	server, serverHost, serverPort := createServerContainer(ctx, postgresHost, postgresPort, storageHost, storagePort)

	jwt := loginToServer(serverHost, serverPort, "admin", "admin")

	defer storage.Terminate(ctx)
	defer postgres.Terminate(ctx)
	defer server.Terminate(ctx)

	test(ServiceStack{
		ServerHost:   serverHost,
		ServerPort:   serverPort,
		StorageHost:  storageHost,
		StoragePort:  storagePort,
		PostgresHost: postgresHost,
		PostgresPort: postgresPort,
		AdminJwt:     jwt,
	})
}

// loginToServer is getting a JWT token for given username and password
func loginToServer(serverHost string, serverPort int, user string, password string) string {
	postBody, _ := json.Marshal(map[string]string{
		"username": user,
		"password": password,
	})
	body := bytes.NewBuffer(postBody)
	response, err := http.Post(fmt.Sprintf("http://%s:%v/api/stable/auth/login", serverHost, serverPort), "application/json", body)

	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot make authorization request"))
	}
	if response.StatusCode > 200 {
		log.Fatalf("Cannot authenticate user login=%s, password=%s, response code: %v", user, password, response.StatusCode)
	}

	responseBuffer, readErr := ioutil.ReadAll(response.Body)
	if readErr != nil {
		log.Fatal(errors.Wrap(readErr, "Cannot read response"))
	}

	var parsedResponse = struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}{}

	if err := json.Unmarshal(responseBuffer, &parsedResponse); err != nil {
		log.Fatalf("Cannot parse server response: %v, Error: %s", string(responseBuffer), err.Error())
	}

	return parsedResponse.Data.Token
}

// Backup Repository server container on-demand
func createServerContainer(ctx context.Context, postgresHost string, postgresPort int, minioHost string, minioPort int) (testcontainers.Container, string, int) {
	version := os.Getenv("TEST_BACKUP_REPOSITORY_VERSION")
	req := testcontainers.ContainerRequest{
		Image: "ghcr.io/riotkit-org/backup-repository:" + version,
		Env: map[string]string{
			"BR_DB_HOSTNAME": postgresHost,
			"BR_DB_USERNAME": "rojava",
			"BR_DB_PASSWORD": "rojava",
			"BR_DB_NAME":     "emma-goldman",
			"BR_DB_PORT":     fmt.Sprintf("%v", postgresPort),

			"BR_STORAGE_DRIVER_URL": fmt.Sprintf("s3://orwell1984?endpoint=%s:%v&disableSSL=true&s3ForcePathStyle=true&region=eu-central-1", minioHost, minioPort),
			"AWS_ACCESS_KEY_ID":     "AKIAIOSFODNN7EXAMPLE",
			"AWS_SECRET_ACCESS_KEY": "wJaFuCKtnFEMI/CApItaliSM/bPxRfiCYEXAMPLEKEY",

			"BR_JWT_SECRET_KEY":   "anarchism-is-the-key",
			"BR_HEALTH_CHECK_KEY": "to-cooperate-for-whole-world",
			"BR_LOG_LEVEL":        "debug",
			"GIN_MODE":            "debug",

			"BR_CONFIG_LOCAL_PATH": "/mnt/filesystem-config",
			"BR_CONFIG_PROVIDER":   "filesystem",
		},
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForHTTP("/ready?code=to-cooperate-for-whole-world").WithPort("8080").WithPollInterval(time.Second),
		Networks:     []string{"backup-repository-e2e"},
		NetworkAliases: map[string][]string{
			"backup-repository-e2e": {
				"server",
			},
		},
	}
	wd, _ := os.Getwd()
	req.Mounts = testcontainers.ContainerMounts{
		testcontainers.ContainerMount{
			Source:   testcontainers.GenericBindMountSource{HostPath: wd + "/../.build"},
			Target:   "/mnt",
			ReadOnly: false,
		},
	}

	return createContainer(ctx, req, 8080)
}

func createMinioContainer(ctx context.Context) (testcontainers.Container, string, int) {
	version := os.Getenv("TEST_MINIO_VERSION")
	req := testcontainers.ContainerRequest{
		Image: "bitnami/minio:" + version,
		Env: map[string]string{
			"MINIO_DEFAULT_BUCKETS": "orwell1984",
			"MINIO_ROOT_USER":       "AKIAIOSFODNN7EXAMPLE",
			"MINIO_ROOT_PASSWORD":   "wJaFuCKtnFEMI/CApItaliSM/bPxRfiCYEXAMPLEKEY",
		},
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForLog("Console:"),
		Networks:     []string{"backup-repository-e2e"},
		NetworkAliases: map[string][]string{
			"backup-repository-e2e": {
				"storage",
			},
		},
	}
	return createContainer(ctx, req, 9000)
}

// PostgreSQL container on-demand
func createPostgresContainer(ctx context.Context) (testcontainers.Container, string, int) {
	version := os.Getenv("TEST_POSTGRES_VERSION")
	req := testcontainers.ContainerRequest{
		Image: "postgres:" + version,
		Env: map[string]string{
			"POSTGRES_USER":     "rojava",
			"POSTGRES_PASSWORD": "rojava",
			"POSTGRES_DB":       "emma-goldman",
		},
		ExposedPorts: []string{"5432/tcp"},
		WaitingFor:   wait.ForLog("database system is ready to accept connections"),
		Networks:     []string{"backup-repository-e2e"},
		NetworkAliases: map[string][]string{
			"backup-repository-e2e": {
				"postgres",
			},
		},
	}
	return createContainer(ctx, req, 5432)
}

func createContainer(ctx context.Context, req testcontainers.ContainerRequest, mappedPort int) (testcontainers.Container, string, int) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		log.Fatal(errors.Wrap(err, "Cannot create container"))
	}
	ip, ipErr := container.ContainerIP(ctx)
	if ipErr != nil {
		log.Fatal(errors.Wrap(err, "Cannot get container IP"))
	}

	return container, strings.ReplaceAll(ip, "/", ""), mappedPort
}
