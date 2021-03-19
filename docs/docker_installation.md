# Docker Installation

This guide is for user that prefer to run the program with docker.

You can either use the [official docker image](#use-the-official-image) or [build it yourself](#build-it-yoursel).

This requires you to have [Docker](https://www.docker.com/get-started) installed:

* *Unix/Linux* users with [Snap](https://snapcraft.io/docs/installing-snapd):

  ```bash
  sudo snap install docker
  ```

* *MacOs* users: [Docker Desktop for Mac](https://hub.docker.com/editions/community/docker-ce-desktop-mac/)
* *Windows* users: [Docker Desktop for Windows](https://hub.docker.com/editions/community/docker-ce-desktop-windows/)

## Option 1: Use the official image

Run

```bash
docker pull cnwan/cnwan-reader:v0.3.0
```

Now you can use the program as

```bash
docker run cnwan/cnwan-reader COMMAND
```

Follow the [Docker Usage](./docker_usage.md) section to learn how to use it.

## Option 2: Build it yourself

This method is most suitable for users that want to modify it or contribute to it.

### Clone the project

Run the commands below to clone the project and navigate to its root directory:

```bash
git clone github.com/CloudNativeSDWAN/cnwan-reader
cd cnwan-reader
```

### Optional: install Make

Now you need to build the program in order to use it.

Although you may use `docker` commands to do so, we recommend using the included `Makefile` as this will automate a lot of commands. To use the `Makefile` you need to have `Make` installed, which comes already pre-installed if you are a *Unix/Linux/Mac* user. If you are a *Windows* user, you can download the binaries from [this page](http://gnuwin32.sourceforge.net/packages/make.htm).

### Build it

Execute:

```bash
make docker-build IMG=<repository/image-name:tag-name>
```

To avoid specifying the `IMG` parameter every time, you can modify the top of the `Makefile` to look like this:

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
docker run <repository/image-name:tag-name> COMMAND
```

Follow the [Docker Usage](./docker_usage.md) section to learn how to use it.

## Optional: Rename it

If you want to avoid specifying its full repository name, you can rename it.

Execute:

```bash
# If you are using the official docker image
docker rename cnwan/cnwan-reader cnwan-reader

# If you have built it yourself
docker rename <repository/image-name:tag-name> cnwan-reader
```

You can now run the program as just

```bash
docker run cnwan-reader COMMAND
```

Note that this is only for running it locally, as it won't be a valid repository name.

## Push it

If you also wish to push the container to a container registry, make sure you are correctly logged in to it. Most of the times, [this guide](https://docs.docker.com/engine/reference/commandline/login/) should do it, but we encourage you to read your container registry's official documentation to learn how to do that.

Your image name should respect the container registry format: i.e. if you are using [DockerHub](https://hub.docker.com/) the name of your image should be something like `your-username/image-name:tag-name`. For other registries the full repository URL should be included, i.e. `registry.com/your-username/image-name:tag-name`.

Finally, to push it to a container registry, and supposing you have modified the `Makefile` as described in the previous sections:

```bash
make docker-push
```
