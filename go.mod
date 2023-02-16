module github.com/MottainaiCI/simplestreams-builder

go 1.16

replace howett.net/plist v0.0.0-20181124034731-591f970eefbb => github.com/DHowett/go-plist v0.0.0-20181124034731-591f970eefbb

replace github.com/juju/clock v0.0.0-20190205081909-9c5c9712527c => github.com/rogpeppe/clock v0.0.0-20190514193443-f0bda0cd88c6

require (
	github.com/MottainaiCI/mottainai-server v0.1.2
	github.com/containerd/containerd v1.6.18 // indirect
	github.com/fsouza/go-dockerclient v1.6.6 // indirect
	github.com/jaypipes/ghw v0.6.1 // indirect
	github.com/lxc/lxd v0.0.0-20211001020858-71fe94be1e89
	github.com/magiconair/properties v1.8.6 // indirect
	github.com/moby/sys/mountinfo v0.6.0 // indirect
	github.com/spf13/afero v1.8.2 // indirect
	github.com/spf13/cobra v1.4.0
	github.com/spf13/viper v1.10.1
	gopkg.in/ini.v1 v1.66.4 // indirect
	gopkg.in/src-d/go-git.v4 v4.13.1 // indirect
)
