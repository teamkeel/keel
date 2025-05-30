package util

import (
	"net"
	"strconv"
)

// GetFreePort finds an available port we can run Postgres on.
// It will favour the ones you specify as preferredPorts, but if they are
// all in use, it return an arbitrary port that is free.
func GetFreePort(preferredPorts ...string) (string, error) {
	var err error
	var port string

	for _, v := range append(preferredPorts, "0") {
		port, err = func() (string, error) {
			addr, err := net.ResolveTCPAddr("tcp", ":"+v)
			if err != nil {
				return "", err
			}

			listener, err := net.ListenTCP("tcp", addr)
			if err != nil {
				return "", err
			}
			defer listener.Close()

			port := listener.Addr().(*net.TCPAddr).Port
			return strconv.FormatInt(int64(port), 10), nil
		}()

		if port != "" {
			return port, nil
		}
	}

	return port, err
}
