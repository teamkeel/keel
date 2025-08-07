package testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go/service/ssm"
)

type S3Object struct {
	Headers http.Header
	Data    []byte
}

type AWSAPIHandler struct {
	m             sync.Mutex
	PathPrefix    string
	SSMParameters map[string]string
	FunctionsURL  string
	FunctionsARN  string
	S3Bucket      map[string]*S3Object
	OnSQSEvent    map[string]func(event events.SQSEvent) // map of event handlers for each queueURL
}

func (h *AWSAPIHandler) HandleHTTP(r *http.Request, w http.ResponseWriter) {
	h.m.Lock()
	if h.S3Bucket == nil {
		h.S3Bucket = map[string]*S3Object{}
	}
	h.m.Unlock()

	s3Prefixes := []string{
		h.PathPrefix + "files/",
		h.PathPrefix + "jobs/",
	}
	isS3 := func() bool {
		for _, prefix := range s3Prefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				return true
			}
		}
		return false
	}

	switch {
	case r.Header.Get("X-Amz-Target") == "AmazonSSM.GetParameters":
		h.ssmGetParameters(w, r)
		return
	case r.Header.Get("X-Amz-Target") == "AmazonSQS.SendMessage":
		h.sqsSendMessage(r, w)
		return
	case r.URL.Path == fmt.Sprintf("/aws/2015-03-31/functions/%s/invocations", h.FunctionsARN):
		h.lambdaInvoke(r, w)
	case r.Method == http.MethodPut && isS3():
		h.s3PutObject(r, w)
		return
	case r.Method == http.MethodGet && isS3():
		h.s3GetObject(r, w)
		return
	default:
		fmt.Println("Unhandled AWS request", r.Method, r.URL.Path, r.Header)
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(""))
		return
	}
}

// https://docs.aws.amazon.com/systems-manager/latest/APIReference/API_GetParameters.html
func (h *AWSAPIHandler) ssmGetParameters(w http.ResponseWriter, r *http.Request) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Request for SSM parameters are done in groups of 10 keys at a time so it's important we
	// only return the ones that were actually asked for
	var input ssm.GetParametersInput
	err = json.Unmarshal(b, &input)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, err.Error())
		return
	}

	res := ssm.GetParametersOutput{}
	for _, key := range input.Names {
		parts := strings.Split(*key, "/")
		name := parts[len(parts)-1]
		value, ok := h.SSMParameters[name]
		if ok {
			res.Parameters = append(res.Parameters, &ssm.Parameter{
				Name:  key,
				Value: aws.String(value),
			})
		} else {
			res.InvalidParameters = append(res.InvalidParameters, key)
		}
	}

	writeJSON(w, http.StatusOK, res)
}

// https://docs.aws.amazon.com/lambda/latest/api/API_Invoke.html
func (h *AWSAPIHandler) lambdaInvoke(r *http.Request, w http.ResponseWriter) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil)
		return
	}

	functionsResponse, err := http.Post(h.FunctionsURL, "application/json", bytes.NewReader(requestBody))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil)
		return
	}

	responseBody, err := io.ReadAll(functionsResponse.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil)
		return
	}

	w.WriteHeader(functionsResponse.StatusCode)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(responseBody)
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_PutObject.html
func (h *AWSAPIHandler) s3PutObject(r *http.Request, w http.ResponseWriter) {
	b, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil)
		return
	}

	key := strings.TrimPrefix(r.URL.Path, h.PathPrefix)

	h.m.Lock()
	h.S3Bucket[key] = &S3Object{
		// Store the headers so we can return them in GetObject
		Headers: r.Header.Clone(),
		Data:    b,
	}
	h.m.Unlock()

	_, _ = w.Write([]byte(""))
}

// https://docs.aws.amazon.com/AmazonS3/latest/API/API_GetObject.html
func (h *AWSAPIHandler) s3GetObject(r *http.Request, w http.ResponseWriter) {
	h.m.Lock()
	defer h.m.Unlock()

	key := strings.TrimPrefix(r.URL.Path, h.PathPrefix)

	f, ok := h.S3Bucket[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(""))
		return
	}

	for key, values := range f.Headers {
		// From what I can tell S3 just returns these headers as they were used in PutObject
		if strings.HasPrefix(key, "Content-") || strings.HasPrefix(key, "X-Amz-") {
			for _, v := range values {
				w.Header().Add(key, v)
			}
		}
	}

	_, _ = w.Write(f.Data)
}

// https://docs.aws.amazon.com/AWSSimpleQueueService/latest/APIReference/API_SendMessage.html
func (h *AWSAPIHandler) sqsSendMessage(r *http.Request, w http.ResponseWriter) {
	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil)
		return
	}

	input := sqs.SendMessageInput{}
	err = json.Unmarshal(requestBody, &input)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, nil)
		return
	}

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{
				MessageId: "test-message",
				Body:      *input.MessageBody,
			},
		},
	}

	eh := h.OnSQSEvent[*input.QueueUrl]

	// handle any delays set on the messages
	if input.DelaySeconds > 0 {
		time.Sleep(time.Duration(input.DelaySeconds) * time.Second)
	}
	eh(event)

	writeJSON(w, http.StatusOK, nil)
}
