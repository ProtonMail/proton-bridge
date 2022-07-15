## Prerequisite

```` bash
sudo apt install build-essential
sudo apt install tar curl zip unzip
sudo apt install linux-headers-$(uname -r)
sudo apt install mesa-common-dev libglu1-mesa-dev
````

## Define Qt5DIR

```` bash
export QT5DIR=/opt/Qt/5.13.0/gcc_64
````

## install vcpkg and define VCPKG_ROOT

```` bash
git clone https://github.com/Microsoft/vcpkg.git
./vcpkg/bootstrap-vcpkg.sh 
export VCPKG_ROOT=$PWD/vcpkg
````

## install grpc & protobuf

```` bash
./vcpkg install grpc
````
