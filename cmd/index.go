/*
Copyright (C) 2019  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"bufio"
	"fmt"
	"os"
	path "path/filepath"

	"github.com/spf13/cobra"

	conf "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	index "github.com/MottainaiCI/simplestreams-builder/pkg/index"
	utils "github.com/MottainaiCI/simplestreams-builder/pkg/tools"
)

func newBuildIndexCommand(config *conf.BuilderTreeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build-index",
		Short: "Build index.json file of the tree",
		Args:  cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			if config.Viper.Get("target-dir") == "" && !config.Viper.GetBool("stdout") {
				fmt.Println("Missing target-dir or stdout option")
				os.Exit(1)
			} else if config.Viper.Get("target-dir") != "" && config.Viper.GetBool("stdout") {
				fmt.Println("Use target-dir or stdout option, not both.")
				os.Exit(1)
			}

		},
		Run: func(cmd *cobra.Command, args []string) {
			var f, sourceDir string
			var err error

			if config.Viper.Get("source-dir-index") != "" {
				sourceDir = config.Viper.GetString("source-dir-index")
			} else {
				sourceDir = config.Viper.GetString("target-dir")
			}
			idx, err := index.BuildIndexStruct(config, sourceDir)
			utils.CheckError(err)

			if config.Viper.GetBool("stdout") {
				index.WriteIndexJson(idx, os.Stdout)
			} else {
				// Create target directory if doesn't exist.
				// NOTE: Current LXD implementation has a static path for
				// index.json for path streams/v1 so I use always this
				// path for now.
				f = fmt.Sprintf("%s/streams/v1/index.json",
					config.Viper.Get("target-dir"))

				if _, err := os.Stat(path.Dir(f)); os.IsNotExist(err) {
					err = os.MkdirAll(path.Dir(f), 0760)
					utils.CheckError(err)
				}

				file, err := os.OpenFile(f, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
				if err != nil {
					fmt.Println("Error on create index file " + err.Error())
					os.Exit(1)
				}
				defer file.Close()

				w := bufio.NewWriter(file)
				err = index.WriteIndexJson(idx, w)
				utils.CheckError(err)
				w.Flush()
			}
		},
	}

	var pflags = cmd.PersistentFlags()
	pflags.Bool("stdout", false, "Print index.json to stdout")
	config.Viper.BindPFlag("stdout", pflags.Lookup("stdout"))
	pflags.StringP("source-dir", "s", "",
		`Directory where retrieve images manifests.
If not set source-dir then target-dir is used.`)
	config.Viper.BindPFlag("source-dir-index", pflags.Lookup("source-dir"))

	return cmd
}
