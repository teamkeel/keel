package node

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/teamkeel/keel/util"
)

type DevelopmentServer struct {
	cmd       *exec.Cmd
	exitError error
	output    io.Writer
	URL       string
}

func (ds *DevelopmentServer) Output() string {
	b, ok := ds.output.(*bytes.Buffer)

	if !ok {
		return ""
	}

	return b.String()
}

func (ds *DevelopmentServer) Kill() error {
	if ds.cmd.Process == nil {
		return nil
	}

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
	Output  io.Writer
	Debug   bool
}

// RunDevelopmentServer will start a new node runtime server serving/handling custom function requests
func RunDevelopmentServer(dir string, options *ServerOpts) (*DevelopmentServer, error) {
	cmd := exec.Command("npx", "tsx", ".build/server.js")
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

	var output io.Writer

	if options == nil || options.Output == nil {
		output = new(bytes.Buffer)
	} else {
		output = options.Output
	}

	d := &DevelopmentServer{
		cmd:    cmd,
		output: output,
		URL:    fmt.Sprintf("http://localhost:%s", port),
	}

	cmd.Stdout = d.output
	cmd.Stderr = d.output

	if options != nil {
		if options.Debug {
			cmd.Env = append(cmd.Env, "DEBUG=true")
		}

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
