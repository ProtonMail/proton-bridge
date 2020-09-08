# ProtonMail Import-Export Qt interface
Import-Export uses [Qt](https://www.qt.io) framework for creating appealing graphical
user interface. Package [therecipe/qt](https://github.com/therecipe/qt) is used
to implement Qt into [Go](https://www.goglang.com).


# For developers
The GUI is designed inside QML files. Communication with backend is done via
[frontend.go](./frontend.go). The API documentation is done via `go-doc`.

## Setup
* if you don't have the system wide `go-1.8.1` download, install localy (e.g.
  `~/build/go-1.8.1`) and setup:

        export GOROOT=~/build/go-1.8.1/go
        export PATH=$GOROOT/bin:$PATH

* go to your working directory and export `$GOPATH`

        export GOPATH=`Pwd`
        mkdir -p $GOPATH/bin
        export PATH=$PATH:$GOPATH/bin


* if you dont have system wide `Qt-5.8.0`
  [download](https://download.qt.io/official_releases/qt/5.8/5.8.0/qt-opensource-linux-x64-5.8.0.run),
  install locally (e.g. `~/build/qt/qt-5.8.0`) and setup:

        export QT_DIR=~/build/qt/qt-5.8.0
        export PATH=$QT_DIR/5.8/gcc_64/bin:$PATH

* `Go-Qt` setup (installation is system dependent see
  [therecipe/qt/README](https://github.com/therecipe/qt/blob/master/README.md)
  for details)

        go get -u -v github.com/therecipe/qt/cmd/...
        $GOPATH/bin/qtsetup

## Compile
* it is necessary to compile the Qt-C++ with go for resources and meta-objects

        make -f Makefile.local

* FIXME the rcc file is implicitly generated with `package main`. This needs to
  be changed to `package qtie` manually
* check that user interface is working

        make -f Makefile.local test

## Test

        make -f Makefile.local qmlpreview

## Deploy
* before compilation of Import-Export it is necessary to run compilation of Qt-C++ part (done in makefile)
