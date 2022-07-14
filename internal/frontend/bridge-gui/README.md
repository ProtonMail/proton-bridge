## Prerequisite

### Linux (debian and derivates)
```` bash
sudo apt install build-essential
sudo apt install tar curl zip unzip
sudo apt install linux-headers-$(uname -r)
sudo apt install mesa-common-dev libglu1-mesa-dev
````

### macOS & WIndows

Coming soon...


###  Define QT6DIR

``` bash
export QT6DIR=/opt/Qt/6.3.1/gcc_64
```

### install vcpkg and define VCPKG_ROOT

``` bash
git clone https://github.com/Microsoft/vcpkg.git
./vcpkg/bootstrap-vcpkg.sh 
export VCPKG_ROOT=$PWD/vcpkg
```

## install grpc & protobuf

``` bash
./vcpkg install grpc
```

## Building

A simple script is provided that run the appropriate CMake command.

``` bash
./build.sh
``` 

## Running

Simply run from the cmake build folder (`cmake-build-debug` by default) 
``` bash
./bridge-gui
```

`bridge-gui` will launch the `bridge` executable that it will try to locate in

- The working directory.
- The application directory.
- `cmd/Desktop-Bridge/`, `../cmd/Desktop-Bridge/`,  `../../cmd/Desktop-Bridge` 
(up to five parent folders above the current folder are inspected).

you can specify the location of the bridge executable using the `-b` or 
`--bridge-exe-path` command-line parameter:

``` bash
./bridge-gui -b "~/bin/bridge"
```

you can also ask bridge-gui to connect to an already running instance of `bridge`
using the `-a` or `--attach` command line parameter.

``` bash
./bridge-gui -a
```
