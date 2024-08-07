// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.
#ifndef BRIDGE_GUI_CLIPBOARDPROXY_H
#define BRIDGE_GUI_CLIPBOARDPROXY_H

// The following class declarations were taken and adapted from:
// https://stackoverflow.com/questions/40092352/passing-qclipboard-to-qml
// Author: krzaq


class ClipboardProxy : public QObject
{
Q_OBJECT
    Q_PROPERTY(QString text READ text NOTIFY textChanged)
public:
    explicit ClipboardProxy(QClipboard*);

    QString text() const;

signals:
    void textChanged();

private:
    QClipboard* clipboard;
};


#endif //BRIDGE_GUI_CLIPBOARDPROXY_H
