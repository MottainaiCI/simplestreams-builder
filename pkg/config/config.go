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
package config

import (
	"fmt"

	v "github.com/spf13/viper"
)

const (
	SSB_ENV_PREFIX = `SSBUILDER`
)

type SimpleStreamsProduct struct {
	Name            string   `mapstructure:"name"`
	Architecture    string   `mapstructure:"arch"`
	Release         string   `mapstructure:"release"`
	ReleaseTitle    string   `mapstructure:"release_title"`
	OperatingSystem string   `mapstructure:"os"`
	Directory       string   `mapstructure:"directory"`
	Version         string   `mapstructure:"version"`
	PrefixPath      string   `mapstructure:"prefix_path"`
	BuildScriptHook string   `mapstructure:"build_script_hook"`
	Aliases         []string `mapstructure:"aliases"`
	Hidden          bool     `mapstructure:"hidden"`
	Days            int      `mapstructure:"days"`
}

type BuilderTreeConfig struct {
	Viper *v.Viper

	Prefix     string                 `mapstructure:"prefix"`
	ImagesPath string                 `mapstructure:"images_path"`
	DataType   string                 `mapstructure:"datatype"`
	Format     string                 `mapstructure:"format"`
	Products   []SimpleStreamsProduct `mapstructure:"products"`
}

func NewBuilderTreeConfig(viper *v.Viper) *BuilderTreeConfig {
	if viper == nil {
		viper = v.New()
	}

	GenDefault(viper)
	return &BuilderTreeConfig{Viper: viper}
}

func GenDefault(viper *v.Viper) {
	viper.SetDefault("prefix", "/")
	viper.SetDefault("images_path", "streams/v1")
	viper.SetDefault("datatype", "image-downloads")
	viper.SetDefault("format", "products:1.0")
}

func (b *BuilderTreeConfig) Unmarshal() error {
	var err error

	err = b.Viper.ReadInConfig()
	if err != nil {
		return err
	}

	err = b.Viper.Unmarshal(&b)
	if err != nil {
		return err
	}

	for idx, v := range b.Products {
		if v.Days <= 0 {
			b.Products[idx].Days = 1
		}
	}

	return err
}

func (b *BuilderTreeConfig) String() string {
	var products string = ""

	for _, v := range b.Products {
		if products == "" {
			products = fmt.Sprintf("%s", v.String())
		} else {
			products = fmt.Sprintf("%s\n%s", products, v.String())
		}
	}

	var ans string = fmt.Sprintf(`
prefix: %s
images_path: %s
datatype: %s
format: %s
products:
%s
`, b.Prefix, b.ImagesPath, b.DataType,
		b.Format, products)

	return ans
}

func (p *SimpleStreamsProduct) String() string {

	var ans string = fmt.Sprintf(`
	name: %s
	arch: %s
	release: %s
	release_title: %s
	os: %s
	directory: %s
	version: %s
	prefix_path: %s
	hidden: %v
	days: %d
	build_script_hook: %s
	aliases: %s`,
		p.Name, p.Architecture, p.Release,
		p.ReleaseTitle, p.OperatingSystem,
		p.Directory, p.Version, p.PrefixPath,
		p.Hidden, p.Days, p.BuildScriptHook, p.Aliases)

	return ans
}
