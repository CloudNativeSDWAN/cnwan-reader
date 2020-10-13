# Usage

This sections describes how to use the CNWAN Reader, i.e. its flags.

It applies to whatever installation method you chose, but if you are running
the program in a docker container, you should also follow
[Docker Usage](./docker_usage.md) after reading this.

You should pre-pend `path/to/cnwan-reader` before these commands, or
`docker run image` if you are running the program inside a docker container.

## Table of Contents

* [CNWAN Adaptor](#cnwan-adaptor)
* [Metadata Key](#metadata-key)
* [Service Directory](#service-directory)
* [Binary Example](#binary-example)

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

Please follow [OpenAPI Specification](../README.md#openapi-specification)
to learn more about adaptors and [Example](#example) for a complete usage
example that includes a CNWAN Adaptor endpoint as well.

### Metadata Key

The CNWAN Reader only reads services that have the provided metadata key.

For example, the following flag

```bash
--metadata-key cnwan.io/traffic-profile
```

will make the program only look for services whose metadata contain
`cnwan.io/traffic-profile` and ignore all services that don't have it.
Please note that it will only look for the *key* and will not do any
type of filtering on the value, as this job is performed by the CNWAN
Adaptor or whomever is in charge of handling the values.

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
run the project:

* if you are running the binary version you can simply read
[Binary Example](#binary-example) for a full example usage.

* if you are running it as a Docker container, follow
[Docker Usage](./docker_usage.md), as it needs some additional steps.

Finally, please make sure your service account has *at least* role
`roles/servicedirectory.viewer`. We suggest you create service account just for
the CNWAN Reader with the aforementioned role.

## Binary Example

In the following example, the CNWAN Reader watches changes in
Google Cloud Service Directory with the following requirements:

* The *allowed* services have at least the `cnwan.io/traffic-profile`
key in their metadata
* The project is called `my-project`
* The region is `us-west2`
* Service account is placed inside `path/to/creds` folder
* The name of the service account file is `serv-acc.json`
* The endpoint of the adaptor is the default one
(`http://localhost/cnwan/events`). In such a case there is no need to use
the `--adaptor-api` flag, but here it is included for clarity.
* Interval between two watches is `10 seconds`

```bash
cnwan-reader sd \
--service-account path/to/creds/serv-acc.json \
--project my-project \
--region us-west2 \
--metadata-key cnwan.io/traffic-profile \
--adaptor-api localhost/cnwan/events \
--interval 10
```
