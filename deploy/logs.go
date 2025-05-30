package deploy

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/TylerBrock/colorjson"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/teamkeel/keel/colors"
	"github.com/teamkeel/keel/config"
	"golang.org/x/sync/errgroup"
)

type StreamLogsArgs struct {
	ProjectRoot string
	Env         string
	StartTime   time.Time
}

// TODO: refactor this function to use an error channel
//
//nolint:staticcheck
func StreamLogs(ctx context.Context, args *StreamLogsArgs) error {
	configFiles, err := config.LoadAll(args.ProjectRoot)
	if err != nil {
		return err
	}

	var projectConfig *config.ProjectConfig
	for _, f := range configFiles {
		if filepath.Base(f.Filename) == fmt.Sprintf("keelconfig.%s.yaml", args.Env) {
			if f.Errors != nil {
				return f.Errors
			}
			projectConfig = f.Config
		}
	}
	if projectConfig == nil {
		return fmt.Errorf("no keelconfig.%s.yaml file found", args.Env)
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(projectConfig.Deploy.Region))
	if err != nil {
		return err
	}

	pulumiConfig, err := setupPulumi(ctx, &SetupPulumiArgs{
		AwsConfig: cfg,
		Config:    projectConfig,
		Env:       args.Env,
	})
	if err != nil {
		return err
	}

	outputs, err := getStackOutputs(ctx, &GetStackOutputsArgs{
		Config:       projectConfig,
		PulumiConfig: pulumiConfig,
	})
	if err != nil {
		return err
	}

	logs := cloudwatchlogs.NewFromConfig(cfg)

	for err == nil {
		g, ctx := errgroup.WithContext(ctx)
		events := []types.FilteredLogEvent{}
		var s sync.Mutex
		lambdaNames := []string{
			outputs.ApiLambdaName,
			outputs.SubscriberLambdaName,
			outputs.JobsLambdaName,
			outputs.FunctionsLambdaName,
		}

		for _, name := range lambdaNames {
			g.Go(func() error {
				e, err := fetchLogs(ctx, logs, name, args)
				if err != nil {
					log(ctx, "%s Error reading logs for %s: %s", IconCross, orange(name), gray(err.Error()))
					return err
				}

				s.Lock()
				defer s.Unlock()
				events = append(events, e...)
				return nil
			})
		}

		err = g.Wait()
		if err != nil {
			break
		}

		if len(events) == 0 {
			// TODO: consider making this value configurable
			time.Sleep(time.Second * 5)
			args.StartTime = args.StartTime.Add(time.Second)
			continue
		}

		sort.Slice(events, func(i, j int) bool {
			return *events[i].Timestamp < *events[j].Timestamp
		})

		for _, e := range events {
			log(ctx, *e.Message)
			t := time.Unix(0, (*e.Timestamp * int64(time.Millisecond)))
			t = t.Add(time.Second)
			args.StartTime = t
		}
	}

	return err
}

func fetchLogs(ctx context.Context, logs *cloudwatchlogs.Client, lambdaName string, args *StreamLogsArgs) ([]types.FilteredLogEvent, error) {
	var nextToken *string
	startTime := args.StartTime.UnixMilli()
	events := []types.FilteredLogEvent{}

	f := colorjson.NewFormatter()
	f.Indent = 2

	for {
		input := &cloudwatchlogs.FilterLogEventsInput{
			LogGroupName: aws.String(fmt.Sprintf("/aws/lambda/%s", lambdaName)),
			StartTime:    &startTime,
			NextToken:    nextToken,
		}

		out, err := logs.FilterLogEvents(ctx, input)
		if err != nil {
			// A log group won't exist until the Lambda has run at least once, so not found is ok and expected if you've just deployed
			if isSmithyAPIError(err, "ResourceNotFoundException") {
				return []types.FilteredLogEvent{}, nil
			}

			return nil, err
		}

		for _, e := range out.Events {
			m := formatLog(lambdaName, *e.Message)
			if m == "" {
				continue
			}

			e.Message = aws.String(formatLog(lambdaName, *e.Message))
			events = append(events, e)
		}

		nextToken = out.NextToken
		if nextToken == nil {
			return events, nil
		}
	}
}

func formatLog(lambdaName string, rawMessage string) string {
	log := map[string]any{}
	err := json.Unmarshal([]byte(rawMessage), &log)
	if err != nil {
		return rawMessage
	}

	// Drop the Lambda platform.* logs
	// TODO: consider adding a CLI flag for including these?
	t, ok := log["type"].(string)
	if ok && strings.HasPrefix(t, "platform.") {
		return ""
	}

	f := colorjson.NewFormatter()
	f.Indent = 2

	result := ""

	nameParts := strings.Split(lambdaName, "-")
	name := strings.Join(nameParts[0:len(nameParts)-1], "-")
	result += colors.Magenta(fmt.Sprintf("[%s]", name)).String()

	// Normalise between Go and Node logs
	ts, ok := log["time"].(string)
	if !ok {
		ts, ok = log["timestamp"].(string)
	}
	if ok {
		result += fmt.Sprintf(" %s", colors.Gray(fmt.Sprintf("[%s]", ts)).String())
	}

	level, ok := log["level"].(string)
	if ok {
		result += fmt.Sprintf(" %s", colors.Yellow(strings.ToUpper(level)).String())
	}

	// Normalise between Go and Node logs
	message, ok := log["message"]
	if !ok {
		message, ok = log["msg"]
	}
	if ok {
		m := map[string]any{}
		switch v := message.(type) {
		case string:
			_ = json.Unmarshal([]byte(v), &m)
		case map[string]any:
			m = v
		}
		if len(m) > 0 {
			b, err := f.Marshal(m)
			if err == nil {
				message = string(b)
			}
		}
		result += fmt.Sprintf(" %s", message)
	}

	rest := map[string]any{}
	for k, v := range log {
		switch k {
		case "msg", "message", "requestId", "time", "timestamp", "level":
			// We've already handled these values so ignore them
		default:
			rest[k] = v
		}
	}
	if len(rest) > 0 {
		b, _ := f.Marshal(rest)
		result += fmt.Sprintf(" %s", string(b))
	}

	return result
}
