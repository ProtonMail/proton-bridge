# bridge-gui

bridge-gui is the graphical user interface for Bridge. It's a C++ [Qt](https://www.qt.io) application  that 
communicates with the bridge executable via gRPC remote procedure call, on a local-only TLS-secured connection.

# Components

bridge-gui consists in 3 components:

- **bridge-gui**: The Bridge-GUI application itself. It's a QML Qt application that implements the bridge 
[gRPC](https://www.grpc.io) service to communicate and interact with the bridge application written in 
[Go](https://go.dev).
- **bridge-gui-tester**: A Qt widgets test application that offers a dummy gRPC server implementing the Bridge gRPC 
service. It can be used as a debugging and development tool, as it can simulate the server side (bridge) portion of 
the gRPC service.
- **bridgepp**: bridgepp (for bridge++) is a C++ static library that contains the code shared by bridge-gui and 
bridge-gui-tester.

# Bridge gRPC service

The bridge gRPC service that allows communications between bridge-gui, and the bridge Go application (or 
bridge-gui-tester) is defined in `../grpc/bridge.proto`.

# Supported platforms

bridge-gui runs on the same platforms as the bridge app:

- Linux x64.
- macOS x64 and Apple Silicon.
- Windows x64.

# Requirements:

- A C++ development toolchain ([GCC](https://gcc.gnu.org)/[Clang](https://clang.llvm.org)/
[MSVC](https://docs.microsoft.com/en-us/cpp/?view=msvc-170), [CMake](https://cmake.org), 
[Ninja](https://ninja-build.org)). An easy way to get all the needed tools is to use a modern IDE such 
as [CLion](https://www.jetbrains.com/clion/), [Qt Creator](https://www.qt.io/product/development-tools) or 
[Visual Studio Code](https://code.visualstudio.com) that will check and install or use their bundled versions
of the tools.
- [Qt 6](https://www.qt.io/download-open-source). Use the online installer to install the latest stable 
release of Qt for your platform and compiler.

# First build

bridge-gui uses [vcpkg](https://vcpkg.io/en/index.html), Microsoft multi-platform C/C++ dependency manager to get 
the source code and build gRPC and its dependencies (protobuf, zlib, ...). vcpkg is managed as a git submodule 
in `extern/vcpkg`,relative to the root of the bridge source tree. A pair of scripts is provided to perform 
the initialization of vcpkg, the retrieval and compilation of gRPC and a first build of the bridge-gui project
components:

- `build.sh`: a shell script to use on macOS and Linux.
- `build.ps1`: a [PowerShell](https://docs.microsoft.com/en-us/powershell/) script for Windows.

