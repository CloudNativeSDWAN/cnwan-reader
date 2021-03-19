# Installation

The following sections detail some of the methods available to install and run the project. The easiest way to install the program is to use [go get](#go-get). Otherwise, if you want to build it yourself, i.e. in case you want to contribute or change its code, you can [build it yourself](#build-it-yourself).

If you prefer to use docker, you can follow [Docker Installation](./docker_installation.md).

## Option 1: Go Get

This is the easiest and fastest way to get and run the program and is recommended for users that just want to use the program without building or modifying it.

It requires Golang to be present on the machine: please follow [the official documentation](https://golang.org/doc/install) to install it in case you don't have it.

Execute

```bash
go get -u github.com/CloudNativeSDWAN/cnwan-reader
```

to download the project to your computer.

From now on, you can run the program by pointing its path, for example:

```bash
$HOME/go/bin/cnwan-reader COMMAND
```

Optionally, but very recommended, you can add it to your `$PATH`, so that you won't have to specify its full/relative path every time. Follow [Add to $Path](#optional:-add-to-path) for the details.

Follow [Usage](./usage.md) to learn how to use the program.

## Option 2: Build it yourself

As the previous section, this requires [Golang](https://golang.org/doc/install) to be installed on the machine in order to run the program and is most suitable for users that want to modify it or contribute to it.

### Clone the project

Run the commands below to clone the project and navigate to its root directory:

```bash
git clone github.com/CloudNativeSDWAN/cnwan-reader
cd cnwan-reader
```

### Optional: install Make

Now you need to build the program in order to use it.

Although you may use `go` commands to do so, we recommend using the included `Makefile` as this will automate a lot of commands. To use the `Makefile` you need to have `Make` installed, which comes already pre-installed if you are a *Unix/Linux/Mac* user. If you are a *Windows* user, you can download the binaries from [this page](http://gnuwin32.sourceforge.net/packages/make.htm).

### Build it

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

Optionally, but very recommended, you can add it to your `$PATH`, so that you won't have to specify its full/relative path every time. Follow [Add to $Path](#optional:-add-to-path) for the details.

Follow [Usage](./usage.md) to learn how to use the program.

## Optional but recommended: Add to Path

If you are a *Unix/Linux/Mac* user, you can execute:

```bash
PATH=$PATH:<path-to-the-binary-folder>
```

Replace `<path-to-the-binary-folder>` with the path of the directory where your executable file is. For example, if you installed the program through `go get` and supposing your golang folder is `$HOME/go/` (the default one usually) run:

```bash
PATH=$PATH:$HOME/go/bin
```

For *Windows* user and supposing your golang folder is in
`%USERPROFILE%\go`:

```powershell
set PATH=%PATH%;%USERPROFILE%\go\bin\
```

Or, still for *Windows*, you can follow [this guide](https://www.computerhope.com/issues/ch000549.htm) using your golang folder - usually `%USERPROFILE%\go\bin\` if you never changed it.

Now you can run the program as

```bash
cnwan-reader COMMAND
```

without having to mention its full/relative path every time.

Please note that this method will be valid only for your current shell session. To add it to your `PATH` permanently, we recommend you to look for the appropriate method for your operating system.
