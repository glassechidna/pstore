// Copyright © 2017 Aidan Steele <aidan.steele@glassechidna.com.au>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"

	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"sort"
)

var showCmd = &cobra.Command{
	Use:   "show <path>",
	Short: "Prints all parameters under a prefix",
	Long: `
Prints all parameters under a given prefix. Defaults to human-friendly format,
pass -j if you'd rather JSON output.
`,
	Run: func(cmd *cobra.Command, args []string) {
		jsonFormat, _ := cmd.PersistentFlags().GetBool("json")
		path := args[0]
		show(path, jsonFormat)
	},
}

func getAllParameters(sess *session.Session, path string) []*ssm.Parameter {
	api := ssm.New(sess)

	params := []*ssm.Parameter{}

	api.GetParametersByPathPages(&ssm.GetParametersByPathInput{
		Path:           &path,
		Recursive:      aws.Bool(true),
		WithDecryption: aws.Bool(true),
	}, func(page *ssm.GetParametersByPathOutput, lastPage bool) bool {
		params = append(params, page.Parameters...)
		return !lastPage
	})

	sort.Sort(byName(params))
	return params
}

func show(path string, jsonFormat bool) {
	sess := session.Must(session.NewSession())
	params := getAllParameters(sess, path)

	if jsonFormat {
		printJson(params, path)
	} else {
		printFriendly(params, path)
	}
}

func printFriendly(params []*ssm.Parameter, path string) {
	longest := longestName(params)
	padding := longest - len(faint("%s", path))

	secret := color.New(color.FgRed, color.Bold).SprintfFunc()

	for _, param := range params {
		prefix, rest := (*param.Name)[:len(path)], (*param.Name)[len(path):]
		value := *param.Value

		if *param.Type == ssm.ParameterTypeSecureString {
			value = secret("%s", value)
		}

		fmt.Printf("%s%-*s : %s\n", faint("%s", prefix), padding, rest, value)
	}
}

func printJson(params []*ssm.Parameter, path string) {
	dict := map[string]string{}

	for _, param := range params {
		dict[*param.Name] = *param.Value
	}

	bytes, _ := json.MarshalIndent(dict, "", "  ")
	fmt.Println(string(bytes))
}

type byName []*ssm.Parameter

func (s byName) Len() int {
	return len(s)
}
func (s byName) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s byName) Less(i, j int) bool {
	return *s[i].Name < *s[j].Name
}

func longestName(params []*ssm.Parameter) int {
	longest := 0

	for _, param := range params {
		plen := len(faint("%s", *param.Name))
		if plen > longest {
			longest = plen
		}
	}

	return longest
}

var faint func(format string, a ...interface{}) string

func init() {
	RootCmd.AddCommand(showCmd)
	showCmd.PersistentFlags().BoolP("json", "j", false, "Emit JSON instead of table")

	faint = color.New(color.Faint).SprintfFunc()
}
