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

	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/glassechidna/pstore/pkg/pstore"
)

var shellCmd = &cobra.Command{
	Use:   "shell",
	Short: "shell will seed your session with environment with the ssm parameters, and will not execute a child process",
	Long: `Example:
	#!/bin/bash
	# do some stuff ...
	eval $(PSTORE_DBSTRING=MyDatabaseString pstore shell)
	echo $DBSTRING # will echo out your secret string!`,
	Run: func(cmd *cobra.Command, args []string) {
		simplePrefix := viper.GetString("prefix")
		tagPrefix := viper.GetString("tag-prefix")
		pathPrefix := viper.GetString("path-prefix")
		verbose := viper.GetBool("verbose")

		doShell(simplePrefix, tagPrefix, pathPrefix, verbose)
	},
}

func doShell(simplePrefix, tagPrefix, pathPrefix string, verbose bool) {
	pstore.Doit(simplePrefix, tagPrefix, pathPrefix, verbose, func(key, val string) {
		escaped := strings.Replace(val, "'", "\\'", -1)
		fmt.Printf("export %s='%s'\n", key, escaped)
	})
}

func init() {
	RootCmd.AddCommand(shellCmd)
}
