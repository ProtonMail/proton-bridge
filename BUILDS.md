# Building ProtonMail Bridge app

## Prerequisites
* Go 1.13
* Bash with basic build utils: make, gcc, sed, find, grep, …
* For Windows it is recommended to use MinGW 64bit shell from [MSYS2](https://www.msys2.org/)
* GCC (linux, windows) or Xcode (macOS)
* Windres (windows)

To enable the sending of crash reports using Sentry please set the
`main.DSNSentry` value with client key of your sentry project before build.
Otherwise sending of crash reports will be disabled.

## Build
* for Windows please unset the `MSYSTEM` variable

```bash
export MSYSTEM=
```

* in project root run

```bash
make build
```

* The result will be stored in `./cmd/Destop-Bridge/deploy/${GOOS}/`
    * for `linux` binary will the name of project directory e.g `bridge`
    * for `windows` the binary has extension `.exe` e.g `bridge.exe`
    * for `darwin` the application will be created with name of project directory e.g `bridge.app`

## Usefull tests, lints and checks
In order to be able to run following commands please install development dependencies:  `make install-dev-dependencies`

* `make test` will run unit test for whole project
* `make lint` will run liter for whole project
* `make -C ./tests test` will run integration tests for Bridge application
* `make run` will compile without GUI and start Bridge application in CLI mode
