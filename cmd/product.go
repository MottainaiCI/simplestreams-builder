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
	"path"
	"strings"

	"github.com/spf13/cobra"

	conf "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	images "github.com/MottainaiCI/simplestreams-builder/pkg/images"
	utils "github.com/MottainaiCI/simplestreams-builder/pkg/tools"
)

func newBuildProductCommand(config *conf.BuilderTreeConfig) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "build-product <name>",
		Short: "Build product image and purge old images.",
		Args:  cobra.RangeArgs(1, 1),
		PreRun: func(cmd *cobra.Command, args []string) {
			if config.Viper.Get("target-dir") == "" {
				fmt.Println("Missing target-dir option")
				os.Exit(1)
			}

			if args[0] == "" {
				fmt.Println("Missing product name.")
				os.Exit(1)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			var ssp *conf.SimpleStreamsProduct = nil
			var err error
			var name string = args[0]
			var sourceDir, imageFile string
			var opts *images.BuildProductOpts = nil

			for _, p := range config.Products {
				if p.Name == name {
					ssp = &p
					break
				}
			}

			if ssp == nil {
				fmt.Println("No product found with name " + name)
				os.Exit(1)
			}

			if config.Viper.GetString("source-dir-product") == "" {
				sourceDir = "."
			} else {
				sourceDir = config.Viper.GetString("source-dir-product")
			}

			imageFile = fmt.Sprintf("%s/%s",
				strings.TrimRight(path.Join(sourceDir, ssp.Directory), "/"),
				config.Viper.GetString("image-filename"),
			)

			if _, err = os.Stat(imageFile); os.IsNotExist(err) {
				fmt.Println(fmt.Sprintf(
					"For product %s no %s file found on path %s. I try to current path.",
					name, config.Viper.GetString("image-filename"), imageFile))

				imageFile = fmt.Sprintf("%s/%s",
					strings.TrimRight(sourceDir, "/"),
					config.Viper.GetString("image-filename"),
				)

				if _, err = os.Stat(imageFile); os.IsNotExist(err) {
					fmt.Println(fmt.Sprintf(
						"No %s file found for product %s.",
						config.Viper.GetString("image-filename"), name))
					os.Exit(1)
				}
			}

			opts = images.NewBuildProductOpts()
			opts.BuildLxc = !config.Viper.GetBool("skip-lxc")
			opts.BuildLxd = !config.Viper.GetBool("skip-lxd")
			opts.PurgeOldImages = !config.Viper.GetBool("skip-purge")

			err = images.BuildProduct(ssp,
				config.Viper.GetString("target-dir"),
				imageFile,
				opts,
			)
			utils.CheckError(err)

		},
	}

	var pflags = cmd.PersistentFlags()
	pflags.Bool("skip-purge", false, "Skip purge of old images.")
	config.Viper.BindPFlag("skip-purge", pflags.Lookup("skip-purge"))
	pflags.Bool("skip-lxc", false, "Skip build of LXC image")
	config.Viper.BindPFlag("skip-lxc", pflags.Lookup("skip-lxc"))
	pflags.Bool("skip-lxd", false, "Skip build of LXD image")
	config.Viper.BindPFlag("skip-lxd", pflags.Lookup("skip-lxd"))
	pflags.StringP("source-dir", "s", "",
		`Directory where retrieve images manifests.
If not set source-dir then target-dir is used.`)
	config.Viper.BindPFlag("source-dir-product", pflags.Lookup("source-dir"))
	pflags.StringP("image-filename", "i", "image.yaml",
		`Name of the file used by distrobuilder.
Default is image.yaml.`)
	config.Viper.BindPFlag("image-filename", pflags.Lookup("image-filename"))

	return cmd
}
