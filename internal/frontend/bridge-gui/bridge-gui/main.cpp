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


#include "Pch.h"
#include "BridgeApp.h"
#include "CommandLine.h"
#include "QMLBackend.h"
#include "SentryUtils.h"
#include "Version.h"
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/FocusGRPC/FocusGRPCClient.h>
#include <bridgepp/Log/Log.h>
#include <bridgepp/ProcessMonitor.h>
#include <sentry.h>
#include <project_sentry_config.h>


#ifdef Q_OS_MACOS


#include "MacOS/SecondInstance.h"


#endif

using namespace bridgepp;

namespace {

/// \brief The file extension for the bridge executable file.
#ifdef Q_OS_WIN32
QString const exeSuffix = ".exe";
#else
QString const exeSuffix;
#endif


QString const bridgeLock = "bridge-v3.lock"; ///< The file name used for the bridge-gui lock file.
QString const bridgeGUILock = "bridge-v3-gui.lock"; ///< The file name used for the bridge-gui lock file.
QString const exeName = "bridge" + exeSuffix; ///< The bridge executable file name.*
qint64 const grpcServiceConfigWaitDelayMs = 180000; ///< The wait delay for the gRPC config file in milliseconds.


} // anonymous namespace


//****************************************************************************************************************************************************
/// \return The path of the bridge executable.
/// \return A null string if the executable could not be located.
//****************************************************************************************************************************************************
QString locateBridgeExe() {
    QFileInfo const fileInfo(QDir(QCoreApplication::applicationDirPath()).absoluteFilePath(exeName));
    return (fileInfo.exists() && fileInfo.isFile() && fileInfo.isExecutable()) ? fileInfo.absoluteFilePath() : QString();
}


//****************************************************************************************************************************************************
/// // initialize the Qt application.
//****************************************************************************************************************************************************
void initQtApplication() {
    QString const qsgInfo = QProcessEnvironment::systemEnvironment().value("QSG_INFO");
    if ((!qsgInfo.isEmpty()) && (qsgInfo != "0")) {
        QLoggingCategory::setFilterRules("qt.scenegraph.general=true");
    }

    QGuiApplication::setApplicationName(PROJECT_FULL_NAME);
    QGuiApplication::setApplicationVersion(PROJECT_VER);
    QGuiApplication::setOrganizationName(PROJECT_VENDOR);
    QGuiApplication::setOrganizationDomain("proton.ch");
    QGuiApplication::setQuitOnLastWindowClosed(false);
#ifdef Q_OS_MACOS
    // on macOS, the app icon as it appears in the dock and file system is defined by in the app bundle plist, not here.
    // We still use this copy (lock icon in white rectangle), so that devs that use the bridge-gui exe directly get a decent looking icon in the dock.
    // Qt does not support the native .icns format, so we use a PNG file.
    QGuiApplication::setWindowIcon(QIcon(":bridgeMacOS.svg"));
#else
    // On non macOS platform, this icon (without the white rectangle background, is used in the OS decoration elements (title bar, task bar, etc...)
    // It's not use as the executable icon.
    QGuiApplication::setWindowIcon(QIcon(":bridge.svg"));
#endif // #ifdef Q_OS_MACOS
}


//****************************************************************************************************************************************************
/// \return A reference to the log.
//****************************************************************************************************************************************************
Log &initLog() {
    Log &log = app().log();
    log.registerAsQtMessageHandler();
    log.setEchoInConsole(true);

    // remove old gui log files
    QDir const logsDir(userLogsDir());
    for (QFileInfo const fileInfo: logsDir.entryInfoList({ "gui_v*.log" }, QDir::Filter::Files)) { // entryInfolist apparently only support wildcards, not regex.
        QFile(fileInfo.absoluteFilePath()).remove();
    }

    // create new GUI log file
    QString error;
    if (!log.startWritingToFile(logsDir.absoluteFilePath(QString("gui_v%1_%2.log").arg(PROJECT_VER).arg(QDateTime::currentSecsSinceEpoch())), &error)) {
        log.error(error);
    }

    log.info("bridge-gui starting");
    QString const qtCompileTimeVersion = QT_VERSION_STR;
    QString const qtRuntimeVersion = qVersion();
    QString msg = QString("Using Qt %1").arg(qtRuntimeVersion);
    if (qtRuntimeVersion != qtCompileTimeVersion) {
        msg += QString(" (compiled against %1)").arg(qtCompileTimeVersion);
    }
    log.info(msg);

    return log;
}


//****************************************************************************************************************************************************
/// \param[in] engine The QML component.
//****************************************************************************************************************************************************
QQmlComponent *createRootQmlComponent(QQmlApplicationEngine &engine) {
    QString const qrcQmlDir = "qrc:/qml";
    qmlRegisterSingletonInstance("Proton", 1, 0, "Backend", &app().backend());
    qmlRegisterType<UserList>("Proton", 1, 0, "UserList");
    qmlRegisterType<bridgepp::User>("Proton", 1, 0, "User");
    qRegisterMetaType<UserState>("UserState");
    qmlRegisterUncreatableType<EUserState>("Proton", 1, 0, "EUserState", "Enum type is not creatable");
    auto rootComponent = new QQmlComponent(&engine, &engine);

    engine.addImportPath(qrcQmlDir);
    engine.addPluginPath(qrcQmlDir);
    QQuickStyle::setStyle("Proton");

    rootComponent->loadUrl(QUrl(qrcQmlDir + "/Bridge.qml"));
    if (rootComponent->status() != QQmlComponent::Status::Ready) {
        app().log().error(rootComponent->errorString());
        throw Exception("Could not load QML component");
    }
    return rootComponent;
}


//****************************************************************************************************************************************************
/// \param[in] lock The lock file to be checked.
/// \return True if the lock can be taken, false otherwise.
//****************************************************************************************************************************************************
bool checkSingleInstance(QLockFile &lock) {
    lock.setStaleLockTime(0);
    if (!lock.tryLock()) {
        qint64 pid;
        QString hostname, appName, details;
        if (lock.getLockInfo(&pid, &hostname, &appName)) {
            details = QString("(PID : %1 - Host : %2 - App : %3)").arg(pid).arg(hostname, appName);
        }

        app().log().error(QString("Instance already exists %1 %2").arg(lock.fileName(), details));
        return false;
    } else {
        app().log().info(QString("lock file created %1").arg(lock.fileName()));
    }
    return true;
}


//****************************************************************************************************************************************************
/// \return QUrl to reach the bridge API.
//****************************************************************************************************************************************************
QUrl getApiUrl() {
    QUrl url;
    // use default url.
    url.setScheme("http");
    url.setHost("127.0.0.1");
    url.setPort(1042);

    // override with what can be found in the prefs.json file.
    QFile prefFile(QString("%1/%2").arg(bridgepp::userConfigDir(), "prefs.json"));
    if (prefFile.exists()) {
        prefFile.open(QIODevice::ReadOnly | QIODevice::Text);
        QByteArray data = prefFile.readAll();
        prefFile.close();
        QJsonDocument doc = QJsonDocument::fromJson(data);
        if (!doc.isNull()) {
            QString userPortApi = "user_port_api";
            QJsonObject obj = doc.object();
            if (!obj.isEmpty() && obj.contains(userPortApi)) {
                url.setPort(doc.object()[userPortApi].toString().toInt());
            }
        }
    }
    return url;
}


//****************************************************************************************************************************************************
/// \brief Check if bridge is running.
///
/// The check is performed by trying to create a lock file for bridge. Priority is to avoid false positive, so we only return true if the locking
/// attempt failed because the file is locked. Any other error (PermissionError, UnknownError) do not lead to considering bridge is running.
/// \note QLockFile removes the lock file in it destructor. So it's removed on function exit if acquired.
///
/// \return true if an instance of bridge is already running.
//****************************************************************************************************************************************************
bool isBridgeRunning() {
    QLockFile lockFile(QDir(userCacheDir()).absoluteFilePath(bridgeLock));
    return (!lockFile.tryLock()) && (lockFile.error() == QLockFile::LockFailedError);
}


//****************************************************************************************************************************************************
/// \brief Use api to bring focus on existing bridge instance.
//****************************************************************************************************************************************************
void focusOtherInstance() {
    try {
        FocusGRPCClient client;
        QString error;
        if (!client.connectToServer(5000, &error)) {
            throw Exception(QString("Could not connect to bridge focus service for a raise call: %1").arg(error));
        }
        if (!client.raise().ok()) {
            throw Exception(QString("The raise call to the bridge focus service failed."));
        }
    }
    catch (Exception const &e) {
        app().log().error(e.qwhat());
        reportSentryException(SENTRY_LEVEL_ERROR, "Exception occurred during focusOtherInstance()", "Exception", e.what());
    }
}


//****************************************************************************************************************************************************
/// \param [in] args list of arguments to pass to bridge.
//****************************************************************************************************************************************************
void launchBridge(QStringList const &args) {
    UPOverseer &overseer = app().bridgeOverseer();
    overseer.reset();

    const QString bridgeExePath = locateBridgeExe();

    if (bridgeExePath.isEmpty()) {
        throw Exception("Could not locate the bridge executable path");
    } else {
        app().log().debug(QString("Bridge executable path: %1").arg(QDir::toNativeSeparators(bridgeExePath)));
    }

    qint64 const pid = qApp->applicationPid();
    QStringList const params = QStringList { "--grpc", "--parent-pid", QString::number(pid) } + args;
    app().log().info(QString("Launching bridge process with command \"%1\" %2").arg(bridgeExePath, params.join(" ")));
    overseer = std::make_unique<Overseer>(new ProcessMonitor(bridgeExePath, params, nullptr), nullptr);
    overseer->startWorker(true);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void closeBridgeApp() {
    app().grpc().quit(); // this will cause the grpc service and the bridge app to close.

    UPOverseer &overseer = app().bridgeOverseer();
    if (!overseer) { // The app was run in 'attach' mode and attached to an existing instance of Bridge. We're not monitoring it.
        return;
    }

    while (!overseer->isFinished()) {
        QThread::msleep(20);
    }
}


//****************************************************************************************************************************************************
/// \param[in] argc The number of command-line arguments.
/// \param[in] argv The list of command-line arguments.
/// \return The exit code for the application.
//****************************************************************************************************************************************************
int main(int argc, char *argv[]) {
    // Init sentry.
    sentry_options_t *sentryOptions = sentry_options_new();
    sentry_options_set_dsn(sentryOptions, SentryDNS);
    {
        const QString sentryCachePath = sentryCacheDir();
        sentry_options_set_database_path(sentryOptions, sentryCachePath.toStdString().c_str());
    }
    sentry_options_set_release(sentryOptions, SentryProductID);
    // Enable this for debugging sentry.
    // sentry_options_set_debug(sentryOptions, 1);
    if (sentry_init(sentryOptions) != 0) {
        std::cerr << "Failed to initialize sentry" << std::endl;
    }

    auto sentryClose = qScopeGuard([] { sentry_close(); });

    // The application instance is needed to display system message boxes. As we may have to do it in the exception handler,
    // application instance is create outside the try/catch clause.
    if (QSysInfo::productType() != "windows") {
        QCoreApplication::setAttribute(Qt::AA_UseSoftwareOpenGL);
    }

    BridgeApp guiApp(argc, argv);

    try {
        initQtApplication();

        Log &log = initLog();

        QLockFile lock(bridgepp::userCacheDir() + "/" + bridgeGUILock);
        if (!checkSingleInstance(lock)) {
            focusOtherInstance();
            return EXIT_FAILURE;
        }

        CommandLineOptions const cliOptions = parseCommandLine(argc, argv);

#ifdef Q_OS_MACOS
        registerSecondInstanceHandler();
        setDockIconVisibleState(!cliOptions.noWindow);
#endif

        // In attached mode, we do not intercept stderr and stdout of bridge, as we did not launch it ourselves, so we output the log to the console.
        // When not in attached mode, log entries are forwarded to bridge, which output it on stdout/stderr. bridge-gui's process monitor intercept
        // these outputs and output them on the command-line.
        log.setLevel(cliOptions.logLevel);

        if (!cliOptions.attach) {
            if (isBridgeRunning()) {
                throw Exception("An orphan instance of bridge is already running. Please terminate it and relaunch the application.");
            }

            // before launching bridge, we remove any trailing service config file, because we need to make sure we get a newly generated one.
            GRPCClient::removeServiceConfigFile();
            launchBridge(cliOptions.bridgeArgs);
        }

        log.info(QString("Retrieving gRPC service configuration from '%1'").arg(QDir::toNativeSeparators(grpcServerConfigPath())));
        app().backend().init(GRPCClient::waitAndRetrieveServiceConfig(cliOptions.attach ? 0 : grpcServiceConfigWaitDelayMs, app().bridgeMonitor()));
        if (!cliOptions.attach) {
            GRPCClient::removeServiceConfigFile();
        }

        // gRPC communication is established. From now on, log events will be sent to bridge via gRPC. bridge will write these to file,
        // and will output then on console if appropriate. If we are not running in attached mode we intercept bridge stdout & stderr and
        // display it in our own output and error, so we only continue to log directly to console if we are running in attached mode.
        log.setEchoInConsole(cliOptions.attach);
        log.info("Backend was successfully initialized.");
        log.stopWritingToFile();

        // The following allows to render QML content in software with a 'Rendering Hardware Interface' (OpenGL, Vulkan, Metal, Direct3D...)
        // Note that it is different from the Qt::AA_UseSoftwareOpenGL attribute we use on some platforms that instruct Qt that we would like
        // to use a software-only implementation of OpenGL.
        QQuickWindow::setSceneGraphBackend(cliOptions.useSoftwareRenderer ? "software" : "rhi");
        log.info(QString("Qt Quick renderer: %1").arg(QQuickWindow::sceneGraphBackend()));

        QQmlApplicationEngine engine;
        std::unique_ptr<QQmlComponent> rootComponent(createRootQmlComponent(engine));
        std::unique_ptr<QObject> rootObject(rootComponent->create(engine.rootContext()));
        if (!rootObject) {
            throw Exception("Could not create root object.");
        }

        ProcessMonitor *bridgeMonitor = app().bridgeMonitor();
        bool bridgeExited = false;
        bool startError = false;
        QMetaObject::Connection connection;
        if (bridgeMonitor) {
            const ProcessMonitor::MonitorStatus &status = bridgeMonitor->getStatus();
            if (status.ended && !cliOptions.attach) {
                // ProcessMonitor already stopped meaning we are attached to an orphan Bridge.
                // Restart the full process to be sure there is no more bridge orphans
                app().log().error("Found orphan bridge, need to restart.");
                app().backend().forceLauncher(cliOptions.launcher);
                app().backend().restart();
                bridgeExited = true;
                startError = true;
            } else {
                app().log().debug(QString("Monitoring Bridge PID : %1").arg(status.pid));

                connection = QObject::connect(bridgeMonitor, &ProcessMonitor::processExited, [&](int returnCode) {
                    bridgeExited = true;// clazy:exclude=lambda-in-connect
                    qGuiApp->exit(returnCode);
                });
            }
        }

        int result = 0;
        if (!startError) {
            // we succeeded in launching bridge, so we can be set as mainExecutable.
            app().grpc().setMainExecutable(QString::fromLocal8Bit(argv[0]));
            result = QGuiApplication::exec();
        }

        QObject::disconnect(connection);
        app().grpc().stopEventStreamReader();
        if (!app().backend().waitForEventStreamReaderToFinish(5000)) {
            log.warn("Event stream reader took too long to finish.");
        }

        // We manually delete the QML components to avoid warnings error due to order of deletion of C++ / JS objects and singletons.
        rootObject.reset();
        rootComponent.reset();

        if (!bridgeExited) {
            closeBridgeApp();
        }
        // release the lock file
        lock.unlock();
        return result;
    }
    catch (Exception const &e) {
        reportSentryException(SENTRY_LEVEL_ERROR, "Exception occurred during main", "Exception", e.what());
        QMessageBox::critical(nullptr, "Error", e.qwhat());
        QTextStream(stderr) << e.qwhat() << "\n";
        return EXIT_FAILURE;
    }
}
