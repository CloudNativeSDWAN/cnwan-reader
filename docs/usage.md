# Usage

This sections describes how to use the CN-WAN Reader, i.e. its flags.

It applies to whatever installation method you chose, but if you are running the program in a docker container, you should also follow [Docker Usage](./docker_usage.md) after reading this.

You should prefix `path/to/cnwan-reader` before these commands, or `docker run image` if you are running the program inside a docker container.

## Table of Contents

* [CN-WAN Adaptor](#cnwan-adaptor)
* [Metadata Key](#metadata-key)
* [Service registries](#service-registries)
  * [Google Cloud Service Directory](#google-cloud-service-directory)
  * [AWS Cloud Map](#aws-cloud-map)
* [Configration File](#configuration-file)
* [Examples](#examples)
  * [With Service Directory](#with-service-directory)
  * [With Cloud Map](#with-cloud-map)

## CN-WAN Adaptor

*Adaptors* are external handlers that will receive the events sent by the CN-WAN Reader and process them.

By default, CN-WAN Reader sends data to `localhost:80/cnwan/events`, so it expects adaptors to provide a server listening on `localhost:80/cnwan`. In case you have a different endpoint or already have a server listening on another host:port or just prefer to use another prefix path - or none at all, you can override this behavior with the `-adaptor-api` argument:

For example:

```bash
--adaptor-api localhost:5588/my/path
```

Events will be now sent to `localhost:5588/my/path/events`. As an example of no prefix path, `--adaptor-api localhost:8080` will instruct the CN-WAN Reader to send events on `localhost:8080/events` instead of `localhost:8080/cnwan/events`. If a port is not provided, `80` will be used as default.

Please follow [OpenAPI Specification](../README.md#openapi-specification) to learn more about adaptors and [Example](#example) for a complete usage example that includes a CN-WAN Adaptor endpoint as well.

## Metadata Key

The CN-WAN Reader only reads services that have the provided metadata key.

For example, the following flag

```bash
--metadata-key cnwan.io/traffic-profile
```

will make the program only look for services whose metadata contain `cnwan.io/traffic-profile` and ignore all services that don't have it. Please note that it will only look for the *key* and will not do any type of filtering on the value, as this job is performed by the CN-WAN Adaptor or whomever is in charge of handling the values.

## Service registries

### Google Cloud Service Directory

To connect to *Google Cloud Service Directory*, you can use the `servicedirectory` command. A region, project and service account path must be provided as flags, like so:

```bash
servicedirectory --project my-project --region us-central1 --service-account ...

# With a shorter alias
sd --project my-project --region us-central1 --service-account ...
```

Providing the service account `JSON` file is different depending on the way you run the project:

* if you are running the binary version you can simply read [Binary Example](#binary-example) for a full example usage.

* if you are running it as a Docker container, follow [Docker Usage](./docker_usage.md), as it needs some additional steps.

Finally, please make sure your service account has *at least* role `roles/servicedirectory.viewer`. We suggest you create service account just for the CN-WAN Reader with the aforementioned role.

**NOTE**: `servicedirectory` command will be moved to `poll` command soon, so the full command will be `cnwan-reader poll servicedirectory [...]`.

### AWS Cloud Map

To use *AWS Cloud Map*, you need to run *CN-WAN Reader* with the `poll` command like `cnwan-reader poll cloudmap [...]`.

You will need to provide a region to look for with `--region` and, optionally, a path where your credentials are with `--credentials-path`. If you have the [aws cli](https://aws.amazon.com/cli/) installed then you don't need to set this flag, as they should reside in `$HOME/.aws/credentials` in Unix-based systems or `%UserProfile%/.aws/credentials` in Windows ones, unless you want to use other credentials.

In order to use CN-WAN Reader with Cloud Map, your IAM identity needs to have *at least* policy `AWSCloudMapReadOnlyAccess` or above.

For more information about AWS credentials, you may take a look at aws' [documentation](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-files.html) about this topic.

## Configuration File

Optionally, a configuration file can be used, which can be used by providing its path with `--conf`. A [configuration model](../examples/config/config.yaml) is there for you on `examples/config`.

The fields in the YAML file map to each CLI flag specified in the sections above and therefore you won't need to include them if you want to use the default value, i.e. if `pollInterval` is not there, then the default value `5` will be used, as specified in `--help`.

In the provided yaml example, we entered `example.com` to specify that the adaptor is not running in the same machine as the reader, and that, if not present, the value for `host` will be `localhost` and `80` for port. If the latter case applies to you, you can just go ahead and omit `adaptor` field entirely: here the fields are complete to show you a full example with all present fields.

`metadataKeys` is a list of metadata keys that need to be watched for, ignoring the oned that don't have them, although keep in mind that, as of now, only one is supported: if you write multiple metadata keys to watch, only the first one will be kept.

Under `serviceRegistry` you will need to specify the service registry that you want to be polled/watched.

Finally, remember that CLI flags will **override** any options defined in the configuration file: for example, if your configuration file includes `pollInterval: 25` but launch the program with `--interval 50`, the former will be completely ignored.

## Examples

### With Service Directory

In the following example, the CN-WAN Reader watches changes in Google Cloud Service Directory with the following requirements:

* The *allowed* services have at least the `cnwan.io/traffic-profile` key in their metadata
* The project is called `my-project`
* The region is `us-west2`
* Service account is placed inside `path/to/creds` folder
* The name of the service account file is `serv-acc.json`
* The endpoint of the adaptor is the default one (`http://localhost:80/cnwan/events`). In such a case there is no need to use the `--adaptor-api` flag, but here it is included for clarity.
* Interval between two watches is `10 seconds`

```bash
cnwan-reader sd \
--service-account /path/to/the/service-account.json \
--project my-project \
--region us-west2 \
--metadata-key cnwan.io/traffic-profile \
--adaptor-api localhost/cnwan/events \
--interval 10
```

You can also use a configuration file to do that. Set the configuration file as this:

```yaml
adaptor: localhost:80/cnwan
metadataKeys:
  - cnwan.io/traffic-profile
serviceRegistry:
  gcpServiceDirectory:
    pollInterval: 10
    region: us-west2
    projectID: my-project
    serviceAccountPath: /path/to/the/service-account.json
```

Execute the following command:

```bash
cnwan-reader servicedirectory --conf /path/to/configuration/file.yaml
```

or just:

```bash
cnwan-reader --conf /path/to/configuration/file.yaml
```

### With Cloud Map

In the following example, the CN-WAN Reader watches changes in AWS Cloud Map with the following requirements:

* The *allowed* services have at least the `cnwan.io/traffic-profile` key in their metadata
* The region is `us-west-2`
* The credentials file is in `$HOME/.aws/credentials`, so there is no need to use the `--credentials-path` flag.
* The endpoint of the adaptor is the default one (`http://localhost:80/cnwan/events`). In such a case there is no need to use the `--adaptor-api` flag, but here it is included for clarity.
* Interval between two watches is `10 seconds`

The command below assumes you want to use the default *aws profile*: run the following command prior to running cnwan reader if you want to use another one:

```bash
export AWS_PROFILE=your_profile
```

```bash
cnwan-reader poll cloudmap \
--region us-west-2 \
--metadata-keys cnwan.io/traffic-profile \
--adaptor-api localhost/cnwan/events \
--interval 10
```

You can also use a configuration file to do that. Set the configuration file as:

```yaml
...
serviceRegistry:
  awsCloudMap:
    region: us-west-2
    pollInterval: 10
```

The file has been truncated with just the relevant parts, follow the sections above or the previous example to get the rest of the file.

Execute the following command:

```bash
cnwan-reader poll cloudmap --conf /path/to/configuration/file.yaml
```

or just:

```bash
cnwan-reader --conf /path/to/configuration/file.yaml
```
