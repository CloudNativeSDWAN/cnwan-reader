# Docker usage

Before reading this, make sure you read [Usage](./usage.md) as it applies
to this guide as well.

This guide describes usage that only applies to users that run the program
inside a docker contaienr.

You should run the program as

```bash
docker run <repository/image-name:tag-name>
```

Please read along to learn usage specific to Docker.

## Table of Contents

* [Mount Service Account](#mount-service-account)
* [Mount Configuration File](#mount-configuration-file)
* [Example](#example)

## Mount Service Account

Providing the service account is exactly the same as specified in
the [Usage section](./usage.md#service-directory), but in order to use the
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

Now you can use all other flags as specified in [Usage](./usage.md) and,
specifically, you can use `--service-account` as
`--service-account /credentials/serv-acc.json`.

Read the next section for a full Docker example.

## Mount Configuration File

If you want to use a configuration file, you must follow the same principle
explained in the [section above](#mount-service-account).

So, the following flag must be provided before any other:

`-v ~/Desktop/conf/conf.yaml:/conf/conf.yaml`

and later the `--conf` must be provided accordingly as
`--conf ./conf/conf.yaml`.

## Example

This example follows the same requirements as this
[example](./usage.md#example) -- replace `cnwan/cnwan-reader` with your image
in case you have built it yourself:

```bash
docker run \
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
cnwan/cnwan-reader \
servicedirectory \
--project my-project \
--region us-west2 \
--metadata-key cnwan.io/traffic-profile \
--interval 10 \
--adaptor-api localhost/cnwan/events \
--service-account ./credentials/serv-acc.json
```

If you want to use a configuration file, you can create the same configuration
file as in the [example](./usage.md#example). For your convenience:

```yaml
adaptor:
  host: localhost
  port: 80
metadataKeys:
  - cnwan.io/traffic-profile
serviceRegistry:
  gcpServiceDirectory:
    pollInterval: 10
    region: us-west2
    projectID: my-project
    serviceAccountPath: ./credentials/serv-acc.json
```

**Important note**: with docker you will need to *mount* both the configuration
file **and** the service account: therefore `serviceAccountPath` needs to be
modified with the *mounted* path, not the one on your computer: take a look at
its value in the example `yaml` above.

Now, supposing that the configuration file is located at
`~/Desktop/options/conf.yaml`, run

```bash
docker run \
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
-v ~/Desktop/options/conf.yaml:/options/conf.yaml \
my-image --conf ./options/conf.yaml
```
