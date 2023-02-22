// Copyright (c) 2023 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.


#include "BridgeLib.h"
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/BridgeUtils.h>


using namespace bridgepp;

#ifdef __cplusplus
extern "C" {
#endif


typedef char *(*FuncReturningCString)();


#ifdef __cplusplus
}
#endif


namespace {


FuncReturningCString goosFunc = nullptr; ///< A pointer to the dynamically loaded GoOS function.
FuncReturningCString userCacheDirFunc = nullptr; ///< A pointer to the dynamically loaded UserCache function.
FuncReturningCString userConfigDirFunc = nullptr; ///< A pointer to the dynamically loaded UserConfig function.
FuncReturningCString userDataDirFunc = nullptr; ///< A pointer to the dynamically loaded UserData function.
void (*deleteCStringFunc)(char *) = nullptr; ///< A pointer to the deleteCString function.


#if defined(Q_OS_WINDOWS)
#include <windows.h>
typedef HINSTANCE LibHandle;
#else
typedef void *LibHandle;
#endif


LibHandle loadDynamicLibrary(QString const &path); ///< Load a dynamic library.
void *getFuncPointer(LibHandle lib, QString const &funcName); ///< Retrieve a function pointer from a dynamic library.


//****************************************************************************************************************************************************
/// \return The path to the bridgelib library file.
//****************************************************************************************************************************************************
QString bridgelibPath() {
    QString const path = QDir(QCoreApplication::applicationDirPath()).absoluteFilePath("bridgelib.");
    switch (os()) {
    case OS::Windows:
        return path + "dll";
    case OS::MacOS:
        return path + "dylib";
    case OS::Linux:
    default:
        return path + "so";
    }
}


#if defined(Q_OS_WINDOWS)


//****************************************************************************************************************************************************
/// \param[in] path The path of the library file.
/// \return A pointer to the library object
//****************************************************************************************************************************************************
LibHandle loadDynamicLibrary(QString const &path) {
    if (!QFileInfo::exists(path)) {
        throw Exception(QString("The dynamic library file bridgelib.dylib could not be found at '%1'.").arg(path));
    }

    LibHandle handle = LoadLibrary(reinterpret_cast<LPCWSTR>(path.toStdWString().c_str()));
    if (!handle) {
        throw Exception(QString("The bridgelib dynamic library file '%1' could not be opened.").arg(path));
    }

    return handle;
}


//****************************************************************************************************************************************************
/// \param[in] lib A handle to the library
/// \param[in] funcName The name of the function.
/// \return A pointer to the function
//****************************************************************************************************************************************************
void *getFuncPointer(LibHandle lib, QString const &funcName) {
    void *pointer = reinterpret_cast<void*>(GetProcAddress(lib, funcName.toLocal8Bit()));
    if (!pointer)
        throw Exception(QString("Could not locate function %1 in bridgelib dynamic library").arg(funcName));

    return pointer;
}

#else


#include <dlfcn.h>


//****************************************************************************************************************************************************
/// \param[in] path The path of the library file.
/// \return A pointer to the library object
//****************************************************************************************************************************************************
void *loadDynamicLibrary(QString const &path) {
    if (!QFileInfo::exists(path)) {
        throw Exception(QString("The dynamic library file bridgelib.dylib could not be found at '%1'.").arg(path));
    }

    void *lib = dlopen(path.toLocal8Bit().data(), RTLD_LAZY);
    if (!lib) {
        throw Exception(QString("The bridgelib dynamic library file '%1' could not be opened.").arg(path));
    }

    return lib;
}


//****************************************************************************************************************************************************
/// \param[in] lib A handle to the library
/// \param[in] funcName The name of the function.
/// \return A pointer to the function
//****************************************************************************************************************************************************
void *getFuncPointer(LibHandle lib, QString const &funcName) {
    void *pointer = dlsym(lib, funcName.toLocal8Bit());
    if (!pointer) {
        throw Exception(QString("Could not locate function %1 in bridgelib dynamic library").arg(funcName));
    }

    return pointer;
}


#endif // defined(Q_OS_WINDOWS)


}


namespace bridgelib {


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void loadLibrary() {
    try {
        LibHandle lib = loadDynamicLibrary(bridgelibPath());
        goosFunc = reinterpret_cast<FuncReturningCString>(getFuncPointer(lib, "GoOS"));
        userCacheDirFunc = reinterpret_cast<FuncReturningCString>(getFuncPointer(lib, "UserCacheDir"));
        userConfigDirFunc = reinterpret_cast<FuncReturningCString>(getFuncPointer(lib, "UserConfigDir"));
        userDataDirFunc = reinterpret_cast<FuncReturningCString>(getFuncPointer(lib, "UserDataDir"));
        deleteCStringFunc = reinterpret_cast<void (*)(char*)>(getFuncPointer(lib, "DeleteCString"));

    } catch (Exception const &e) {
        throw Exception("Error loading the bridgelib dynamic library file.", e.qwhat());
    }
}


//****************************************************************************************************************************************************
/// \brief Converts a C-style string returned by a go library function to a QString, and release the memory allocated for the C-style string.
/// \param[in] cString The C-style string, in UTF-8 format.
/// \return A QString.
//****************************************************************************************************************************************************
QString goToQString(char *const cString) {
    if (!cString) {
        return QString();
    }
    QString const result = QString::fromUtf8(cString);
    deleteCStringFunc(cString);

    return result;
}


//****************************************************************************************************************************************************
/// \return The value of the Go runtime.GOOS constant.
//****************************************************************************************************************************************************
QString goos() {
    return goToQString(goosFunc());
}


//****************************************************************************************************************************************************
/// \return The path to the user cache folder.
//****************************************************************************************************************************************************
QString userCacheDir() {
    return goToQString(userCacheDirFunc());
}


//****************************************************************************************************************************************************
/// \return The path to the user cache folder.
//****************************************************************************************************************************************************
QString userConfigDir() {
    return goToQString(userConfigDirFunc());
}


//****************************************************************************************************************************************************
/// \return The path to the user data folder.
//****************************************************************************************************************************************************
QString userDataDir() {
    return goToQString(userDataDirFunc());
}


}