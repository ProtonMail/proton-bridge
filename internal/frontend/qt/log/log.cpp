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


// +build build_qt

#include "log.h"
#include "_cgo_export.h"

#include <QObject>
#include <QByteArray>
#include <QString>
#include <QVector>
#include <QtGlobal>

void messageHandler(QtMsgType type, const QMessageLogContext &context, const QString &msg)
{
    Q_UNUSED( type )
        Q_UNUSED( context )
        QByteArray localMsg = msg.toUtf8().prepend("WHITESPACE");
    logMsgPacked(
            const_cast<char*>( (localMsg.constData()) +10 ),
            localMsg.size()-10
            );
}

void InstallMessageHandler() {
    qInstallMessageHandler(messageHandler);
}

