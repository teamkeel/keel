package node

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/teamkeel/keel/util"
)

type DevelopmentServer struct {
	cmd       *exec.Cmd
	exitError error
	output    *bytes.Buffer
	URL       string
}

func (ds *DevelopmentServer) Output() string {
	return ds.output.String()
}

func (ds *DevelopmentServer) Kill() error {
	if ds.cmd.Process == nil {
		return nil
	}

	// For some reason ts-node doesn't seem to die on SIGKILL, but
	// SIGINT seems to work fine
	err := ds.cmd.Process.Signal(os.Interrupt)
	if err != nil {
		return err
	}

	// wait for ProcessState to be set
	n := time.Now()
	maxWait := time.Second * 10
	for ds.cmd.ProcessState == nil {
		if time.Since(n) > maxWait {
			return fmt.Errorf("development server failed to exit after %s", maxWait.String())
		}
		time.Sleep(time.Millisecond * 100)
	}

	return ds.exitError
}

type ServerOpts struct {
	Port    string
	EnvVars map[string]string
}

// RunDevelopmentServer will start a new node runtime server serving/handling custom function requests
func RunDevelopmentServer(dir string, options *ServerOpts) (*DevelopmentServer, error) {
	cmd := exec.Command("npx", "ts-node", ".build/server.js")
	cmd.Dir = dir
	cmd.Env = os.Environ()

	var port string
	if options != nil && options.Port != "" {
		port = options.Port
	} else {
		var err error
		port, err = util.GetFreePort()
		if err != nil {
			return nil, err
		}
	}

	d := &DevelopmentServer{
		cmd:    cmd,
		output: &bytes.Buffer{},
		URL:    fmt.Sprintf("http://localhost:%s", port),
	}

	cmd.Stdout = d.output
	cmd.Stderr = d.output

	if options != nil {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%s", port))

		for key, value := range options.EnvVars {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	err := cmd.Start()
	if err != nil {
		return nil, err
	}

	go func() {
		d.exitError = cmd.Wait()
	}()

	// Wait for process to have started successfully
	n := time.Now()
	maxWait := time.Second * 10
	for {
		if cmd.ProcessState != nil {
			return d, d.exitError
		}
		res, _ := http.Get(d.URL + "/_health")
		if res != nil && res.StatusCode == http.StatusOK {
			break
		}
		if time.Since(n) > maxWait {
			_ = d.Kill()
			return d, fmt.Errorf("development server failed to start after %s", maxWait.String())
		}
		time.Sleep(time.Millisecond * 500)
	}

	return d, nil
}
