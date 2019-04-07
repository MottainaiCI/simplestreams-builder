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
package images

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	lxd_streams "github.com/lxc/lxd/shared/simplestreams"

	config "github.com/MottainaiCI/simplestreams-builder/pkg/config"
)

func BuildImagesFile(config *config.BuilderTreeConfig, sourceDir string) (*lxd_streams.SimpleStreamsManifest, error) {
	// NOTE: currently SimpleStreamsManifest struct doesn't contain
	//       content_id field.
	var ipath, prefix, ssbPath string
	var ans *lxd_streams.SimpleStreamsManifest
	var prodMap map[string]lxd_streams.SimpleStreamsManifestProduct
	var manifest *VersionsSSBuilderManifest
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

	if config.ImagesPath[0:1] == "/" {
		ipath = config.ImagesPath[1:len(config.ImagesPath)]
	} else {
		ipath = config.ImagesPath
	}
	if ipath[len(ipath)-1:] == "/" {
		ipath = ipath[:len(ipath)-1]
	}

	if config.Prefix[0:1] == "/" {
		if len(config.Prefix) > 1 {
			prefix = config.Prefix
		} else {
			prefix = ""
		}
	} else {
		prefix = fmt.Sprintf("/%s", config.Prefix)
	}
	if len(prefix) > 1 && prefix[len(prefix)-1:] == "/" {
		prefix = prefix[:len(prefix)-1]
	}

	prodMap = make(map[string]lxd_streams.SimpleStreamsManifestProduct)

	ans = &lxd_streams.SimpleStreamsManifest{
		// TODO: See what is format of updated field.
		DataType: config.DataType,
		Format:   config.Format,
		Products: prodMap,
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
			manifest, err = ReadVersionsManifestJsonFromUrl(ssbPath)

		} else {
			// POST: Try to search ssb.json file under local filesystem.

			if sourceDir == "" {
				return nil, fmt.Errorf(
					"Product %s without prefix_path but source-dir is empty.",
					v.Name)
			}

			ssbPath = path.Join(sourceDir, v.Directory, "/ssb.json")
			fmt.Println("Check ssb file ", ssbPath)

			if _, err := os.Stat(ssbPath); os.IsNotExist(err) {
				// POST: Ignore product
				fmt.Println(fmt.Sprintf("Product %s is skipped. ssb.json file not found.", v.Name))
				continue
			}

			// Parse ssb.json file
			manifest, err = ReadVersionsManifestJson(ssbPath)
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

		prodManifest := lxd_streams.SimpleStreamsManifestProduct{
			Architecture:    v.Architecture,
			OperatingSystem: v.OperatingSystem,
			Release:         v.Release,
			ReleaseTitle:    v.ReleaseTitle,
			Versions:        manifest.Versions,
		}

		if v.Version != "" {
			prodManifest.Version = v.Version
		}

		if len(v.Aliases) > 0 {
			prodManifest.Aliases = strings.Join(v.Aliases, ",")
		}

		ans.Products[v.Name] = prodManifest
	}

	return ans, nil
}

func WriteImagesJson(imgs *lxd_streams.SimpleStreamsManifest, out io.Writer) error {
	enc := json.NewEncoder(out)
	return enc.Encode(imgs)
}
