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
	"os"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/spf13/cobra"

	config "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	index "github.com/MottainaiCI/simplestreams-builder/pkg/index"
)

func newBuildIndexCommand(config *config.BuilderTreeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build-index",
		Short: "Build index.json file of the tree",
		Args:  cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
		},
		Run: func(cmd *cobra.Command, args []string) {
			idx, err := index.BuildIndexStruct(config)
			utils.CheckError(err)

			if config.Viper.GetBool("stdout") {
				index.WriteIndexJson(idx, os.Stdout)
			}
		},
	}

	var pflags = cmd.PersistentFlags()
	pflags.Bool("stdout", false, "Print index.json to stdout")
	config.Viper.BindPFlag("stdout", pflags.Lookup("stdout"))

	return cmd
}
