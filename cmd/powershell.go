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

	"github.com/glassechidna/pstore/common"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var powershellCmd = &cobra.Command{
	Use:   "powershell",
	Short: "same as shell but for windows",
	Long: `Example:
	$Env:PSTORE_DBSTRING = "MyDatabaseString"
	$Cmd = (pstore powershell mycompany-prod) | Out-String
	Invoke-Expression $Cmd
	Do-SomethingWith -DbString $DBSTRING
	`,
	Run: func(cmd *cobra.Command, args []string) {
		simplePrefix := viper.GetString("prefix")
		tagPrefix := viper.GetString("tag-prefix")
		pathPrefix := viper.GetString("path-prefix")
		verbose := viper.GetBool("verbose")

		doPowershell(simplePrefix, tagPrefix, pathPrefix, verbose)
	},
}

func doPowershell(simplePrefix, tagPrefix, pathPrefix string, verbose bool) {
	common.Doit(simplePrefix, tagPrefix, pathPrefix, verbose, func(key, val string) {
		escaped := strings.Replace(val, "\"", "\\\"", -1)
		fmt.Printf("${Env:%s}=\"%s\"\n", key, escaped)
	})
}

func init() {
	RootCmd.AddCommand(powershellCmd)
}
