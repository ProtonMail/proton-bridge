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


#include "Pch.h"
#include "Exception.h"
#include "QMLBackend.h"
#include "Log.h"
#include "EventStreamWorker.h"


//****************************************************************************************************************************************************
/// // initialize the Qt application.
//****************************************************************************************************************************************************
std::shared_ptr<QGuiApplication> initQtApplication(int argc, char *argv[])
{
    // Note the two following attributes must be set before instantiating the QCoreApplication/QGuiApplication class.
    QCoreApplication::setAttribute(Qt::AA_EnableHighDpiScaling, false);
    if (QSysInfo::productType() != "windows")
        QCoreApplication::setAttribute(Qt::AA_UseSoftwareOpenGL);

    QString const qsgInfo = QProcessEnvironment::systemEnvironment().value("QSG_INFO");
    if ((!qsgInfo.isEmpty()) && (qsgInfo != "0"))
        QLoggingCategory::setFilterRules("qt.scenegraph.general=true");

    auto app = std::make_shared<QGuiApplication>(argc, argv);

    /// \todo GODT-1670 Get version from go backend.
    QGuiApplication::setApplicationName("Proton Mail Bridge");
    QGuiApplication::setApplicationVersion("3.0");
    QGuiApplication::setOrganizationName("Proton AG");
    QGuiApplication::setOrganizationDomain("proton.ch");
    QGuiApplication::setQuitOnLastWindowClosed(false);

    return app;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void initLog()
{
    Log &log = app().log();
    log.setEchoInConsole(true);
    log.setLevel(Log::Level::Debug);
}


//****************************************************************************************************************************************************
/// \param[in] engine The QML engine.
//****************************************************************************************************************************************************
QQmlComponent *createRootQmlComponent(QQmlApplicationEngine &engine)
{
    QString const qrcQmlDir = "qrc:/qml";

    qmlRegisterType<QMLBackend>("CppBackend", 1, 0, "QMLBackend");
    qmlRegisterType<UserList>("CppBackend", 1, 0, "UserList");
    qmlRegisterType<User>("CppBackend", 1, 0, "User");

    auto rootComponent = new QQmlComponent(&engine, &engine);

    engine.addImportPath(qrcQmlDir);
    engine.addPluginPath(qrcQmlDir);

    QQuickStyle::addStylePath(qrcQmlDir);
    QQuickStyle::setStyle("Proton");

    rootComponent->loadUrl(QUrl(qrcQmlDir + "/Bridge.qml"));
    if (rootComponent->status() != QQmlComponent::Status::Ready)
        throw Exception("Could not load QML component");

    return rootComponent;
}


//****************************************************************************************************************************************************
/// \param[in] argc The number of command-line arguments.
/// \param[in] argv The list of command-line arguments.
/// \return The exit code for the application.
//****************************************************************************************************************************************************
int main(int argc, char *argv[])
{
    try
    {
        std::shared_ptr<QGuiApplication> guiApp = initQtApplication(argc, argv);
        initLog();

        /// \todo GODT-1667 Locate & Launch go backend (and wait for it).

        app().backend().init();

        QQmlApplicationEngine engine;
        QQmlComponent *rootComponent = createRootQmlComponent(engine);

        QObject *rootObject = rootComponent->beginCreate(engine.rootContext());
        if (!rootObject)
            throw Exception("Could not create root object.");
        rootObject->setProperty("backend", QVariant::fromValue(&app().backend()));
        rootComponent->completeCreate();

        int result = QGuiApplication::exec();

        app().log().info(QString("Exiting app with return code %1").arg(result));

        app().grpc().stopEventStream();
        app().backend().clearUserList();

        /// \todo GODT-1667 shutdown go backend.

        return result;
    }
    catch (Exception const &e)
    {
        app().log().error(e.qwhat());
        return EXIT_FAILURE;
    }
}


