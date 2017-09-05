package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"os"
	"strings"
	"context"
	"github.com/aws/aws-sdk-go/service/resourcegroupstaggingapi"
)

const usageError = 64            // incorrect usage of "pstore"
const pstoreError = 69           // parameter store issues
const execError = 126            // cannot execute the specified command
const commandNotFoundError = 127 // cannot find the specified command

const appName = "pstore"
var ApplicationVersion = "devel"

func main() {

	app := cli.NewApp()
	app.Name = appName
	app.Version = ApplicationVersion
	app.Usage = "AWS SSM Parameter Store command shim"

	app.Commands = []cli.Command{
		{
			Name:  "exec",
			Usage: "Execute a command",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "prefix",
					Usage: "environment variable prefix",
					Value: "PSTORE_",
				},
				cli.StringFlag{
					Name:  "tag-prefix",
					Usage: "tagged params environment variable prefix",
					Value: "PSTORETAG_",
				},
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "verbose mode",
				},
			},
			Action: func(c *cli.Context) {
				simplePrefix := c.String("prefix")
				tagPrefix := c.String("tag-prefix")
				verbose := c.Bool("verbose")

				doit(simplePrefix, tagPrefix, verbose, c.Args())
			},
		},
	}

	app.Run(os.Args)

}

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

func PopulateEnv(params []ParamResult, verbose bool) bool {
	anyFailed := false

	for _, param := range params {
		if !param.Success {
			color.Red("✗ Failed to decrypt %s=%s (request ID: %s)", param.ParamName, param.EnvName, param.RequestId)
			anyFailed = true
		} else if verbose {
			color.Green("✔ Decrypted %s︎=%s (request ID: %s)",  param.ParamName, param.EnvName, param.RequestId)
		}

		os.Setenv(param.EnvName, param.Value)
	}

	return !anyFailed
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

func doit(simplePrefix, tagPrefix string, verbose bool, cmd []string) {
	sess, _ := session.NewSession()
	sess.Handlers.Build.PushBackNamed(userAgentHandler)

	req := GetParamRequestFromEnv(simplePrefix, tagPrefix)

	results := GetParamsByNames(sess, req.SimpleParams)
	success := PopulateEnv(results, verbose)

	for key, val := range req.TaggedParams {
		results = GetParametersByTag(sess, key, val)
		success = PopulateEnv(results, verbose) && success
	}

	if !success {
		abort(pstoreError, "Failed to decrypt some secret values")
	}

	ExecCommand(cmd)
}

func abort(status int, message interface{}) {
	color.New(color.FgRed).Fprintf(os.Stderr, "ERROR: %s\n", message)
	os.Exit(status)
}
