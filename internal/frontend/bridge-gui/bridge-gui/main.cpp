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


#include "QMLBackend.h"
#include "BridgeMonitor.h"
#include "Version.h"
#include <bridgepp/Log/Log.h>
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/Exception/Exception.h>


using namespace bridgepp;


namespace
{
    QString const launcherFlag = "--launcher"; ///< launcher flag parameter used for bridge.
    QString const bridgeLock = "bridge-gui.lock"; ///< file name used for the lock file.
}


//****************************************************************************************************************************************************
/// // initialize the Qt application.
//****************************************************************************************************************************************************
void initQtApplication()
{
    QString const qsgInfo = QProcessEnvironment::systemEnvironment().value("QSG_INFO");
    if ((!qsgInfo.isEmpty()) && (qsgInfo != "0"))
        QLoggingCategory::setFilterRules("qt.scenegraph.general=true");

    QGuiApplication::setApplicationName("Proton Mail Bridge");
    QGuiApplication::setApplicationVersion(PROJECT_VER);
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
    log.registerAsQtMessageHandler();
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
    qmlRegisterType<bridgepp::User>("Proton", 1, 0, "User");

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
/// \param[in] lock The lock file to be checked.
/// \return True if the lock can be taken, false otherwize. 
//****************************************************************************************************************************************************
bool checkSingleInstance(QLockFile &lock)
{
    lock.setStaleLockTime(0);
    if (!lock.tryLock())
    {
        qint64 pid;
        QString hostname, appName, details;
        if (lock.getLockInfo(&pid, &hostname, &appName))
            details = QString("(PID : %1 - Host : %2 - App : %3)").arg(pid).arg(hostname, appName);

        app().log().error(QString("Instance already exists %1 %2").arg(lock.fileName(), details));
        return false;
    }
    else
    {
        app().log().info(QString("lock file created %1").arg(lock.fileName()));
    }
    return true;
}


//****************************************************************************************************************************************************
/// \param [in]  argc number of arguments passed to the application.
/// \param [in]  argv list of arguments passed to the application.
/// \param [out] args list of arguments passed to the application as a QStringList.
/// \param [out] launcher launcher used in argument, forced to self application if not specify.
/// \param[out] outAttach The value for the 'attach' command-line parameter.
//****************************************************************************************************************************************************
void parseArguments(int argc, char *argv[], QStringList& args, QString& launcher, bool &outAttach) {
    bool flagFound = false;
    launcher = QString::fromLocal8Bit(argv[0]);
    // for unknown reasons, on Windows QCoreApplication::arguments() frequently returns an empty list, which is incorrect, so we rebuild the argument
    // list from the original argc and argv values.
   for (int i = 1; i < argc; i++) {
        QString const &arg = QString::fromLocal8Bit(argv[i]);
        // we can't use QCommandLineParser here since it will fail on unknown options.
        // Arguments may contain some bridge flags.
        if (arg == launcherFlag)
        {
            args.append(arg);
            launcher = QString::fromLocal8Bit(argv[++i]);
            args.append(launcher);
            flagFound = true;
        }
#ifdef QT_DEBUG
        else if (arg == "--attach" || arg == "-a")
        {
            // we don't keep the attach mode within the args since we don't need it for Bridge.
            outAttach = true;
        }
#endif
        else
        {
            args.append(arg);
        }
    }
    if (!flagFound)
    {
        // add bridge-gui as launcher
        args.append(launcherFlag);
        args.append(launcher);
    }
}


//****************************************************************************************************************************************************
/// \param [in] args list of arguments to pass to bridge.
//****************************************************************************************************************************************************
void launchBridge(QStringList const &args)
{
    UPOverseer& overseer = app().bridgeOverseer();
    overseer.reset();

    const QString bridgeExePath = BridgeMonitor::locateBridgeExe();

    if (bridgeExePath.isEmpty())
        throw Exception("Could not locate the bridge executable path");
    else
        app().log().debug(QString("Bridge executable path: %1").arg(QDir::toNativeSeparators(bridgeExePath)));

    overseer = std::make_unique<Overseer>(new BridgeMonitor(bridgeExePath, args, nullptr), nullptr);
    overseer->startWorker(true);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void closeBridgeApp()
{
    UPOverseer& overseer = app().bridgeOverseer();
    if (!overseer) // The app was run in 'attach' mode and attached to an existing instance of Bridge. No need to close.
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

        Log &log = initLog();

        QLockFile lock(bridgepp::userCacheDir() + "/" + bridgeLock);
        if (!checkSingleInstance(lock))
            return EXIT_FAILURE;

        QStringList args;
        QString launcher;
        bool attach = false;
        parseArguments(argc, argv, args, launcher, attach);

        if (!attach)
            launchBridge(args);

        app().backend().init();

        QQmlApplicationEngine engine;
        std::unique_ptr<QQmlComponent> rootComponent(createRootQmlComponent(engine));
        std::unique_ptr<QObject>rootObject(rootComponent->create(engine.rootContext()));
        if (!rootObject)
            throw Exception("Could not create root object.");

        BridgeMonitor *bridgeMonitor = app().bridgeMonitor();
        bool bridgeExited = false;
        bool startError = false;
        QMetaObject::Connection connection;
        if (bridgeMonitor)
        {
            const BridgeMonitor::MonitorStatus& status = bridgeMonitor->getStatus();
            if (!status.running && !attach)
            {
                // BridgeMonitor already stopped meaning we are attached to an orphan Bridge.
                // Restart the full process to be sure there is no more bridge orphans
                app().log().error("Found orphan bridge, need to restart.");
                app().backend().forceLauncher(launcher);
                app().backend().restart();
                bridgeExited = true;
                startError = true;
            }
            else
            {
                app().log().debug(QString("Monitoring Bridge PID : %1").arg(status.pid));
                connection = QObject::connect(bridgeMonitor, &BridgeMonitor::processExited, [&](int returnCode) {
                    bridgeExited = true;// clazy:exclude=lambda-in-connect
                    qGuiApp->exit(returnCode);
                });
            }
        }

        int result = 0;
        if (!startError)
            result = QGuiApplication::exec();

        QObject::disconnect(connection);
        app().grpc().stopEventStream();

        // We manually delete the QML components to avoid warnings error due to order of deletion of C++ / JS objects and singletons.
        rootObject.reset();
        rootComponent.reset();

        if (!bridgeExited)
            closeBridgeApp();
        // release the lock file
        lock.unlock();
        return result;
    }
    catch (Exception const &e)
    {
        app().log().error(e.qwhat());
        return EXIT_FAILURE;
    }
}
