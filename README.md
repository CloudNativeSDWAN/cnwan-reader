# CN-WAN Reader

![GitHub](https://img.shields.io/github/license/CloudNativeSDWAN/cnwan-reader)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/CloudNativeSDWAN/cnwan-reader)
<!-- markdown-link-check-disable-next-line -->
[![Go Report Card](https://goreportcard.com/badge/github.com/CloudNativeSDWAN/cnwan-reader)](https://goreportcard.com/report/github.com/CloudNativeSDWAN/cnwan-reader)
![OpenAPI version](https://img.shields.io/badge/OpenAPI-3.0.1-green)
![GitHub Workflow Status](https://img.shields.io/github/workflow/status/CloudNativeSDWAN/cnwan-reader/Build)
![GitHub release (latest semver)](https://img.shields.io/github/v/release/CloudNativeSDWAN/cnwan-reader)
![Docker Image Version (latest semver)](https://img.shields.io/docker/v/cnwan/cnwan-reader?label=docker%20image%20version)
[![DevNet published](https://static.production.devnetcloud.com/codeexchange/assets/images/devnet-published.svg)](https://developer.cisco.com/codeexchange/github/repo/CloudNativeSDWAN/cnwan-reader)

CN-WAN Reader watches a service registry for changes and sends events to an external handler for processing.

The CN-WAN Reader is part of the Cloud Native SD-WAN (CN-WAN) project. Please check the [CN-WAN documentation](https://github.com/CloudNativeSDWAN/cnwan-docs) for the general project overview and architecture. You can contact the CN-WAN team at [cnwan@cisco.com](mailto:cnwan@cisco.com).

## Overview

The CN-WAN Reader makes use of the [service discovery](https://en.wikipedia.org/wiki/Service_discovery) pattern by connecting to a service registry and observing changes in published services/endpoints. Detected changes are then processed and sent as events to an *adaptor*, which can be created following the `OpenAPI` specification included in this repository.

Please follow this readme to know more about *OpenAPI*, *Adaptors* and *Supported Service Registries*.

## Supported Service Registries

Currently, the CN-WAN Reader can discover services/endpoints published to Google Cloud's [Service directory](https://cloud.google.com/service-directory) and AWS [Cloud Map](https://aws.amazon.com/cloud-map/).

### Google Cloud Service Directory

In order to connect correctly, a [service account](https://cloud.google.com/iam/docs/service-accounts) is needed. To learn more about Google Cloud Service Accounts, you can also consult [this guide](https://cloud.google.com/iam/docs/creating-managing-service-accounts). Finally, you can read Service Directory's [documentation](https://cloud.google.com/service-directory/docs) to know more about how it works.

Finally, please make sure your service account has *at least* role `roles/servicedirectory.viewer`. We suggest you create service account just for the CN-WAN Reader with the aforementioned role.

### AWS Cloud Map

You will need valid [credentials](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) in able to watch changes correctly.

In order to use CN-WAN Reader with Cloud Map, your IAM identity needs to have *at least* policy `AWSCloudMapReadOnlyAccess` or above.

Please note that, as of now, the reader is only able to read up to `100` services at a time and `100` instances per service. While this should more than enough for the vast majority of use-cases, if demand for supporting a higher number is there, the reader will be able to read more on next updates.

## Documentation

To learn how to install or use the program, please follow documentation provided in the [docs](./docs) directory.

You can start by reading [Installation](./docs/installation.md) or [Docker Installation](./docs/docker_installation.md) if you want to install and run the program inside a docker container.

## OpenAPI Specification

The CN-WAN Reader acts as a *client*, sending detected changes in form of events to an external handler - an *Adaptor* - for processing. Therefore, any program interested in receiving and processing these events must generate the *server* code starting from the OpenAPI specification, or you can just implement the appropriate endpoint in your already existing server.

The specification, along with documentation on what you need to implement and what data is sent by the CN-WAN Reader, is included in this repository at this [link](./api/README.md).

To learn more about OpenAPI please take a look at [this repository](https://github.com/OAI/OpenAPI-Specification). To generate your code, you can use the [OpenAPI Generator](https://github.com/OpenAPITools/openapi-generator).

## Contributing

Thank you for interest in contributing to this project. Before starting, please make sure you know and agree to our [Code of conduct](./code-of-conduct.md).

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

CN-WAN Reader is released under the Apache 2.0 license. See [LICENSE](./LICENSE)
