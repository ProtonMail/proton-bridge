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


#include "MainWindow.h"
#include "AppController.h"
#include "GRPCServerWorker.h"
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/Worker/Overseer.h>


#ifndef BRIDGE_APP_VERSION
#error "BRIDGE_APP_VERSION is not defined"
#endif


namespace {


QString const applicationName = "Proton Mail Bridge GUI Tester"; ///< The name of the application.


}


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] argc The number of command-line arguments.
/// \param[in] argv The list of command-line arguments.
/// \return The exit code for the application.
//****************************************************************************************************************************************************
int main(int argc, char **argv) {

    try {
        QApplication a(argc, argv);
        QApplication::setApplicationName(applicationName);
        QApplication::setOrganizationName("Proton AG");
        QApplication::setOrganizationDomain("proton.ch");
        QApplication::setQuitOnLastWindowClosed(true);

        Log &log = app().log();
        log.setEchoInConsole(true);
        log.setLevel(Log::Level::Debug);
        log.info(QString("%1 started.").arg(applicationName));

        MainWindow window(nullptr);
        app().setMainWindow(&window);
        window.setWindowTitle(QApplication::applicationName());
        window.show();

        auto *serverWorker = new GRPCServerWorker(nullptr);
        QObject::connect(serverWorker, &Worker::started, []() { app().log().info("Server worker started."); });
        QObject::connect(serverWorker, &Worker::finished, []() { app().log().info("Server worker finished."); });
        QObject::connect(serverWorker, &Worker::error, [&](QString const &message) {
            app().log().error(message);
            qApp->exit(EXIT_FAILURE);
        });
        UPOverseer overseer = std::make_unique<Overseer>(serverWorker, nullptr);
        overseer->startWorker(true);

        qint32 const exitCode = QApplication::exec();

        serverWorker->stop();
        if (!overseer->wait(5000)) {
            log.warn("gRPC server took too long to finish.");
        }

        app().log().info(QString("%1 exiting with code %2.").arg(applicationName).arg(exitCode));
        return exitCode;
    }
    catch (Exception const &e) {
        QString message = e.qwhat();
        if (!e.details().isEmpty())
            message += "\n\nDetails:\n" + e.details();
        QTextStream(stderr) << QString("A fatal error occurred: %1\n").arg(message);
        return EXIT_FAILURE;
    }
}


