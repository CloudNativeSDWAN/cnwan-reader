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

## Example

This example follows the same requirements as the
[example](./usage#binary-examples) included for the binary version:

```bash
docker run \
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
my-image \
servicedirectory \
--project my-project \
--region us-west2 \
--metadata-key cnwan.io/traffic-profile \
--interval 10 \
--adaptor-api localhost/cnwan/events \
--service-account ./credentials/serv-acc.json
```
