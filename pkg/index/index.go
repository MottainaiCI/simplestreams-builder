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
package index

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	lxd_streams "github.com/lxc/lxd/shared/simplestreams"

	config "github.com/MottainaiCI/simplestreams-builder/pkg/config"
	images "github.com/MottainaiCI/simplestreams-builder/pkg/images"
)

func BuildIndexStruct(config *config.BuilderTreeConfig, sourceDir string) (*lxd_streams.SimpleStreamsIndex, error) {
	var ans *lxd_streams.SimpleStreamsIndex
	var products lxd_streams.SimpleStreamsIndexStream
	var ipath, prefix, ssbPath string
	var manifest *images.VersionsSSBuilderManifest
	var err error

	if config.DataType == "" {
		return nil, fmt.Errorf("Invalid datatype")
	}
	if config.ImagesPath == "" {
		return nil, fmt.Errorf("Invalid images path")
	}
	if len(config.Products) == 0 {
		return nil, fmt.Errorf("No products defined")
	}

	ipath = strings.TrimLeft(config.ImagesPath, "/")
	prefix = strings.TrimRight(config.Prefix, "/")

	products = lxd_streams.SimpleStreamsIndexStream{
		DataType: config.DataType,
		Path: fmt.Sprintf("%s/images.json",
			strings.TrimRight(path.Join(prefix, ipath), "/")),
		Format: config.Format,
	}

	for _, v := range config.Products {
		if v.Hidden {
			continue
		}

		// For every product I try to retrieve ssb.json file
		// for set versions map.
		if v.PrefixPath != "" {

			ssbPath = fmt.Sprintf("%s/%s/ssb.json",
				strings.TrimRight(v.PrefixPath, "/"),
				strings.TrimRight(v.Directory, "/"),
			)
			// Fetch and parse remote ssb.json file
			manifest, err = images.ReadVersionsManifestJsonFromUrl(ssbPath,
				config.Viper.GetString("apikey"))

		} else {
			// POST: Try to search ssb.json file under local filesystem.

			if sourceDir == "" {
				return nil, fmt.Errorf(
					"Product %s without prefix_path but source-dir is empty.",
					v.Name)
			}

			ssbPath = path.Join(sourceDir, v.Directory, "/ssb.json")
			fmt.Println("ssb Path ", ssbPath)

			if _, err := os.Stat(ssbPath); os.IsNotExist(err) {
				// POST: Ignore product
				fmt.Println(fmt.Sprintf("Product %s is skipped. ssb.json file not found.", v.Name))
				continue
			}

			// Parse ssb.json file
			manifest, err = images.ReadVersionsManifestJson(ssbPath)
		}

		if err != nil {
			fmt.Println(fmt.Sprintf("Product %s is skipped. Error on parse ssb.json: %s",
				v.Name, err.Error()))
			continue
		}

		if manifest.Name != v.Name {
			fmt.Println(fmt.Sprintf(
				"Product %s is skipped. It contains invalid ssb.json file.", v.Name))
			continue
		}

		products.Products = append(products.Products, v.Name)
	}

	ans = &lxd_streams.SimpleStreamsIndex{
		// Statically use always index 1.0 format
		Format: "index:1.0",
		Index:  make(map[string]lxd_streams.SimpleStreamsIndexStream),
	}

	ans.Index["images"] = products

	return ans, nil
}

func WriteIndexJson(index *lxd_streams.SimpleStreamsIndex, out io.Writer) error {
	enc := json.NewEncoder(out)
	return enc.Encode(index)
}
