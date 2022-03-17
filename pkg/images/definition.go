/*

Copyright (C) 2019-2022  Daniele Rondina <geaaru@sabayonlinux.org>
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
	"regexp"
	"strconv"
	"time"
)

// Code imported by distrobuilder project and avoid injection of the project
// in the dependencies.

// ImageTarget represents the image target.
type ImageTarget int

// Filter represents a filter.
type Filter interface {
	GetReleases() []string
	GetArchitectures() []string
	GetVariants() []string
	GetTypes() []string
}

// A DefinitionFilter defines filters for various actions.
type DefinitionFilter struct {
	Releases      []string `yaml:"releases,omitempty"`
	Architectures []string `yaml:"architectures,omitempty"`
	Variants      []string `yaml:"variants,omitempty"`
	Types         []string `yaml:"types,omitempty"`
}

// GetReleases returns a list of releases.
func (d *DefinitionFilter) GetReleases() []string {
	return d.Releases
}

// GetArchitectures returns a list of architectures.
func (d *DefinitionFilter) GetArchitectures() []string {
	return d.Architectures
}

// GetVariants returns a list of variants.
func (d *DefinitionFilter) GetVariants() []string {
	return d.Variants
}

// GetTypes returns a list of types.
func (d *DefinitionFilter) GetTypes() []string {
	return d.Types
}

// A DefinitionPackagesSet is a set of packages which are to be installed
// or removed.
type DefinitionPackagesSet struct {
	DefinitionFilter `yaml:",inline"`
	Packages         []string `yaml:"packages"`
	Action           string   `yaml:"action"`
	Early            bool     `yaml:"early,omitempty"`
	Flags            []string `yaml:"flags,omitempty"`
}

// A DefinitionPackagesRepository contains data of a specific repository
type DefinitionPackagesRepository struct {
	DefinitionFilter `yaml:",inline"`
	Name             string `yaml:"name"`           // Name of the repository
	URL              string `yaml:"url"`            // URL (may differ based on manager)
	Type             string `yaml:"type,omitempty"` // For distros that have more than one repository manager
	Key              string `yaml:"key,omitempty"`  // GPG armored keyring
}

// CustomManagerCmd represents a command for a custom manager.
type CustomManagerCmd struct {
	Command string   `yaml:"cmd"`
	Flags   []string `yaml:"flags,omitempty"`
}

// DefinitionPackagesCustomManager represents a custom package manager.
type DefinitionPackagesCustomManager struct {
	Clean   CustomManagerCmd `yaml:"clean"`
	Install CustomManagerCmd `yaml:"install"`
	Remove  CustomManagerCmd `yaml:"remove"`
	Refresh CustomManagerCmd `yaml:"refresh"`
	Update  CustomManagerCmd `yaml:"update"`
	Flags   []string         `yaml:"flags,omitempty"`
}

// A DefinitionPackages represents a package handler.
type DefinitionPackages struct {
	Manager       string                           `yaml:"manager,omitempty"`
	CustomManager *DefinitionPackagesCustomManager `yaml:"custom_manager,omitempty"`
	Update        bool                             `yaml:"update,omitempty"`
	Cleanup       bool                             `yaml:"cleanup,omitempty"`
	Sets          []DefinitionPackagesSet          `yaml:"sets,omitempty"`
	Repositories  []DefinitionPackagesRepository   `yaml:"repositories,omitempty"`
}

// A DefinitionImage represents the image.
type DefinitionImage struct {
	Description  string `yaml:"description"`
	Distribution string `yaml:"distribution"`
	Release      string `yaml:"release,omitempty"`
	Architecture string `yaml:"architecture,omitempty"`
	Expiry       string `yaml:"expiry,omitempty"`
	Variant      string `yaml:"variant,omitempty"`
	Name         string `yaml:"name,omitempty"`
	Serial       string `yaml:"serial,omitempty"`

	// Internal fields (YAML input ignored)
	ArchitectureMapped      string `yaml:"architecture_mapped,omitempty"`
	ArchitectureKernel      string `yaml:"architecture_kernel,omitempty"`
	ArchitecturePersonality string `yaml:"architecture_personality,omitempty"`
}

// A DefinitionSource specifies the download type and location
type DefinitionSource struct {
	Downloader       string   `yaml:"downloader"`
	URL              string   `yaml:"url,omitempty"`
	Keys             []string `yaml:"keys,omitempty"`
	Keyserver        string   `yaml:"keyserver,omitempty"`
	Variant          string   `yaml:"variant,omitempty"`
	Suite            string   `yaml:"suite,omitempty"`
	SameAs           string   `yaml:"same_as,omitempty"`
	SkipVerification bool     `yaml:"skip_verification,omitempty"`
}

// A DefinitionTargetLXCConfig represents the config part of the metadata.
type DefinitionTargetLXCConfig struct {
	Type    string `yaml:"type"`
	Before  uint   `yaml:"before,omitempty"`
	After   uint   `yaml:"after,omitempty"`
	Content string `yaml:"content"`
}

// A DefinitionTargetLXC represents LXC specific files as part of the metadata.
type DefinitionTargetLXC struct {
	CreateMessage string                      `yaml:"create_message,omitempty"`
	Config        []DefinitionTargetLXCConfig `yaml:"config,omitempty"`
}

// DefinitionTargetLXDVM represents LXD VM specific options.
type DefinitionTargetLXDVM struct {
	Size       uint64 `yaml:"size,omitempty"`
	Filesystem string `yaml:"filesystem,omitempty"`
}

// DefinitionTargetLXD represents LXD specific options.
type DefinitionTargetLXD struct {
	VM DefinitionTargetLXDVM `yaml:"vm,omitempty"`
}

// A DefinitionTarget specifies target dependent files.
type DefinitionTarget struct {
	LXC  DefinitionTargetLXC `yaml:"lxc,omitempty"`
	LXD  DefinitionTargetLXD `yaml:"lxd,omitempty"`
	Type string              // This field is internal only and used only for simplicity.
}

// A DefinitionFile represents a file which is to be created inside to chroot.
type DefinitionFile struct {
	DefinitionFilter `yaml:",inline"`
	Generator        string                 `yaml:"generator"`
	Path             string                 `yaml:"path,omitempty"`
	Content          string                 `yaml:"content,omitempty"`
	Name             string                 `yaml:"name,omitempty"`
	Template         DefinitionFileTemplate `yaml:"template,omitempty"`
	Templated        bool                   `yaml:"templated,omitempty"`
	Mode             string                 `yaml:"mode,omitempty"`
	GID              string                 `yaml:"gid,omitempty"`
	UID              string                 `yaml:"uid,omitempty"`
	Pongo            bool                   `yaml:"pongo,omitempty"`
	Source           string                 `yaml:"source,omitempty"`
}

// A DefinitionFileTemplate represents the settings used by generators
type DefinitionFileTemplate struct {
	Properties map[string]string `yaml:"properties,omitempty"`
	When       []string          `yaml:"when,omitempty"`
}

// A DefinitionAction specifies a custom action (script) which is to be run after
// a certain action.
type DefinitionAction struct {
	DefinitionFilter `yaml:",inline"`
	Trigger          string `yaml:"trigger"`
	Action           string `yaml:"action"`
}

// DefinitionMappings defines custom mappings.
type DefinitionMappings struct {
	Architectures   map[string]string `yaml:"architectures,omitempty"`
	ArchitectureMap string            `yaml:"architecture_map,omitempty"`
}

// DefinitionEnvVars defines custom environment variables.
type DefinitionEnvVars struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

// DefinitionEnv represents the config part of the environment section.
type DefinitionEnv struct {
	ClearDefaults bool                `yaml:"clear_defaults,omitempty"`
	EnvVariables  []DefinitionEnvVars `yaml:"variables,omitempty"`
}

// A Definition a definition.
type Definition struct {
	Image       DefinitionImage    `yaml:"image"`
	Source      DefinitionSource   `yaml:"source"`
	Targets     DefinitionTarget   `yaml:"targets,omitempty"`
	Files       []DefinitionFile   `yaml:"files,omitempty"`
	Packages    DefinitionPackages `yaml:"packages,omitempty"`
	Actions     []DefinitionAction `yaml:"actions,omitempty"`
	Mappings    DefinitionMappings `yaml:"mappings,omitempty"`
	Environment DefinitionEnv      `yaml:"environment,omitempty"`
}

//GetExpiryDate returns an expiry date based on the creationDate and format.
func GetExpiryDate(creationDate time.Time, format string) time.Time {
	regex := regexp.MustCompile(`(?:(\d+)(s|m|h|d|w))*`)
	expiryDate := creationDate

	for _, match := range regex.FindAllStringSubmatch(format, -1) {
		// Ignore empty matches
		if match[0] == "" {
			continue
		}

		var duration time.Duration

		switch match[2] {
		case "s":
			duration = time.Second
		case "m":
			duration = time.Minute
		case "h":
			duration = time.Hour
		case "d":
			duration = 24 * time.Hour
		case "w":
			duration = 7 * 24 * time.Hour
		}

		// Ignore any error since it will be an integer.
		value, _ := strconv.Atoi(match[1])
		expiryDate = expiryDate.Add(time.Duration(value) * duration)
	}

	return expiryDate
}
