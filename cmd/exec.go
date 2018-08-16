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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/glassechidna/pstore/pkg/pstore"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "exec will seed the current environment with ssm parameters (if it exists) execute a command",
	Long: `example:
	AWS_REGION=us-east-1 PSTORE_DBSTRING=MyDatabaseString pstore exec -- 'echo val is $DBSTRING'
	val is SomeSuperSecretDbString`,

	Run: func(cmd *cobra.Command, args []string) {
		simplePrefix := viper.GetString("prefix")
		tagPrefix := viper.GetString("tag-prefix")
		pathPrefix := viper.GetString("path-prefix")
		verbose := viper.GetBool("verbose")

		doExec(simplePrefix, tagPrefix, pathPrefix, verbose, args)
	},
}

func doExec(simplePrefix, tagPrefix, pathPrefix string, verbose bool, args []string) {
	pstore.Doit(simplePrefix, tagPrefix, pathPrefix, verbose, func(key, val string) {
		os.Setenv(key, val)
	})

	pstore.ExecCommand(args)
}

func init() {
	RootCmd.AddCommand(execCmd)
}
