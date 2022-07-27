// Copyright (c) 2022 Proton AG
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


#ifndef BRIDGE_GUI_PCH_H
#define BRIDGE_GUI_PCH_H


#include <QtCore>
#include <QtQuick>
#include <QtQml>
#include <QtQuickControls2>
#include <AppController.h>


#if defined(Q_OS_WIN32) && defined(ERROR)
// The folks at Microsoft have decided that it was OK to `#define ERROR 0` in wingdi.h. It is not OK, because
// any occurrence of ERROR, even scoped, will be substituted. For instance Log::Level::ERROR (case imposed by gRPC).
#undef ERROR
#endif


#endif // BRIDGE_GUI_PCH_H
