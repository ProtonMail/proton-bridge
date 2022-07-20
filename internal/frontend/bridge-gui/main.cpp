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
#include "BridgeMonitor.h"


//****************************************************************************************************************************************************
/// // initialize the Qt application.
//****************************************************************************************************************************************************
void initQtApplication()
{
    QString const qsgInfo = QProcessEnvironment::systemEnvironment().value("QSG_INFO");
    if ((!qsgInfo.isEmpty()) && (qsgInfo != "0"))
        QLoggingCategory::setFilterRules("qt.scenegraph.general=true");

    /// \todo GODT-1670 Get version from go backend.
    QGuiApplication::setApplicationName("Proton Mail Bridge");
    QGuiApplication::setApplicationVersion("3.0");
    QGuiApplication::setOrganizationName("Proton AG");
    QGuiApplication::setOrganizationDomain("proton.ch");
    QGuiApplication::setQuitOnLastWindowClosed(false);
 }


//****************************************************************************************************************************************************
/// \return A reference to the log.
//****************************************************************************************************************************************************
Log &initLog()
{
    Log &log = app().log();
    log.setEchoInConsole(true);
    log.setLevel(Log::Level::Debug);
    return log;
}


//****************************************************************************************************************************************************
/// \param[in] engine The QML component.
//****************************************************************************************************************************************************
QQmlComponent *createRootQmlComponent(QQmlApplicationEngine &engine)
{
    QString const qrcQmlDir = "qrc:/qml";

    qmlRegisterSingletonInstance("Proton", 1, 0, "Backend", &app().backend());
    qmlRegisterType<UserList>("Proton", 1, 0, "UserList");
    qmlRegisterType<User>("Proton", 1, 0, "User");

    auto rootComponent = new QQmlComponent(&engine, &engine);

    engine.addImportPath(qrcQmlDir);
    engine.addPluginPath(qrcQmlDir);
    QQuickStyle::setStyle("Proton");

    rootComponent->loadUrl(QUrl(qrcQmlDir + "/Bridge.qml"));
    if (rootComponent->status() != QQmlComponent::Status::Ready)
    {
        app().log().error(rootComponent->errorString());
        throw Exception("Could not load QML component");
    }
    return rootComponent;
}


//****************************************************************************************************************************************************
/// \param[in] exePath The path of the Bridge executable. If empty, the function will try to locate the bridge application.
//****************************************************************************************************************************************************
void launchBridge(QString const &exePath)
{
    UPOverseer& overseer = app().bridgeOverseer();
    overseer.reset();

    QString bridgeExePath = exePath;
    if (exePath.isEmpty())
        bridgeExePath = BridgeMonitor::locateBridgeExe();

    if (bridgeExePath.isEmpty())
        throw Exception("Could not locate the bridge executable path");
    else
        app().log().debug(QString("Bridge executable path: %1").arg(QDir::toNativeSeparators(bridgeExePath)));

    overseer = std::make_unique<Overseer>(new BridgeMonitor(bridgeExePath, nullptr), nullptr);
    overseer->startWorker(true);
}


//****************************************************************************************************************************************************
/// \param[in] argc The number of command-line arguments.
/// \param[in] argv The list of command line arguments.
//****************************************************************************************************************************************************
void parseArguments(int argc, char **argv, bool &outAttach, QString &outExePath)
{
    // for unknown reasons, on Windows QCoreApplication::arguments() frequently returns an empty list, which is incorrect, so we rebuild the argument
    // list from the original argc and argv values.
    QStringList args;
    for (int i = 0; i < argc; i++)
        args.append(QString::fromLocal8Bit(argv[i]));

    // We do not want to 'advertise' the following switches, so we do not offer a '-h/--help' option.
    // we have not yet connected to Bridge, we do not know the application version number, so we do not offer a -v/--version switch.
    QCommandLineParser parser;
    parser.setApplicationDescription("Proton Mail Bridge");
    QCommandLineOption attachOption(QStringList() << "attach" << "a", "attach to an existing bridge process");
    parser.addOption(attachOption);
    QCommandLineOption exePathOption(QStringList() << "bridge-exe-path" << "b", "bridge executable path", "path", QString());
    parser.addOption(exePathOption);

    parser.process(args);
    outAttach = parser.isSet(attachOption);
    outExePath = parser.value(exePathOption);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void closeBridgeApp()
{
    UPOverseer& overseer = app().bridgeOverseer();
    if (!overseer) // The app was ran in 'attach' mode and attached to an existing instance of Bridge. No need to close.
        return;

    app().grpc().quit(); // this will cause the grpc service and the bridge app to close.
    while (!overseer->isFinished())
    {
        QThread::msleep(20);
    }
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
        if (QSysInfo::productType() != "windows")
            QCoreApplication::setAttribute(Qt::AA_UseSoftwareOpenGL);

        QGuiApplication guiApp(argc, argv);
        initQtApplication();

        bool attach = false;
        QString exePath;
        parseArguments(argc, argv, attach, exePath);

        Log &log = initLog();

        if (!attach)
            launchBridge(exePath);

        app().backend().init();

        QQmlApplicationEngine engine;
        std::unique_ptr<QQmlComponent> rootComponent(createRootQmlComponent(engine));
        std::unique_ptr<QObject>rootObject(rootComponent->create(engine.rootContext()));
        if (!rootObject)
            throw Exception("Could not create root object.");

        BridgeMonitor *bridgeMonitor = app().bridgeMonitor();
        bool bridgeExited = false;
        QMetaObject::Connection connection;
        if (bridgeMonitor)
            connection = QObject::connect(bridgeMonitor, &BridgeMonitor::processExited, [&](int returnCode) {
                // GODT-1671 We need to find a 'safe' way to check if brige crashed and restart instead of just quitting. Is returnCode enough?
                bridgeExited = true;// clazy:exclude=lambda-in-connect
                qGuiApp->exit(returnCode);
            });

        int const result = QGuiApplication::exec();

        QObject::disconnect(connection);
        app().grpc().stopEventStream();

        // We manually delete the QML components to avoid warnings error due to order of deletion of C++ / JS objects and singletons.
        rootObject.reset();
        rootComponent.reset();

        if (!bridgeExited)
            closeBridgeApp();

        return result;
    }
    catch (Exception const &e)
    {
        app().log().error(e.qwhat());
        return EXIT_FAILURE;
    }
}


