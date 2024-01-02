// Copyright (c) 2024 Proton AG
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


#ifndef BRIDGE_GUI_APP_H
#define BRIDGE_GUI_APP_H


//**********************************************************************************************************************
/// \brief Bridge application class
//**********************************************************************************************************************
class BridgeApp : public QApplication {
Q_OBJECT
public: // member functions.
    BridgeApp(int &argc, char **argv);
    bool notify(QObject *object, QEvent *event) override;
    ///< Constructor.
    BridgeApp(BridgeApp const &) = delete; ///< Disabled copy-constructor.
    BridgeApp(BridgeApp &&) = delete; ///< Disabled assignment copy-constructor.
    ~BridgeApp() = default; ///< Destructor.
    BridgeApp &operator=(BridgeApp const &) = delete; ///< Disabled assignment operator.
    BridgeApp &operator=(BridgeApp &&) = delete; ///< Disabled move assignment operator.

signals:
    void fatalError(QString const &function, QString const &message); ///< Signal emitted when an fatal error occurs.
};


#endif //BRIDGE_GUI_APP_H
