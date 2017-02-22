package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/fatih/color"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"strings"
	"syscall"
)

const usageError = 64            // incorrect usage of "pstore"
const pstoreError = 69           // parameter store issues
const execError = 126            // cannot execute the specified command
const commandNotFoundError = 127 // cannot find the specified command

const appName = "pstore"
const appVersion = "1.0.0"

func main() {

	app := cli.NewApp()
	app.Name = appName
	app.Version = appVersion
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
				cli.BoolFlag{
					Name:  "verbose",
					Usage: "verbose mode",
				},
			},
			Action: func(c *cli.Context) {
				prefix := c.String("prefix")
				verbose := c.Bool("verbose")
				populateEnv(prefix, verbose)
				execCommand(c.Args())
			},
		},
	}

	app.Run(os.Args)

}

var userAgentHandler = request.NamedHandler{
	Name: "pstore.UserAgentHandler",
	Fn:   request.MakeAddToUserAgentHandler(appName, appVersion),
}

func populateEnv(prefix string, verbose bool) {
	sess, _ := session.NewSession()
	sess.Handlers.Build.PushBackNamed(userAgentHandler)

	client := ssm.New(sess)

	failCount := 0

	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		name := pair[0]
		value := pair[1]

		if strings.HasPrefix(name, prefix) {
			shortName := name[len(prefix):]

			names := []*string{aws.String(value)}
			// TODO: chunk array of names into blocks of 10, pass in batches to this api call
			resp, err := client.GetParameters(&ssm.GetParametersInput{Names: names, WithDecryption: aws.Bool(true)})

			if err != nil {
				panic(err)
			}

			for _, param := range resp.InvalidParameters {
				failCount++
				color.Red("✗ Failed to decrypt %s (%s)", *param, shortName)
			}

			for _, param := range resp.Parameters {
				if verbose {
					color.Green("✔ Decrypted %s︎", shortName)
				}
				decrypted := param.Value
				os.Setenv(shortName, *decrypted)
			}
		}
	}

	if failCount > 0 {
		abort(pstoreError, "Failed to decrypt some secret values")
	}
}

func abort(status int, message interface{}) {
	color.New(color.FgRed).Fprintf(os.Stderr, "ERROR: %s\n", message)
	os.Exit(status)
}

func execCommand(args []string) {
	if len(args) == 0 {
		abort(usageError, "no command specified")
	}
	commandName := args[0]
	commandPath, err := exec.LookPath(commandName)
	if err != nil {
		abort(commandNotFoundError, fmt.Sprintf("cannot find '%s'", commandName))
	}
	err = syscall.Exec(commandPath, args, os.Environ())
	if err != nil {
		abort(execError, err)
	}
}
