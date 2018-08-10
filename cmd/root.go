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
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var RootCmd = &cobra.Command{
	Use:   "pstore",
	Short: "pstore is a tiny utility to make usage of AWS Parameter Store an absolute breeze. Simply prefix your application launch with pstore exec <yourapp> and you're up and running - in dev or prod.",
	Long: `pstore is usable out of the box. By default it looks for environment variables with a PSTORE_ prefix. For example, PSTORE_DBSTRING=MyDatabaseString asks AWS to decrypt the parameter named MyDatabaseString and stores the decrypted value in a new environment variable named DBSTRING. If there are no envvars with the PSTORE_ prefix, it's essentially a noop - so the same command can be used in local dev and in prod.

	If pstore fails to decrypt any envvars it will exit instead of launching your application.`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().String("prefix", "PSTORE_", "")
	RootCmd.PersistentFlags().String("tag-prefix", "PSTORETAG_", "")
	RootCmd.PersistentFlags().String("path-prefix", "PSTOREPATH_", "")
	RootCmd.PersistentFlags().Bool("verbose", false, "")

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.pstore.yaml)")

	viper.BindPFlags(RootCmd.PersistentFlags())
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(".pstore") // name of config file (without extension)
	viper.AddConfigPath("$HOME")   // adding home directory as first search path
	viper.AutomaticEnv()           // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
