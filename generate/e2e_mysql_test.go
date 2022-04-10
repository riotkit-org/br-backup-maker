package generate_test

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/riotkit-org/backup-maker/generate"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"os/exec"
	"testing"
)

// TestEndToEnd_MariaDBBackupAndRestore an End-To-End testing procedure for MariaDB/MySQL
func TestEndToEnd_MariaDBBackupAndRestoreProcedureBetweenInstances(t *testing.T) {
	WithBackupRepositoryDockerStack(func(stack ServiceStack) {
		ctx := context.Background()

		// ========================================================
		//  Send backup of MariaDB instance #1 to Backup Repository
		// ========================================================
		c, dbHostname, dbPort := CreateMariaDBContainer(ctx)
		writeDefinitionForLaterSnippetGeneration(`
Params:
	hostname: "` + dbHostname + `"
	user: "rojava"
	password: "rojava"
	port: "` + fmt.Sprintf("%v", dbPort) + `"
	db: ""

Repository:
	url: "http://` + stack.ServerHost + `:` + fmt.Sprintf("%v", stack.ServerPort) + `"
	token: "` + stack.AdminJwt + `"
	encryptionKeyPath: "resources/test/gp-key.asc"
	passphrase: "riotkit"
	recipient: "test@riotkit.org"
	collectionId: "iwa-ait"

`)
		generateMySQLSnippet("backup")
		subTestMySQLDumpBackup(t)
		_ = c.Terminate(ctx)

		// =================================================================================
		//  Receive Backup from Backup Repository and restore on a new MariaDB instance (#2)
		// =================================================================================
		c, dbHostname, dbPort = CreateMariaDBContainer(ctx)
		writeDefinitionForLaterSnippetGeneration(`
Params:
	hostname: "` + dbHostname + `"
	user: "rojava"
	password: "rojava"
	port: "` + fmt.Sprintf("%v", dbPort) + `"
	db: ""

Repository:
	url: "http://` + stack.ServerHost + `:` + fmt.Sprintf("%v", stack.ServerPort) + `"
	token: "` + stack.AdminJwt + `"
	encryptionKeyPath: "resources/test/gp-key.asc"
	passphrase: "riotkit"
	recipient: "test@riotkit.org"
	collectionId: "iwa-ait"

`)
		generateMySQLSnippet("restore")
		subTestMySQLRestoreBackup(t)
	})
}

func subTestMySQLDumpBackup(t *testing.T) {
	// run backup.sh
	cmd := exec.Command("/bin/bash", "-c", "export PATH=$PATH:./; bash backup.sh 2>&1")
	cmd.Dir = "../.build"
	cmd.Stderr = nil
	out, err := cmd.Output()

	assert.Nil(t, err)
	assert.Contains(t, string(out), "Version uploaded")
}

func subTestMySQLRestoreBackup(t *testing.T) {
	// run restore.sh
	cmd := exec.Command("/bin/bash", "-c", "export PATH=$PATH:./; bash restore.sh 2>&1")
	cmd.Dir = "../.build"
	cmd.Stderr = nil
	out, err := cmd.Output()

	assert.Nil(t, err)
	assert.Contains(t, string(out), "Backup restored")
}

func generateMySQLSnippet(operation string) {
	bs := generate.SnippetGenerationCommand{
		Template:       "mysql-dump",
		DefinitionFile: "../.build/definition.yaml",
		IsKubernetes:   false,
		KeyPath:        "../resources/test/gpg-key.asc",
		OutputDir:      "../.build/",
		Schedule:       "",
		JobName:        "",
		Image:          "",
		Operation:      operation,
		Namespace:      "backup-repository",
	}

	if err := bs.Run(); err != nil {
		log.Fatal(errors.Wrap(err, "Cannot generate backup snippet"))
	}
}
