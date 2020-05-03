# Building ProtonMail Bridge app

## Prerequisites
* Go 1.13
* Bash with basic build utils: make, gcc, sed, find, grep, ...
* For Windows it is recommended to use MinGW 64bit shell from [MSYS2](https://www.msys2.org/)
* GCC (linux, windows) or Xcode (macOS)
* libsecret (Linux)
* Windres (windows)

To enable the sending of crash reports using Sentry please set the
`main.DSNSentry` value with the client key of your sentry project before build.
Otherwise, the sending of crash reports will be disabled.

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
    * for `linux`, the binary will have the name of the project directory (e.g `bridge`)
    * for `windows`, the binary will have the file extension `.exe` (e.g `bridge.exe`)
    * for `darwin`, the application will be created with name of the project directory (e.g `bridge.app`)

## Useful tests, lints and checks
In order to be able to run following commands please install the development dependencies: 
`make install-dev-dependencies`

* `make test` will run all unit tests
* `make lint` will lint the whole project
* `make -C ./tests test` will run the integration tests
* `make run` will build Bridge without a GUI and start it in CLI mode
