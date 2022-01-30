module github.com/MottainaiCI/simplestreams-builder

go 1.16

replace howett.net/plist v0.0.0-20181124034731-591f970eefbb => github.com/DHowett/go-plist v0.0.0-20181124034731-591f970eefbb

replace github.com/juju/clock v0.0.0-20190205081909-9c5c9712527c => github.com/rogpeppe/clock v0.0.0-20190514193443-f0bda0cd88c6

require (
	github.com/MottainaiCI/mottainai-server v0.1.2
	github.com/fsouza/go-dockerclient v1.6.6 // indirect
	github.com/jaypipes/ghw v0.6.1 // indirect
	github.com/lxc/distrobuilder v0.0.0-20190705160854-7540ba58ac15
	github.com/lxc/lxd v0.0.0-20211001020858-71fe94be1e89
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/cobra v1.3.0
	github.com/spf13/viper v1.10.1
	golang.org/x/sys v0.0.0-20220128215802-99c3d69c2c27 // indirect
	gopkg.in/flosch/pongo2.v3 v3.0.0-20141028000813-5e81b817a0c4 // indirect
	gopkg.in/ini.v1 v1.66.3 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
)
