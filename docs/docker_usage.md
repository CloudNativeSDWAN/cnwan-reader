# Docker usage

Before reading this, make sure you read [Usage](./usage.md) as it applies to this guide as well.

This guide describes usage that only applies to users that run the program inside a docker container.

You should run the program as

```bash
docker run <repository/image-name:tag-name>
```

Please read along to learn usage specific to Docker.

## Mount files

In order to use files from your local computer and make them available for use to the docker container, you must first *mount* those files.

This is done with the `-v` flag that must be entered **before** any other flag. With `-v` you first specify where the file is stored in your computer. Then, after a `:`, you specify where you wish to mount that file in the container, which is going to be the argument that other flags that accept a path, i.e. `--conf` will take.

For example, suppose that the path to the configuration file on your computer is `~/Desktop/cnwan/conf/cnwan-reader-conf.yaml` and that for simplicity you want to mount it in the container as `/conf/cnwan-reader.yaml`: then flag looks like this:

```bash
-v ~/Desktop/cnwan/conf/cnwan-reader-conf.yaml:/conf/cnwan-reader.yaml
```

Now, when you use a flag to load that file, you must enter the *mounted* path, that is the one you wrote after `:` -- remember that `-v` must be entered *before* all other flags:

```bash
docker run \
-v ~/Desktop/cnwan/conf/cnwan-reader-conf.yaml:/conf/cnwan-reader.yaml\
cnwan/cnwan-reader \
[command] \
--conf /conf/cnwan-reader.yaml
```

## Examples

### With Service Directory

This example follows the same requirements as this [example](./usage.md#with-service-directory) -- replace `cnwan/cnwan-reader` with your image in case you have built it yourself:

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

If you want to use a configuration file, you can create the same configuration file as in the [example](./usage.md#with-service-directory). For your convenience:

```yaml
adaptor: localhost:80/cnwan
metadataKeys:
  - cnwan.io/traffic-profile
serviceRegistry:
  gcpServiceDirectory:
    pollInterval: 10
    region: us-west2
    projectID: my-project
    serviceAccountPath: ./credentials/serv-acc.json
```

**Important note**: with docker you will need to *mount* both the configuration file **and** the service account: therefore `serviceAccountPath` needs to be modified with the *mounted* path, not the one on your computer: take a look at its value in the example `yaml` above.

Now, supposing that the configuration file is located at `~/Desktop/options/conf.yaml`, run

```bash
docker run \
-v ~/Desktop/cnwan-credentials/serv-acc.json:/credentials/serv-acc.json \
-v ~/Desktop/options/conf.yaml:/options/conf.yaml \
cnwan/cnwan-reader servicedirectory --conf ./options/conf.yaml
```

## With Cloud Map

This example follows the same requirements as this [example](./usage.md#with-cloud-map) -- replace `cnwan/cnwan-reader` with your image in case you have built it yourself:

```bash
docker run \
-v ~/Desktop/cnwan-credentials/aws-credentials:/credentials/aws-credentials \
cnwan/cnwan-reader \
poll cloudmap \
--region us-west-2 \
--metadata-keys cnwan.io/traffic-profile \
--adaptor-api localhost/cnwan/events \
--interval 10 \
--credentials-path /credentials/aws-credentials
```

The command above assumes you want to use the default *aws profile*: append the following flag to the command above in case you want to use another one:

```bash
-e AWS_PROFILE=your_profile
```

If you want to use a configuration file, you can create the same configuration file as in the [example](./usage.md#with-cloud-map). For your convenience:

```yaml
adaptor: localhost:80/cnwan
metadataKeys:
  - cnwan.io/traffic-profile
serviceRegistry:
  awsCloudMap:
    region: us-west-2
    pollInterval: 10
    credentialsPath: /credentials/aws-credentials
```

**Important note**: with docker you will need to *mount* both the configuration file **and** the credentials file: therefore `credentialsPath` needs to be modified with the *mounted* path, not the one on your computer: take a look at its value in the example `yaml` above.

Now, supposing that the configuration file is located at `~/Desktop/options/conf.yaml`, run

```bash
docker run \
-v ~/Desktop/cnwan-credentials/aws-credentials:/credentials/aws-credentials \
-v ~/Desktop/options/conf.yaml:/options/conf.yaml \
cnwan/cnwan-reader poll cloudmap --conf ./options/conf.yaml
```
