package generate_test

import (
	"fmt"
	"github.com/riotkit-org/backup-maker/generate"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os/exec"
	"testing"
)

// TestEndToEnd_TarBackupAndRestore is testing End-To-End a procedure of Backup & Restore of a directory packed into a TAR.GZ archive
func TestEndToEnd_TarBackupAndRestore(t *testing.T) {
	WithBackupRepositoryDockerStack(func(stack ServiceStack) {
		writeDefinition(`
Params:
    path: ../cmd

Repository:
    url: "http://` + stack.ServerHost + `:` + fmt.Sprintf("%v", stack.ServerPort) + `"
    token: "` + stack.AdminJwt + `"
    encryptionKeyPath: "resources/test/gp-key.asc"
    passphrase: "riotkit"
    recipient: "test@riotkit.org"
    collectionId: "iwa-ait"

`)
		subTestTarBackup(t)
		subTestTarRestore(t)
	})
}

func subTestTarBackup(t *testing.T) {
	bs := generate.SnippetGenerationCommand{
		Template:       "tar",
		DefinitionFile: "../.build/definition.yaml",
		IsKubernetes:   false,
		KeyPath:        "../resources/test/gpg-key.asc",
		OutputDir:      "../.build/",
		Schedule:       "",
		JobName:        "",
		Image:          "",
		Operation:      "backup",
		Namespace:      "backup-repository",
	}

	// generate backup.sh
	err := bs.Run()
	assert.Nil(t, err)

	// run backup.sh
	cmd := exec.Command("/bin/bash", "-c", "export PATH=$PATH:./; bash backup.sh 2>&1")
	cmd.Dir = "../.build"
	cmd.Stderr = nil
	out, err := cmd.Output()

	// backup verification
	assert.Nil(t, err, string(out))
	assert.Contains(t, string(out), "Version uploaded")
	assert.Contains(t, string(out), "cmd/backupmaker/main.go")
	assert.Contains(t, string(out), "cmd/bmg/main.go")
}

func subTestTarRestore(t *testing.T) {
	rs := generate.SnippetGenerationCommand{
		Template:       "tar",
		DefinitionFile: "../.build/definition.yaml",
		IsKubernetes:   false,
		KeyPath:        "../resources/test/gpg-key.asc",
		OutputDir:      "../.build/",
		Schedule:       "",
		JobName:        "",
		Image:          "",
		Operation:      "restore",
		Namespace:      "backup-repository",
	}

	// generate restore.sh
	restoreErr := rs.Run()
	assert.Nil(t, restoreErr)

	// run restore.sh
	cmd := exec.Command("/bin/bash", "-c", "export PATH=$PATH:./; bash restore.sh 2>&1")
	cmd.Dir = "../.build"
	cmd.Stderr = nil
	out, err := cmd.Output()

	assert.Nil(t, err, string(out))
	assert.Contains(t, string(out), "Backup restored")
	assert.Contains(t, string(out), "cmd/backupmaker/main.go")
	assert.Contains(t, string(out), "cmd/bmg/main.go")
}

func writeDefinition(content string) {
	_ = ioutil.WriteFile("../.build/definition.yaml", []byte(content), 0755)
}
