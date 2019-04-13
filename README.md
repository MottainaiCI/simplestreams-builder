# Simplestreams Builder

LXD permits to retrieve images in two different ways:
through LXD protocol or simplestreams protocol.

LXD protocol is a TLS based protocol that isn't so easy
to use for exposing images through the guest mirrors chain.

Simplestreams protocol instead is an HTTP based protocol
that could be used to share images through the guest mirrors nodes.

**simplestreams-builder** module try to simplify process
for creating the files and directories tree for exposing
LXC/LXD images over simplestreams protocol.

Images shared through Simplestreams protocol could be bigger
and not so easy to process to a single Mottainai namespace.
For this reason, `simplestreams-builder` is been created
with a scalable vision in mind where the same remote node
could expose different images through different
Mottainai namespaces.

```bash

$# simplestreams-builder --help
Simplestreams Builder
Copyright (c) 2019 Mottainai

Mottainai - LXC/LXD Simplestreams Tree Builder

Usage:
   [command]

Available Commands:
  build-images-file       Build images.json file of the tree
  build-index             Build index.json file of the tree
  build-product           Build product image and purge old images.
  build-versions-manifest Build ssb.json file of one product
  help                    Help about any command
  print                   Show configuration params

Flags:
  -c, --config string       SimpleStreams Builder configuration file
  -h, --help                help for this command
  -t, --target-dir string   Target dir of operations.
      --version             version for this command

Use " [command] --help" for more information about a command.

```

## Getting Started

### Prepare tree.yml

**simplestreams-builder** requires a YAML configuration file that describe
what products must be built and with what properties.

A completed example is available [here](https://github.com/Sabayon/sbi-tasks/blob/master/lxd/tree-images.yaml).

In general, inside tree.yml file you can define:

  * **prefix**: it used to customize path where share LXD/LXC images.
    Usually, path inside images.json and index.json are relative paths
    so we could leave this option with an empty string.

  * **images\_path**: path where create the files images.json and index.json. Currently, LXD implementation require *streams/v1*.
    Usually this option is not modified by users.

  * **datatype**: define datatype of products. No changes are needed.

  * **format**: define the version of simplestreams schema version for
    products. Normally this field is left unchanged.

 * **products**: contains list of products to build.

Every product contains:

  * **name**: Name used for identify image. Must be unique.

  * **arch**: Architecture of the image to build.

  * **release**: Release of the image to build

  * **os**: OS of the image to build

  * **release_title**: Title of the image

  * **directory**: directory of the tree where build image, find/create ssb.json file

  * **prefix_path**: this option must contain the URL where retrieve
    ssb.json file of the product. Normally this option is used with Mottainai to use different
    namespaces for every product. If build of the tree is done directly to a specific
    directory this option must be empty.

  * **days**: Number of images to maintains on the tree. (I consider to have only one image for a day).

  * **aliases**: Aliases of the image to build.

### Build images

For every images it's needed prepare a YAML file to use with [distrobuiler](https://github.com/lxc/distrobuilder).
By default *simplestreams-builder* search for a file with name `image.yaml` but could be
defined with `-i|--image-filename` option.



```bash

$# simplestreams-builder build-product --help
Build product image and purge old images.

Usage:
   build-product <name> [flags]

Flags:
  -h, --help                    help for build-product
  -i, --image-filename string   Name of the file used by distrobuilder.
                                Default is image.yaml. (default "image.yaml")
      --skip-lxc                Skip build of LXC image
      --skip-lxd                Skip build of LXD image
      --skip-purge              Skip purge of old images.
  -s, --source-dir string       Directory where retrieve images manifests.
                                If not set source-dir then target-dir is used.

Global Flags:
  -c, --config string       SimpleStreams Builder configuration file
  -t, --target-dir string   Target dir of operations.

```

An example of Mottainai task that build images of a specific product is available
[here](https://github.com/Sabayon/sbi-tasks/blob/master/lxd/sabayon-builder/task.yaml#17).

For every execution of sub-command `build-product` is created an image under a directory
that has the name in the format `YYYYMMDD_HH24:MM`.

After that image is been built it's needed create `ssb.json` file used for create
`images.json` file required by Simplestreams Protocol.
[Here](https://github.com/Sabayon/sbi-tasks/blob/master/lxd/sabayon-builder/task.yaml#L18)
and example.

```bash
$# simplestreams-builder build-versions-manifest --help
Build ssb.json file of one product

Usage:
   build-versions-manifest [flags]

Flags:
  -h, --help                help for build-versions-manifest
  -p, --product string      Name of the product to elaborate.
  -s, --source-dir string   Directory where retrieve images for Manifest.
      --stdout              Print ssb.json to stdout

Global Flags:
  -c, --config string       SimpleStreams Builder configuration file
  -t, --target-dir string   Target dir of operations.
```

### Create images.json file

When all images are ready it's needed call `build-images-file` command for create images.json
file.

An example is available [here](https://github.com/Sabayon/sbi-tasks/blob/master/lxd/build-index.yaml#L19).

```bash
simplestreams-builder build-images-file --help
Build images.json file of the tree

Usage:
   build-images-file [flags]

Flags:
  -h, --help                help for build-images-file
  -s, --source-dir string   Directory where retrieve images manifests.
                            If not set source-dir then target-dir is used.
      --stdout              Print index.json to stdout

Global Flags:
  -c, --config string       SimpleStreams Builder configuration file
  -t, --target-dir string   Target dir of operations.

```

### Create index.json file

The last step is create `index.json` file with command `build-index`.
This command could be execute also before images.json file creation.

An example is available [here](https://github.com/Sabayon/sbi-tasks/blob/master/lxd/build-index.yaml#L17).

```bash
simplestreams-builder build-index --help
Build index.json file of the tree

Usage:
   build-index [flags]

Flags:
  -h, --help                help for build-index
  -s, --source-dir string   Directory where retrieve images manifests.
                            If not set source-dir then target-dir is used.
      --stdout              Print index.json to stdout

Global Flags:
  -c, --config string       SimpleStreams Builder configuration file
  -t, --target-dir string   Target dir of operations.
```

## Use Simplestreams Tree over HTTP/HTTPS

### Add a remote with LXD

When all hard job is done to use shared images
with an LXD installation, it's needed only to add
the remote with `lxc` tool or through editing
of `config.yml` file.

```
  $# lxc remote add my-mottainai-remote \
    https://mynode.mottainai.org/namespace/ssb \
    --protocol simplestreams --public
```

Or on editing `config.yml` file:

```yaml
default-remote: local
remotes:
  images:
    addr: https://mynode.mottainai.org/namespace/ssb
    public: true
    protocol: simplestreams
```

