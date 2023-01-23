package web

import (
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"

	"github.com/pkg/browser"
	"github.com/teamkeel/keel/cmd/config"
)

type Controller struct {
	Cfg *config.Config
}

type LoginResponse struct {
	Status string `json:"status,omitempty"`
	Error  string `json:"error,omitempty"`
}

const (
	loginInvalidResponse string = "Invalid code"
	loginSuccessResponse string = "Ok"
)

func (c *Controller) Login(ctx context.Context) (*string, error) {
	token, err := c.browserBasedLogin(ctx)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (c *Controller) browserBasedLogin(ctx context.Context) (*string, error) {
	var token string
	var returnedCode string
	port, err := getPort()

	if err != nil {
		return nil, err
	}

	code := fmt.Sprintf("%016d", rand.Int63n(1e16))

	wg := &sync.WaitGroup{}
	wg.Add(2)
	go func() {
		ctx := context.Background()
		srv := &http.Server{Addr: strconv.Itoa(port)}
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")

			if r.Method == http.MethodGet {
				w.Header().Set("Content-Type", "application/json")
				token = r.URL.Query().Get("token")
				returnedCode = r.URL.Query().Get("code")

				if code != returnedCode {
					res := LoginResponse{Error: loginInvalidResponse}
					byteRes, err := json.Marshal(&res)
					if err != nil {
						fmt.Println(err)
					}
					w.WriteHeader(400)
					_, err = w.Write(byteRes)
					if err != nil {
						fmt.Println("response is invalid")
					}
					return
				}

				res := LoginResponse{Status: loginSuccessResponse}
				byteRes, err := json.Marshal(&res)
				if err != nil {
					fmt.Println(err)
				}
				w.WriteHeader(200)
				_, err = w.Write(byteRes)
				if err != nil {
					fmt.Println("response is invalid")
				}
			} else if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, PUT, PATCH, POST, DELETE")
				w.Header().Set("Access-Control-Allow-Headers", "*")
				w.Header().Set("Content-Length", "0")
				w.WriteHeader(204)
				return
			}

			wg.Done()

			if err := srv.Shutdown(ctx); err != nil {
				fmt.Println(err)
			}
		})

		if err := http.ListenAndServe(fmt.Sprintf("localhost:%d", port), nil); err != nil {
			fmt.Println("Login server handshake failed!")
		}
	}()

	url := createLoginURL(port, code)
	err = c.BrowserPrompt("Logging in", url)
	if err != nil {
		return nil, err
	}

	wg.Wait()

	if code != returnedCode {
		return nil, errors.New("login failed")
	}

	err = c.Cfg.SetUserConfig(&config.UserConfig{
		Token: token,
	})
	if err != nil {
		return nil, err
	}

	err = c.Cfg.SetProjectConfig()
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (c *Controller) BrowserPrompt(spinnerMsg string, url string) error {
	fmt.Printf("Press Enter to open default web browser (^C to quit)")
	fmt.Fscanln(os.Stdin)

	err := browser.OpenURL(url)

	if err != nil {
		return err
	}

	return nil
}

func getPort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func createLoginURL(port int, code string) string {
	hostname := getHostName()
	buffer := b64.URLEncoding.EncodeToString([]byte(fmt.Sprintf("port=%d&code=%s&hostname=%s", port, code, hostname)))
	url := fmt.Sprintf("%s/login/cli?d=%s", "http://localhost:5173", buffer)
	return url
}

func getHostName() string {
	name, err := os.Hostname()
	if err != nil {
		return ""
	}

	return name
}
