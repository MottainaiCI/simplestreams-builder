/*
Copyright (C) 2019-2023  Daniele Rondina <geaaru@funtoo.org>
Credits goes also to Gogs authors, some code portions and re-implemented design
are also coming from the Gogs project, which is using the go-macaron framework
and was really source of ispiration. Kudos to them!

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program. If not, see <http://www.gnu.org/licenses/>.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	conf "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	utils "github.com/MottainaiCI/simplestreams-builder/pkg/tools"
)

const (
	cliName = `Simplestreams Builder
Copyright (c) 2019-2023 Mottainai

Mottainai - LXC/LXD/Incus Simplestreams Tree Builder`

	SSB_VERSION = `0.3.1`
)

var (
	BuildTime   string
	BuildCommit string
)

func initConfig(config *conf.BuilderTreeConfig) {
	// Set env variable
	config.Viper.SetEnvPrefix(conf.SSB_ENV_PREFIX)
	config.Viper.BindEnv("config")
	config.Viper.SetDefault("config", "")

	config.Viper.AutomaticEnv()

	config.Viper.SetTypeByDefaultValue(true)
}

func initCommand(rootCmd *cobra.Command, config *conf.BuilderTreeConfig) {
	var pflags = rootCmd.PersistentFlags()

	pflags.StringP("config", "c", "", "SimpleStreams Builder configuration file")
	pflags.StringP("target-dir", "t", "", "Target dir of operations.")
	pflags.StringP("apikey", "k", "", "Mottainai API Key")

	config.Viper.BindPFlag("config", pflags.Lookup("config"))
	config.Viper.BindPFlag("target-dir", pflags.Lookup("target-dir"))
	config.Viper.BindPFlag("apikey", pflags.Lookup("apikey"))

	rootCmd.AddCommand(
		newPrintCommand(config),
		newBuildIndexCommand(config),
		newBuildVersionsManifestCommand(config),
		newBuildImagesFileCommand(config),
		newBuildProductCommand(config),
	)
}

func Execute() {
	// Create Main Instance Config object
	var config *conf.BuilderTreeConfig = conf.NewBuilderTreeConfig(nil)

	initConfig(config)

	var rootCmd = &cobra.Command{
		Short:        cliName,
		Version:      fmt.Sprintf("%s-g%s %s", SSB_VERSION, BuildCommit, BuildTime),
		Args:         cobra.OnlyValidArgs,
		SilenceUsage: true,
		PreRun: func(cmd *cobra.Command, args []string) {
			if len(args) == 0 {
				cmd.Help()
				os.Exit(0)
			}
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var err error
			var v *viper.Viper = config.Viper

			if v.Get("config") == "" {
				fmt.Println("Missing configuration file")
				os.Exit(1)
			}

			v.SetConfigType("yml")
			v.SetConfigFile(v.Get("config").(string))

			// Parse configuration file
			err = config.Unmarshal()
			utils.CheckError(err)
		},
	}

	initCommand(rootCmd, config)

	// Start command execution
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}
