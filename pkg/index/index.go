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

	lxd_streams "github.com/lxc/lxd/shared/simplestreams"

	config "github.com/MottainaiCI/simplestreams-builder/pkg/config"
)

func BuildIndexStruct(config *config.BuilderTreeConfig) (*lxd_streams.SimpleStreamsIndex, error) {
	var ans *lxd_streams.SimpleStreamsIndex
	var products lxd_streams.SimpleStreamsIndexStream
	var ipath, prefix string

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

	products = lxd_streams.SimpleStreamsIndexStream{
		DataType: config.DataType,
		Path: fmt.Sprintf("%s%s/images.json",
			prefix, ipath),
	}

	for _, v := range config.Products {
		if v.Hidden {
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
