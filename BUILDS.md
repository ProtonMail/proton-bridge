# Building Proton Mail Bridge

## Prerequisites
* 64-bit OS:
    - the go-rfc5322 module cannot currently be compiled for 32-bit OSes
* Go 1.21.9
* Bash with basic build utils: make, gcc, sed, find, grep, ...
  - For Windows, it is recommended to use MinGW 64bit shell from [MSYS2](https://www.msys2.org/)
* GCC (Linux), msvc (Windows) or Xcode (macOS)
* Windres (Windows)
* libglvnd and libsecret development files (Linux)
* pkg-config (Linux)
* cmake, ninja-build and Qt 6.4.3 are required to build the graphical user interface. On Linux, 
the Mesa OpenGL development files are also needed.

To enable the sending of crash reports using Sentry please set the
`DSN_SENTRY` environment variable with the client key of your sentry project before build.
Otherwise, the sending of crash reports will be disabled.

## Build
In order to build Bridge app with Qt interface we are using
[Qt 6.4.3](https://doc.qt.io/qt-6/gettingstarted.html).

Please note that qmake path must be in your `PATH` to ensure Qt to be found.
Also, before you start build **on Windows**, please unset the `MSYSTEM` variable

```bash
export MSYSTEM=
```

### Build Bridge
* in project root run

```bash
make build
```

* The result will be stored in `./cmd/Destop-Bridge/deploy/${GOOS}/`
    * for `linux`, the binary will have the name of the project directory (e.g `proton-bridge`)
    * for `windows`, the binary will have the file extension `.exe` (e.g `proton-bridge.exe`)
    * for `darwin`, the application will be created with name of the project directory (e.g `proton-bridge.app`)

#### Build Bridge without GUI
* If you need to build bridge without Qt dependencies, you can do so by running

```bash
make build-nogui
```

* To launch Bridge without GUI, you can invoke the `bridge` executable with one the following command-line switches:
  * `--noninteractive` or `-n` to start Bridge without any interface (i.e., there is no way to add or remove client, get bridge password, etc.)
  * `--cli` or `-c` to start Bridge with an interactive terminal interface.
* NOTE: You still need to set up a supported keychain on your system.

## Launchers
Launchers are only included in official distributions and provide the public
key used to verify signed app binaries, allowing the automatic update feature.
See README for more information.

## Tags
Note that repository contains both Bridge and Import-Export apps and they are
not released together. Therefore, each app has own tag prefix. Bridge tags
starts with `br-` and Import-Export tags starts with `ie-`. Both tags continue
with semantic versioning `MAJOR.MINOR.PATCH`. An example of full tag is
`br-1.4.4` or `ie-1.1.2` (current versions in October 2020).

## Useful tests, lints and checks
In order to be able to run following commands please install the development dependencies: 
`make install-dev-dependencies`

* `make test` will run all unit tests
* `make lint` will lint the whole project
* `make -C ./test test` will run the integration tests
* `make run` will build Bridge without a GUI and start it in CLI mode
