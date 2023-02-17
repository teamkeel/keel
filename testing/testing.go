package testing

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"

	keelconfig "github.com/teamkeel/keel/config"
	"github.com/teamkeel/keel/db"
	"github.com/teamkeel/keel/functions"
	"github.com/teamkeel/keel/node"
	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime"
	"github.com/teamkeel/keel/runtime/runtimectx"
	"github.com/teamkeel/keel/schema"
	"github.com/teamkeel/keel/testhelpers"
	"github.com/teamkeel/keel/util"
)

type TestOutput struct {
	Output  string
	Success bool
}

type RunnerOpts struct {
	Dir             string
	Pattern         string
	DbConnInfo      *db.ConnectionInfo
	FunctionsOutput io.Writer
	EnvVars         map[string]string
}

func Run(opts *RunnerOpts) (*TestOutput, error) {
	builder := &schema.Builder{}

	schema, err := builder.MakeFromDirectory(opts.Dir)
	if err != nil {
		return nil, err
	}

	testApi := &proto.Api{
		// TODO: make random so doesn't clash
		Name: "TestingActionsApi",
	}
	for _, m := range schema.Models {
		testApi.ApiModels = append(testApi.ApiModels, &proto.ApiModel{
			ModelName: m.Name,
		})
	}

	schema.Apis = append(schema.Apis, testApi)

	context := context.Background()

	dbName := "keel_test"
	mainDB, err := testhelpers.SetupDatabaseForTestCase(context, opts.DbConnInfo, schema, dbName)
	if err != nil {
		return nil, err
	}

	dbConnString := opts.DbConnInfo.WithDatabase(dbName).String()

	files, err := node.Generate(
		context,
		opts.Dir,
		node.WithDevelopmentServer(true),
	)

	if err != nil {
		return nil, err
	}

	err = files.Write()
	if err != nil {
		return nil, err
	}

	var functionsServer *node.DevelopmentServer
	var functionsTransport functions.Transport

	if node.HasFunctions(schema) {
		keelEnvVars := map[string]string{
			"KEEL_DB_CONN_TYPE": "pg",
			"KEEL_DB_CONN":      dbConnString,
		}

		for key, value := range keelEnvVars {
			opts.EnvVars[key] = value
		}

		functionsServer, err = node.RunDevelopmentServer(opts.Dir, &node.ServerOpts{
			EnvVars: opts.EnvVars,
			Output:  opts.FunctionsOutput,
		})

		if err != nil {
			if functionsServer != nil && functionsServer.Output() != "" {
				return nil, errors.New(functionsServer.Output())
			}
			return nil, err
		}

		defer func() {
			_ = functionsServer.Kill()
		}()

		functionsTransport = functions.NewHttpTransport(functionsServer.URL)
	}

	runtimePort, err := util.GetFreePort()
	if err != nil {
		return nil, err
	}

	config, err := keelconfig.Load(opts.Dir)
	if err != nil {
		return nil, err
	}

	envVars := config.GetEnvVars("test")
	for key, value := range envVars {
		os.Setenv(key, value)
	}

	// Server to handle API calls to the runtime
	runtimeServer := http.Server{
		Addr: fmt.Sprintf(":%s", runtimePort),
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx := r.Context()
			database, _ := db.LocalFromConnection(ctx, mainDB)
			ctx = runtimectx.WithDatabase(ctx, database)
			if functionsTransport != nil {
				ctx = functions.WithFunctionsTransport(ctx, functionsTransport)
			}
			r = r.WithContext(ctx)

			runtime.NewHttpHandler(schema).ServeHTTP(w, r)
		}),
	}

	go func() {
		_ = runtimeServer.ListenAndServe()
	}()

	defer func() {
		_ = runtimeServer.Shutdown(context)
	}()

	cmd := exec.Command("npx", "tsc", "--noEmit", "--pretty")
	cmd.Dir = opts.Dir

	b, err := cmd.CombinedOutput()
	exitError := &exec.ExitError{}
	if err != nil && !errors.As(err, &exitError) {
		return nil, err
	}
	if err != nil {
		return &TestOutput{Output: string(b), Success: false}, nil
	}

	if opts.Pattern == "" {
		opts.Pattern = "(.*)"
	}

	cmd = exec.Command("npx", "vitest", "run", "--color", "--reporter", "verbose", "--config", "./.build/vitest.config.mjs", "--testNamePattern", opts.Pattern)
	cmd.Dir = opts.Dir
	cmd.Env = append(os.Environ(), []string{
		fmt.Sprintf("KEEL_TESTING_ACTIONS_API_URL=http://localhost:%s/testingactionsapi/json", runtimePort),
		"KEEL_DB_CONN_TYPE=pg",
		fmt.Sprintf("KEEL_DB_CONN=%s", dbConnString),
		// Disables experimental fetch warning that pollutes console experience when running tests
		"NODE_NO_WARNINGS=1",
	}...)

	b, err = cmd.CombinedOutput()
	if err != nil && !errors.As(err, &exitError) {
		return nil, err
	}

	return &TestOutput{Output: string(b), Success: err == nil}, nil
}
