# CNWAN Reader

CNWAN Reader watches a service registry for changes and sends events to an
external handler for processing.

The CNWAN Reader is part of the Cloud Native SD-WAN (CNWAN) project.
Please check the [CNWAN documentation](https://github.com/CloudNativeSDWAN/cnwan-docs)
for the general project overview and architecture.
You can contact the CNWAN team at [cnwan@cisco.com](mailto:cnwan@cisco.com).

## Table of contents

* [Overview](#overview)  
* [Supported Service Registries](#supported-service-registries)  
* [Installation](#installation)
  * [Go Get](#go-get)
  * [Clone the Project](#clone-the-project)
  * [Run as a Docker Container](#run-as-a-docker-container)
* [Usage](#usage)
  * [CNWAN Adaptor](#cnwan-adaptor)
  * [Metadata Key](#metadata-key)
  * [Service Directory](#service-directory)
  * [Binary Example](#binary-example)
  * [Docker Usage](#docker-usage)
    * [Mount Service Account](#mount-service-account)
    * [Docker Example](#docker-example)
* [OpenAPI Specification](#openapi-specification)
* [Contributing](#contributing)
* [License](#license)

## Overview

The CNWAN Reader makes use of the [service discovery](https://en.wikipedia.org/wiki/Service_discovery)
pattern by connecting to a service registry and observing changes in published
services/endpoints. Detected changes are then processed and sent as events to
an *adaptor*, which can be created following the `OpenAPI` specification
included in this repository.

Please follow this readme to know more about *OpenAPI*, *Adaptors* and
*Supported Service Registries*.

## Supported Service Registries

Currently, the CNWAN Reader can discover services/endpoints published to
Google Cloud's [Service directory](https://cloud.google.com/service-directory).

In order to connect correctly, a
[service account](https://cloud.google.com/iam/docs/service-accounts) is
needed.  
To learn more about Google Cloud Service Accounts, you can also consult
[this guide](https://cloud.google.com/iam/docs/creating-managing-service-accounts).
Finally, you can read Service Directory's [documentation](https://cloud.google.com/service-directory/docs)
to know more about how it works.

## Installation

The following sections detail some of the methods available to install and run
the project. After you chose the method you prefer the most, follow
[Usage](#usage) to learn how to use the program flags regardless of the method
you chose.

### Go Get

This is the easiest and fastest way to get and run the program and is
recommended for users that just want to use the program without building
or modifying it.
It requires [Golang](https://golang.org/doc/install) to be present on the
machine.

Execute

```bash
go get -u github.com/CloudNativeSDWAN/cnwan-reader
```

to download the project to your computer.

Optionally, but very recommended, you can add it to your `$PATH`, so that you
won't have to specify its full/relative path every time.  
To do so, if you are a *Unix/Linux/Mac* user and supposing your golang folder
is `$HOME/go/` (the default one usually) run:

```bash
PATH=$PATH:$HOME/go/bin
```

and for *Windows* user and supposing your golang folder is in
`%USERPROFILE%\go`:

```powershell
set PATH=%PATH%;%USERPROFILE%\go\bin\
```

Or, still for *Windows*, you can follow
[this guide](https://www.computerhope.com/issues/ch000549.htm) using your
golang folder - usually `%USERPROFILE%\go\bin\` if you never changed it.

Now you can run the program as

```bash
cnwan-reader [...]
```

without having to mention its full/relative path every time.  
Follow [Usage](#usage) to learn how to use the program.

### Clone the project

As the previous section, this requires [Golang](https://golang.org/doc/install)
to be installed on the machine in order to run the program and is most suitable
for users that want to modify it or contribute to it.  
Run

```bash
git clone github.com/CloudNativeSDWAN/cnwan-reader
cd cnwan-reader
```

Now you need to build the program in order to use it. Although you may use
`go` commands to do so, we recommend using the included `Makefile` as this
will automate a lot of commands.  
To use the `Makefile` you need to have `Make` installed, which comes already
pre-installed if are a *Unix/Linux/Mac* user. If you are a *Windows* user,
you can download the binaries from
[this page](http://gnuwin32.sourceforge.net/packages/make.htm).

Execute

```bash
make build
```

Now you can run the program

```bash
# From the root folder
./cnwan-reader [...]

# From a different folder
path/to/cnwan-reader [...]
```

Follow [Usage](#usage) to learn how to use the program.

### Run as a Docker Container

If you wish, you can build and run the docker container out of the project.
To do so, please first follow the [Clone the Project](#clone-the-project)
section and make sure you have [Docker](https://www.docker.com/get-started)
installed:

* *Unix/Linux* users with
  [Snap](https://snapcraft.io/docs/installing-snapd):

  ```bash
  sudo snap install docker
  ```

* *MacOs* users:
  [Docker Desktop for Mac](https://hub.docker.com/editions/community/docker-ce-desktop-mac/)
* *Windows* users:
  [Docker Desktop for Windows](https://hub.docker.com/editions/community/docker-ce-desktop-windows/)

Now navigate to the root folder of the project and run:

```bash
make docker-build IMG=<repository/image-name:tag-name>
```

To avoid specifying the `IMG` parameter every time, you can modify the top
of the `Makefile` to look like this:

```Makefile
# Image URL to use all building/pushing image targets
IMG ?= <repository/image-name:tag-name>
```

Now you can build the image just as

```bash
make docker-build
```

Now you can run the program as

```bash
docker run <repository/image-name:tag-name> [...]
```

Follow the [Docker Usage](#docker-usage) section to learn how to use it.

As a final note, if you also wish to push the container to a container
registry, make sure you are correctly logged in to it.  
Most of the times, [this guide](https://docs.docker.com/engine/reference/commandline/login/)
should do it, but we encourage you to read your container registry's official
documentation to learn how to do that.  
Your image name should respect the container registry format: i.e. if you are
using [DockerHub](https://hub.docker.com/) the name of your image should be
something like `your-username/image-name:tag-name`.  
For other registries
the full repository URL should be included, i.e.
`registry.com/your-username/image-name:tag-name`.

Finally, to push it to a container registry, and supposing you have modified
the `Makefile` as described above:

```bash
make docker-push
```

## Usage

The following sections describe how to use the program flags regardless
of the installation method you chose.

[Binary Example](#binary-example) contains an example for users running the
program as a binary, i.e. when installed through `go get` or cloned.
The last section, [Docker Usage](#docker-usage) is only for users running the
program as a Docker container.

### CNWAN Adaptor

*Adaptors* are external handlers that will receive the events sent by the CNWAN
Reader and process them.

By default, CNWAN Reader sends data to `localhost:80/cnwan/events`, so it
expects adaptors to provide a server listening on `localhost:80/cnwan`.  
In case you have a different endpoint or already have a server listening on
another host:port or just prefer to use another prefix path - or none at all,
you can override this behavior with the `-adaptor-api` argument:

For example:

```bash
--adaptor-api localhost:5588/my/path
```

Events will be now sent to `localhost:5588/my/path/events`.  
As an example of no prefix path, `--adaptor-api localhost:8080` will instruct
the CNWAN Reader to send events on `localhost:8080/events` instead of
`localhost:8080/cnwan/events`. If a port is not provided, `80` will be used
as default.

Please follow [OpenAPI Specification](#openapi-specification) to learn more
about adaptors and [Example](#example) for a complete usage example that
includes a CNWAN Adaptor endpoint as well.

### Metadata Key

The CNWAN Reader only reads services that have the provided metadata key.

For example, the following flag

```bash
--metadata-key prefix/key
```

will make the program only look for services that contain `prefix/key` in their
metadata key and ignore all the others.

### Service Directory

To connect to *Google Cloud Service Directory*, you can use the
`servicedirectory` command. A region, project and service account path must be
provided as flags, like so:

```bash
servicedirectory --project my-project --region us-central1 --service-account ...

# With a shorter alias
sd --project my-project --region us-central1 --service-account ...
```

Providing the service account `JSON` file is different depending on the way you
run the project: if you are running the binary version you can simply read
[Binary Example](#binary-example) for a full example usage.  
If you are running it as a Docker container, follow
[Mount Service Account](#mount-service-account) to learn how to do that.

### Binary Example

In the following example, the CNWAN Reader watches changes in
Google Cloud Service Directory with the following requirements:

* The *allowed* services have at least the `key-name` key in their metadata
* The project is called `my-project`
* The region is `us-west2`
* Service account is placed inside `path/to/creds` folder
* The name of the service account file is `serv-acc.json`
* The endpoint of the adaptor is `http://example.com:5588/my/path`
* Interval between two watches is `10 seconds`

```bash
cnwan-reader sd \
--service-account path/to/creds/serv-acc.json \
--project my-project \
--region us-west2 \
--adaptor-api example.com:5588/my/path \
--metadata-key key-name \
--interval 10
```

### Docker usage

All the previous sections apply for Docker as well.  
As specified in [Run as a Docker Container](#run-as-a-docker-container),
the program is executed as

```bash
docker run <repository/image-name:tag-name>
```

Please read along to learn usage specific to Docker.

#### Mount Service Account

Providing the service account is exactly the same as specified in
[Service Directory](#service-directory), but in order to use the
`--service-account` flag you need to first mount the file in the container with
`-v` **before** any other flag.

With `-v` you first specify where the file is stored in your computer. Then,
after a `:`, you specify where you wish to mount that file in the container,
which is going to be the argument that `--service-account` will take.

For example: supposing the path to the service account on your computer is
`~/Desktop/cnwan-credentials/serv-acc.json` and you want to mount it in the
container as `/credentials/serv-acc.json`, the flag looks like this:

```bash
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
```

Now you can use all other flags as specified in [Usage](#usage) and,
specifically, you can use `--service-account` as
`--service-account /credentials/serv-acc.json`.

Read the next section for a full Docker example.

#### Docker Example

```bash
docker run \
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
my-image \
servicedirectory \
--project my-project \
--region us-west2 \
--adaptor-api example.com:5588 \
--metadata-key key-name \
--interval 10 \
--service-account ./credentials/serv-acc.json
```

## OpenAPI Specification

The CNWAN Reader acts as a *client*, sending detected changes in form of
events to an external handler - an *Adaptor* - for processing. Therefore, any
program interested in receiving and processing these events must generate the
*server* code starting from the OpenAPI specification, or you can just
implement the appropriate endpoint in your already existing server.

The specification, along with documentation on what you need to implement
and what data is sent by the CNWAN Reader, is included in this repository
at this [link](./api/README.md).

To learn more about OpenAPI please take a look at [this repository](https://github.com/OAI/OpenAPI-Specification).  
To generate your code, you can use the [OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator).

## Contributing

Thank you for interest in contributing to this project.  
Before starting, please make sure you know and agree to our [Code of conduct](./code-of-conduct.md).

1. Fork it
2. Download your fork  
    `git clone https://github.com/your_username/cnwan-reader && cd cnwan-reader`
3. Create your feature branch  
    `git checkout -b my-new-feature`
4. Make changes and add them  
    `git add .`
5. Commit your changes  
    `git commit -m 'Add some feature'`
6. Push to the branch  
    `git push origin my-new-feature`
7. Create new pull request to this repository

## License

CNWAN Reader is released under the Apache 2.0 license. See [LICENSE](./LICENSE)
