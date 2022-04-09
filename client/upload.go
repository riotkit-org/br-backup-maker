package client

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	actionCtx "github.com/riotkit-org/backup-maker/context"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

// gracefullyKillProcess attempts to clean up created process tree
// to avoid keeping zombie processes
func gracefullyKillProcess(cmd *exec.Cmd) error {
	var killErr error = nil

	log.Println("Stopping process")

	if cmd.ProcessState == nil || cmd.ProcessState.Exited() {
		return nil
	}

	// protect against zombie processes
	for retry := 0; retry < 5; retry++ {
		killErr = cmd.Process.Kill()

		if killErr == nil {
			break
		}
		log.Print("Attempting to end backup process")
		time.Sleep(5 * time.Second)
	}

	// after multiple retries just kill all children processes with force
	if killErr != nil {
		log.Print("Cannot end main process, killing all children processes")
		proc := exec.Command("/bin/bash", "-c", fmt.Sprintf("kill -KILL -\"%v\"", cmd.Process.Pid))
		killErr = proc.Run()
	}

	if killErr != nil {
		return errors.New(
			fmt.Sprintf(
				"Cannot kill backup process with it's children processes after "+
					"successful upload. Watch out for zombie processes. %v", killErr))
	}

	return nil
}

// Upload is uploading bytes read from io.Reader stream into HTTP endpoint of Backup Repository server
func Upload(ctx context.Context, domainWithSchema string, collectionId string, authToken string, body io.ReadCloser, timeout int64) (string, string, error) {
	if timeout == 0 {
		timeout = int64(time.Second * 60 * 20)
	}

	url := fmt.Sprintf("%v/api/alpha/repository/collection/%v/version", domainWithSchema, collectionId)
	log.Printf("Uploading to %v", url)

	client := http.Client{}
	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		url,
		body)

	client.Timeout = time.Second * 3600 // todo parametrize
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %v", authToken))

	if err != nil {
		log.Println(err)
		return "?", "-", errors.Wrap(err, "Request creation failed")
	}
	var content []byte
	resp, respError := client.Do(req)
	if resp == nil {
		return "?", "-", errors.Wrap(respError, "Cannot read response, unknown error")
	}
	if resp.Body != nil {
		var readErr error
		content, readErr = ioutil.ReadAll(resp.Body)
		if readErr != nil {
			log.Println(readErr)
			return resp.Status, string(content), errors.Wrap(readErr, "Failed to read response sent by server")
		}
	}
	if respError != nil {
		log.Println(respError)
		return resp.Status, string(content), errors.Wrap(respError, "Request execution failed")
	}

	if resp.Status != "200 OK" {
		return resp.Status,
			string(content),
			errors.New(fmt.Sprintf("Request to server failed, server returned %v", string(content)))
	}

	return resp.Status, string(content), nil
}

// UploadFromCommandOutput pushes a stdout of executed command through HTTP endpoint of Backup Repository under specified domain
// Upload is used to perform HTTP POST request
func UploadFromCommandOutput(app actionCtx.Action) error {
	log.Print("/bin/bash", GetShellCommand(app.GetCommand("")))
	cmd := exec.Command("/bin/bash", GetShellCommand(app.GetCommand(""))...)
	cmd.Stderr = os.Stderr
	stdout, pipeErr := cmd.StdoutPipe()
	if pipeErr != nil {
		log.Println(pipeErr)
		return pipeErr
	}

	log.Print("Starting cmd.Run()")
	execErr := cmd.Start()
	if execErr != nil {
		log.Println("Cannot start backup process ", execErr)
		return execErr
	}

	ctx, cancel := context.WithCancel(context.TODO())

	log.Printf("Starting Upload() for PID=%v", cmd.Process.Pid)
	status, out, uploadErr := Upload(ctx, app.Url, app.CollectionId, app.AuthToken, ReadCloserWithCancellationWhenProcessFails{stdout, cmd, cancel}, app.Timeout)
	if uploadErr != nil {
		log.Errorf("Status: %v, Out: %v, Err: %v", status, out, uploadErr)
		return uploadErr
	} else {
		killErr := gracefullyKillProcess(cmd)
		if killErr != nil {
			return killErr
		}
	}
	log.Info("Version uploaded")

	return nil
}

type ReadCloserWithCancellationWhenProcessFails struct {
	Parent  io.ReadCloser
	Process *exec.Cmd
	Cancel  func()
}

func (r ReadCloserWithCancellationWhenProcessFails) Read(p []byte) (n int, err error) {
	return r.Parent.Read(p)
}

func (r ReadCloserWithCancellationWhenProcessFails) Close() error {
	err := r.Process.Wait()
	exitCode := 0
	if err != nil {
		// try to get the exit code
		if exitError, ok := err.(*exec.ExitError); ok {
			ws := exitError.Sys().(syscall.WaitStatus)
			exitCode = ws.ExitStatus()
		} else {
			exitCode = 1
		}
	} else {
		ws := r.Process.ProcessState.Sys().(syscall.WaitStatus)
		exitCode = ws.ExitStatus()
	}

	if exitCode > 0 {
		log.Errorf("Canceling upload due to process failure - exitCode: %v", exitCode)
		r.Cancel()
	}

	// do not block the response from being read in case of e.g. 401, 403, 500
	closeErr := r.Parent.Close()
	if strings.Contains(closeErr.Error(), "file already closed") {
		return nil
	}

	return err
}
