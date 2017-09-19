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
	"github.com/spf13/cobra"
	"github.com/glassechidna/pstore/common"
	"os"
	"github.com/spf13/viper"
)

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		simplePrefix := viper.GetString("prefix")
		tagPrefix := viper.GetString("tag-prefix")
		verbose := viper.GetBool("verbose")

		doExec(simplePrefix, tagPrefix, verbose, args)
	},
}

func doExec(simplePrefix, tagPrefix string, verbose bool, args []string) {
	common.Doit(simplePrefix, tagPrefix, verbose, func(key, val string) {
		os.Setenv(key, val)
	})

	common.ExecCommand(args)
}

func init() {
	RootCmd.AddCommand(execCmd)

	execCmd.PersistentFlags().String("prefix", "PSTORE_", "")
	execCmd.PersistentFlags().String("tag-prefix", "PSTORETAG_", "")
	execCmd.PersistentFlags().Bool("verbose", false, "")
	
	viper.BindPFlags(execCmd.PersistentFlags())
}
