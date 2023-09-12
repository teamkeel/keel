//go:build unix

package node

import "syscall"

func (ds *DevelopmentServer) getSysProcAttr() *syscall.SysProcAttr {
	// See https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// for more info on this
	return &syscall.SysProcAttr{Setpgid: true}
}

func (ds *DevelopmentServer) kill() error {
	// See https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// for more info on this
	return syscall.Kill(-ds.cmd.Process.Pid, syscall.SIGKILL)
}
