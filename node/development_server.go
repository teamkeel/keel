package node

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/teamkeel/keel/util"
)

type DevelopmentServer struct {
	cmd       *exec.Cmd
	exitError error
	output    io.Writer
	stdin     *io.PipeWriter
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

	if ds.stdin != nil {
		_ = ds.stdin.Close()
	}

	// See https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// for more info on this
	err := syscall.Kill(-ds.cmd.Process.Pid, syscall.SIGKILL)
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

// Rebuild triggers tsx to re-start the app in watch mode.
// tsx does not watch node_modules or dot-directories like .build and given
// we change files in these places we need a way to manually trigger a restart..
// The tsx docs say that in watch mode you can "Press Return to manually rerun"
// so we just send a newline to stdin.
func (ds *DevelopmentServer) Rebuild() error {
	if ds.stdin == nil {
		return nil
	}

	// Running in a go routine otherwise this will block until the go routine
	// reading from the other end of the pipe reads the data. We don't need to
	// wait for this to happen though.
	go func() {
		_, _ = ds.stdin.Write([]byte("\n"))
	}()

	return nil
}

type ServerOpts struct {
	Port    string
	EnvVars map[string]string
	Output  io.Writer
	Debug   bool
	Watch   bool
}

// RunDevelopmentServer will start a new node runtime server serving/handling custom function requests
func RunDevelopmentServer(dir string, options *ServerOpts) (*DevelopmentServer, error) {
	args := []string{".build/server.js"}
	if options != nil && options.Watch {
		args = append([]string{"watch"}, args...)
	}

	cmd := exec.Command("./node_modules/.bin/tsx", args...)

	// See https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// for more info on this
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	// Use a pipe for stdin so we can send "Enter" key presses in watch mode
	// See Rebuild func for more info
	var stdin *io.PipeWriter
	if options != nil && options.Watch {
		var r *io.PipeReader
		r, stdin = io.Pipe()
		cmd.Stdin = r
	}

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
		stdin:  stdin,
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
