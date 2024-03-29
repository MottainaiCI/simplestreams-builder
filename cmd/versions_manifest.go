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
	"path"
	"strings"

	"github.com/spf13/cobra"

	conf "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	images "github.com/MottainaiCI/simplestreams-builder/pkg/images"
	utils "github.com/MottainaiCI/simplestreams-builder/pkg/tools"
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
			var productDir, f string

			for _, p := range config.Products {
				if p.Name == config.Viper.Get("product") {
					ssp = &p
					break
				}
			}

			if ssp == nil {
				fmt.Println("No product found with name " + config.Viper.GetString("product"))
				os.Exit(1)
			}

			productDir = fmt.Sprintf("%s/%s", config.Viper.Get("source-dir"),
				ssp.Directory)

			manifest, err := images.BuildVersionsManifest(ssp, images.BuildVersionsManifestOptions{
				ProductDir:          productDir,
				PrefixPath:          config.Prefix,
				ImageFile:           config.Viper.GetString("product-image-file"),
				ForceExpireDuration: config.Viper.GetString("force-expire"),
			})
			utils.CheckError(err)

			if config.Viper.GetBool("stdout-manifest") {
				images.WriteVersionsManifestJson(manifest, os.Stdout)
			} else {

				if config.Viper.Get("target-dir") != "" {
					f = fmt.Sprintf("%s/ssb.json",
						strings.TrimRight(
							path.Join(config.Viper.GetString("target-dir"), ssp.Directory),
							"/",
						),
					)
				} else {
					// I write files under source dir
					f = fmt.Sprintf("%s/ssb.json", productDir)
				}

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
				err = images.WriteVersionsManifestJson(manifest, w)
				utils.CheckError(err)
				w.Flush()
			}

		},
	}

	var pflags = cmd.PersistentFlags()
	pflags.Bool("stdout", false, "Print ssb.json to stdout")
	config.Viper.BindPFlag("stdout-manifest", pflags.Lookup("stdout"))
	pflags.StringP("product", "p", "", "Name of the product to elaborate.")
	config.Viper.BindPFlag("product", pflags.Lookup("product"))
	pflags.StringP("source-dir", "s", "", "Directory where retrieve images for Manifest.")
	config.Viper.BindPFlag("source-dir", pflags.Lookup("source-dir"))
	pflags.StringP("force-expire", "e", "", "Force expire duration and ignore image file.")
	config.Viper.BindPFlag("force-expire", pflags.Lookup("force-expire"))
	pflags.StringP("product-image-file", "i", "",
		`Name of the file used by distrobuilder.
Default is image.yaml.`)
	config.Viper.BindPFlag("product-image-file", pflags.Lookup("product-image-file"))

	return cmd
}
