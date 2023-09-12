//go:build windows

package node

import "syscall"

func (ds *DevelopmentServer) getSysProcAttr() *syscall.SysProcAttr {
	return nil
}

func (ds *DevelopmentServer) kill() error {
	// See https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
	// for more info on this
	return ds.cmd.Process.Kill()
}
