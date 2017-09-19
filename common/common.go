package common

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/fatih/color"
	"os"
	"strings"
	"context"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"net/http"
	"time"
)

const usageError = 64            // incorrect usage of "pstore"
const pstoreError = 69           // parameter store issues
const execError = 126            // cannot execute the specified command
const commandNotFoundError = 127 // cannot find the specified command

const appName = "pstore"
var ApplicationVersion = "devel"

var userAgentHandler = request.NamedHandler{
	Name: "pstore.UserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler(appName, ApplicationVersion),
}

func TagValueWithKey(tags []*resourcegroupstaggingapi.Tag, key string) *string {
	for _, tag := range tags {
		if *tag.Key == key {
			return tag.Value
		}
	}

	return nil
}

func GetParametersByTag(sess *session.Session, key, value string) []ParamResult {
	api := resourcegroupstaggingapi.New(sess)
	api2 := ssm.New(sess)

	resources, _ := api.GetResources(&resourcegroupstaggingapi.GetResourcesInput{
		TagFilters: []*resourcegroupstaggingapi.TagFilter{
			{Key: &key, Values: aws.StringSlice([]string{value})},
		},
		ResourceTypeFilters: aws.StringSlice([]string{"ssm:parameter"}),
	})

	results := []ParamResult{}

	for _, r := range resources.ResourceTagMappingList {
		split := strings.SplitN(*r.ResourceARN, "parameter", 2)
		name := split[1]
		envName := TagValueWithKey(r.Tags, "pstore:name")

		if envName == nil { continue } // TODO: maybe emit logs

		requestId := ""
		input := &ssm.GetParametersInput{Names: aws.StringSlice([]string{name}), WithDecryption: aws.Bool(true)}

		resp, _ := api2.GetParametersWithContext(context.Background(), input, func(r *request.Request) {
			r.Handlers.Complete.PushBack(func(req *request.Request) {
				requestId = req.RequestID
			})
		})

		for _, p := range resp.Parameters {
			result := ParamResult{
				ParamName: *p.Name,
				EnvName: *envName,
				Value: *p.Value,
				RequestId: requestId,
				Success: true,
			}
			results = append(results, result)
		}

		for _, name := range resp.InvalidParameters {
			result := ParamResult{
				ParamName: *name,
				EnvName: *envName,
				RequestId: requestId,
				Success: false,
			}
			results = append(results, result)
		}

	}

	return results
}

type ParamResult struct {
	ParamName string
	EnvName string
	Value string
	RequestId string
	Success bool
}

func GetParamsByNames(sess *session.Session, input map[string]string) []ParamResult {
	api2 := ssm.New(sess)
	results := []ParamResult{}

	for envName, paramName := range input {
		requestId := ""

		input := &ssm.GetParameterInput{Name: &paramName, WithDecryption: aws.Bool(true)}
		resp, err := api2.GetParameterWithContext(context.Background(), input, func(r *request.Request) {
			r.Handlers.Complete.PushBack(func(req *request.Request) {
				requestId = req.RequestID
			})
		})

		if err == nil {
			result := ParamResult{
				ParamName: paramName,
				EnvName:   envName,
				Value:     *resp.Parameter.Value,
				RequestId: requestId,
				Success:   true,
			}
			results = append(results, result)
		} else {
			result := ParamResult{
				ParamName: paramName,
				EnvName: envName,
				RequestId: requestId,
				Success: false,
			}
			results = append(results, result)
		}
	}

	return results
}

type ParamsRequest struct {
	SimpleParams map[string]string
	TaggedParams map[string]string
}

func GetParamRequestFromEnv(simplePrefix, tagPrefix string) ParamsRequest {
	req := ParamsRequest{
		SimpleParams: make(map[string]string),
		TaggedParams: make(map[string]string),
	}

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		name := pair[0]
		value := pair[1]

		if strings.HasPrefix(name, simplePrefix) {
			shortName := name[len(simplePrefix):]
			req.SimpleParams[shortName] = value
		} else if strings.HasPrefix(name, tagPrefix) {
			shortName := name[len(tagPrefix):]
			req.TaggedParams[shortName] = value
		}
	}

	return req
}

func printErrors(params []ParamResult, verbose bool) bool {
	anyFailed := false

	for _, param := range params {
		if !param.Success {
			color.Red("✗ Failed to decrypt %s=%s (request ID: %s)", param.ParamName, param.EnvName, param.RequestId)
			anyFailed = true
		} else if verbose {
			color.Green("✔ Decrypted %s︎=%s (request ID: %s)",  param.ParamName, param.EnvName, param.RequestId)
		}
	}

	return !anyFailed
}

func awsRegion() string {
	config := aws.NewConfig().
		WithHTTPClient(&http.Client{Timeout: 2 * time.Second}).
		WithMaxRetries(1)

	meta := ec2metadata.New(session.Must(session.NewSession(config)))
	region := os.Getenv("AWS_REGION")

	if meta.Available() {
		regionp, err := meta.Region()
		if err == nil {
			region = regionp
		}
	}

	return region
}

func Doit(simplePrefix, tagPrefix string, verbose bool, callback func(key, value string)) {
	req := GetParamRequestFromEnv(simplePrefix, tagPrefix)
	if len(req.TaggedParams) + len(req.SimpleParams) == 0 { return }

	region := awsRegion()
	if len(region) == 0 {
		abort(usageError, "No AWS region specified. Either run on EC2 or specify AWS_REGION env var")
	}

	sess, _ := session.NewSession(aws.NewConfig().WithRegion(region))
	sess.Handlers.Build.PushBackNamed(userAgentHandler)

	results := GetParamsByNames(sess, req.SimpleParams)

	for key, val := range req.TaggedParams {
		results = append(results, GetParametersByTag(sess, key, val)...)
	}

	if !printErrors(results, verbose) {
		abort(pstoreError, "Failed to decrypt some secret values")
	}

	for _, param := range results {
		callback(param.EnvName, param.Value)
	}
}

func abort(status int, message interface{}) {
	color.New(color.FgRed).Fprintf(os.Stderr, "ERROR: %s\n", message)
	os.Exit(status)
}
