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
	"fmt"
	"os"

	utils "github.com/MottainaiCI/mottainai-server/pkg/utils"
	"github.com/spf13/cobra"

	conf "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	images "github.com/MottainaiCI/simplestreams-builder/pkg/images"
)

func newBuildVersionsManifestCommand(config *conf.BuilderTreeConfig) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "build-versions-manifest",
		Short: "Build ssb.json file of one product",
		Args:  cobra.NoArgs,
		PreRun: func(cmd *cobra.Command, args []string) {
			if config.Viper.Get("product") == "" {
				fmt.Println("No product choice.")
				os.Exit(1)
			}
			if config.Viper.Get("source-dir") == "" {
				fmt.Println("Missing source-dir option.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var ssp *conf.SimpleStreamsProduct = nil
			var productDir string

			for _, p := range config.Products {
				if p.Name == config.Viper.Get("product") {
					ssp = &p
					break
				}
			}

			if ssp == nil {
				fmt.Println("No product found with name " + config.Viper.Get("product").(string))
				os.Exit(1)
			}

			productDir = fmt.Sprintf("%s/%s", config.Viper.Get("source-dir"),
				ssp.Directory)

			manifest, err := images.BuildVersionsManifest(ssp, productDir, config.Prefix)
			utils.CheckError(err)

			if config.Viper.GetBool("stdout") {
				images.WriteVersionsManifestJson(manifest, os.Stdout)
			}
		},
	}

	var pflags = cmd.PersistentFlags()
	pflags.Bool("stdout", false, "Print ssb.json to stdout")
	config.Viper.BindPFlag("stdout", pflags.Lookup("stdout"))
	pflags.StringP("product", "p", "", "Name of the product to elaborate.")
	config.Viper.BindPFlag("product", pflags.Lookup("product"))
	pflags.StringP("source-dir", "s", "", "Directory where retrieve images for Manifest.")
	config.Viper.BindPFlag("source-dir", pflags.Lookup("source-dir"))

	return cmd
}
