# CNWAN Reader

CNWAN Reader watches a service registry for changes and sends events to an
external handler for processing.

The CNWAN Reader is part of the Cloud Native SD-WAN (CNWAN) project. Please check the [CNWAN documentation](https://github.com/CloudNativeSDWAN/cnwan-docs) for the general project overview and architecture. You can contact the CNWAN team at [cnwan@cisco.com](mailto:cnwan@cisco.com).

## Table of contents

* [Overview](#overview)  
* [Supported Service Registries](#supported-service-registries)  
* [Installing](#installation)
* [Usage](#usage)
  * [CNWAN Adaptor](#cnwan-adaptor)
  * [Service Directory](#service-directory)
  * [Example](#example)
* [OpenAPI Specification](#openapi-specification)
* [Docker](#docker)
  * [Mount Credentials](#mount-credentials)
  * [Docker Example](#docker-example)
* [Contributing](#contributing)
* [License](#license)

## Overview

The CNWAN Reader makes use of the [service discovery](https://en.wikipedia.org/wiki/Service_discovery)
pattern by connecting to a service registry and observing changes in published
services/endpoints. Detected changes are then processed and sent as events to
an *adaptor*, which can be created following the `OpenAPI` specification included in this
repository.

Please follow this readme to know more about *OpenAPI*, *Adaptors* and *Supported Services Registries*.

## Supported Service Registries

Currently, the CNWAN Reader can discover services/endpoints published to
Google Cloud's [Service directory](https://cloud.google.com/service-directory).
The project and region must be provided as arguments in the command line.  
Read [Service Directory](#service-directory) to learn how CNWAN Reader works
with Service Directory.

## Installing

To install, run the following command:

```bash
go get -d github.com/CloudNativeSDWAN/cnwan-reader
```

Or you can also clone the project by running:

```bash
git clone github.com/CloudNativeSDWAN/cnwan-reader
```

Lastly, to use this as a Docker container, please follow the [Docker](#docker)
section.

## Usage

In order to run the program, one must first build the project.  
Navigate to the project's root directory and execute

```bash
make build
```

This will generate an executable file called `cnwan-reader`. Once built,
the program can be run by executing the aforementioned executable file.

To learn more about commands and flags, please run

```bash
./cnwan-reader --help
```

### CNWAN Adaptor

*Adaptors* are external handlers that will receive the events sent by the CNWAN
Reader and process them.

By default, CNWAN Reader sends data to `localhost:80/cnwan/events`, so it
expects adaptors to provide a server listening on `localhost:80/cnwan`.  
In case you have a different endpoint or already have a server listening on
another host:port or just prefer to use another prefix path - or none at all,
you can override this behavior with the
`-adaptor-api` argument, like so:

```bash
./cnwan-reader \
servicedirectory \
--adaptor-api localhost:5588/my/path
```

With the above command, events will be sent to `localhost:5588/my/path/events`.  
As an example of no prefix path, `--adaptor-api localhost:8080` will instruct
the CNWAN Reader to send events on `localhost:8080/events` instead of
`localhost:8080/cnwan/events`. If a port is not provided, `80` will be used
as default.

Please follow [OpenAPI Specification](#openapi-specification) to learn more
about adaptors and [Example](#example) for a complete usage example that
includes a CNWAN Adaptor endpoint as well.

### Service Directory

To connect to *Google Cloud Service Directory*, you can use the
`servicedirectory` command, like so:

```bash
./cnwan-reader servicedirectory --project my-project

# With a shorter alias
./cnwan-reader sd --project my-project
```

When you use the CNWAN Reader with Service Directory you need to provide
additional arguments, i.e. the project, the region and the path of your
Google Cloud service account `JSON` file.

To learn more about Google Cloud Service Accounts, please visit
[this page](https://cloud.google.com/iam/docs/service-accounts)
or read [this guide](https://cloud.google.com/iam/docs/creating-managing-service-accounts).

```bash
./cnwan-reader sd \
--service-account ./credentials/serv-account.json \
--project my-project \
--region us-west2
```

You can read Service Directory's [documentation](https://cloud.google.com/service-directory/docs)
to learn more about it.

### Example

In the following example, the CNWAN Reader watches changes in
Google Cloud Service Directory with the following requirements:

* The *allowed* services have at least the `key-name` key in their metadata
* The project is called `my-project`
* The region is `us-west2`
* Service account is placed inside the `creds` folder
* The name of the service account file is `serv-acc.json`
* The endpoint of the adaptor is `http://example.com:5588/my/path`
* Interval between two watches is `10 seconds`

```bash
./cnwan-reader sd \
--service-account ./creds/serv-acc.json \
--project my-project \
--region us-west2 \
--adaptor-api example.com:5588/my/path \
--metadata-key key-name \
--interval 10
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

## Docker

If you prefer, you can run the docker version of the program instead of
building the executable file.

You can build your own image by first running

```bash
make docker-build IMG=example.com/your_name/image_name:tag_name
```

If you want to avoid having to write the image repository every time, you can
do so by modifying the provided `Makefile` in the project's root directory:
replace `IMG ?= <repository>` with
`IMG ?= example.com/your_name/image_name:tag_name` on the top.
You can even just name it as `IMG ?= image_name:tag_name` if you later want to
push it to your DockerHub account or if you just plan to use it locally.

Please refrain from building the container with docker commands directly, i.e.
`docker build . -t name:tag` as the provided method will
also test the program before building it.

Run the docker image with:

```bash
docker run example.com/your_name/image_name:tag_name
```

Please follow along for a usage example with docker and to learn how to
provide a valid credentials file.

### Mount Credentials

In order to work properly, CNWAN Reader needs a valid credentials file, i.e. a
Google Cloud Service Account.  
To learn more about Google Cloud Service Accounts, please visit [this page](https://cloud.google.com/iam/docs/service-accounts)
or read [this guide](https://cloud.google.com/iam/docs/creating-managing-service-accounts).

Supposing your service account is stored in `Desktop/cnwan-credentials` and
is called `serv-acc.json`, use this command to mount the file under
`/credentials` with name `serv-acc.json` in the docker image:

```bash
docker run \
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
my-image \
servicedirectory \
--project my-project \
--region us-west2 \
--service-account ./credentials/serv-acc.json
```

As you can see, the path to the credentials file in the `--service-account`
flag must match the one where the file is mounted in the docker image, i.e.
the part after the semicolon `:` in `-v`.  
Please remember that the `-v` flag must come *before* the name of the
image.

All other arguments are the same as described in [Usage](#usage).

### Docker Example

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
