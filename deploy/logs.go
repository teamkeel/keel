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
	// Events      chan Output
}

func StreamLogs(ctx context.Context, args *StreamLogsArgs) error {
	// defer func() {
	// 	close(args.Events)
	// }()

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

	// Not interested in passing the output from setupPulumi so we'll just consume these events here
	// checking for any erors
	ch := make(chan Output)
	go func() {
		for o := range ch {
			if o.Error != nil {
				err = o.Error
			}
		}
	}()

	pulumiConfig := setupPulumi(ctx, &SetupPulumiArgs{
		AwsConfig: cfg,
		Config:    projectConfig,
		Env:       args.Env,
		Events:    ch,
	})
	if err != nil {
		return err
	}
	close(ch)

	outputs, err := getStackOutputs(ctx, pulumiConfig)
	if err != nil {
		return err
	}

	logs := cloudwatchlogs.NewFromConfig(cfg)

	for {
		g, ctx := errgroup.WithContext(ctx)
		events := []types.FilteredLogEvent{}
		var s sync.Mutex
		lambdaNames := []string{outputs.ApiLambdaName, outputs.SubscriberLambdaName, outputs.FunctionsLambdaName}

		for _, name := range lambdaNames {
			name := name
			g.Go(func() error {
				e, err := fetchLogs(ctx, logs, name, args)
				if err != nil {
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
			return err
		}

		if len(events) == 0 {
			time.Sleep(time.Second * 5)
			args.StartTime = args.StartTime.Add(time.Second)
			continue
		}

		sort.Slice(events, func(i, j int) bool {
			return *events[i].Timestamp < *events[j].Timestamp
		})

		for _, e := range events {
			fmt.Println(*e.Message)
			t := time.Unix(0, (*e.Timestamp * int64(time.Millisecond)))
			t = t.Add(time.Second)
			args.StartTime = t
		}
	}
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
			return nil, err
		}

		for _, e := range out.Events {
			// // Message should be in JSON format
			// jsonLog := map[string]interface{}{}
			// err = json.Unmarshal([]byte(*e.Message), &jsonLog)
			// if err != nil {
			// 	// If not then just output as-is
			// 	e.Message = aws.String(fmt.Sprintf("[%s] %s", lambdaName, *e.Message))
			// 	events = append(events, e)
			// 	continue
			// }

			// // Drop the Lambda platform.* logs
			// // TODO: maybe add a CLI flag for including these?
			// if isLambdaSystemLog(jsonLog) {
			// 	continue
			// }

			// // Colorise JSON
			// s, _ := f.Marshal(jsonLog)
			// message := string(s)

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

// func isLambdaSystemLog(log map[string]interface{}) bool {
// 	t, ok := log["type"].(string)
// 	return ok && strings.HasPrefix(t, "platform.")
// }

func formatLog(lambdaName string, rawMessage string) string {
	log := map[string]any{}
	err := json.Unmarshal([]byte(rawMessage), &log)
	if err != nil {
		return rawMessage
	}

	//  Drop the Lambda platform.* logs
	// TODO: maybe add a CLI flag for including these?
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
			// nothing
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
