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

#include "BridgeApp.h"
#include "BuildConfig.h"
#include "CommandLine.h"
#include "LogUtils.h"
#include "QMLBackend.h"
#include "SentryUtils.h"
#include "Settings.h"
#include <bridgepp/BridgeUtils.h>
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/FocusGRPC/FocusGRPCClient.h>
#include <bridgepp/Log/Log.h>
#include <bridgepp/Log/LogUtils.h>
#include <bridgepp/ProcessMonitor.h>
#include <ClipboardProxy.h>

#include "bridgepp/CLI/CLIUtils.h"

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
qint64 constexpr grpcServiceConfigWaitDelayMs = 180000; ///< The wait delay for the gRPC config file in milliseconds.
QString const waitFlag = "--wait"; ///< The wait command-line flag.
QString const orphanInstanceException =  "An orphan instance of bridge is already running. Please terminate it and relaunch the application.";

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
        QString const &err = rootComponent->errorString();
        app().log().error(err);
        throw Exception("Could not load QML component", err);
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
        if (lock.error() == QLockFile::LockFailedError) {
            // This happens if a stale lock file exists and another process uses that PID.
            // Try removing the stale file, which will fail if a real process is holding a
            // file-level lock. A false error is more problematic than not locking properly
            // on corner-case systems.
            if (lock.removeStaleLockFile() && lock.tryLock()) {
                app().log().info("Removed stale lock file");
                app().log().info(QString("lock file created %1").arg(lock.fileName()));
                return true;
            }
        }
        app().log().error(QString("Instance already exists %1 %2").arg(lock.fileName(), details));
        return false;
    }
    app().log().info(QString("lock file created %1").arg(lock.fileName()));
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
    QLockFile lockFile(QDir(bridgepp::userCacheDir()).absoluteFilePath(bridgeLock));
    return (!lockFile.tryLock()) && (lockFile.error() == QLockFile::LockFailedError);
}

//****************************************************************************************************************************************************
/// \brief Use api to bring focus on existing bridge instance.
//****************************************************************************************************************************************************
void focusOtherInstance() {
    try {
        FocusGRPCClient client(app().log());
        GRPCConfig sc;
        QString const path = FocusGRPCClient::grpcFocusServerConfigPath(bridgepp::userConfigDir());
        QFile file(path);
        if (file.exists()) {
            if (!sc.load(path)) {
                throw Exception("The gRPC focus service configuration file is invalid.");
            }
        } else {
            throw Exception("Server did not provide gRPC Focus service configuration.");
        }

        QString error;
        if (!client.connectToServer(5000, sc.port, &error)) {
            throw Exception("Could not connect to bridge focus service for a raise call.", error);
        }
        if (!client.raise("focusOtherInstance").ok()) {
            throw Exception(QString("The raise call to the bridge focus service failed."));
        }
    } catch (Exception const &e) {
        app().log().error(e.qwhat());
        auto uuid = reportSentryException("Exception occurred during focusOtherInstance()", e);
        app().log().fatal(QString("reportID: %1 Captured exception: %2").arg(QByteArray(uuid.bytes, 16).toHex(), e.qwhat()));
    }
}

//****************************************************************************************************************************************************
/// \param [in] args list of arguments to pass to bridge.
/// \return bridge executable path
//****************************************************************************************************************************************************
QString launchBridge(QStringList const &args) {
    UPOverseer &overseer = app().bridgeOverseer();
    overseer.reset();

    const QString bridgeExePath = locateBridgeExe();

    if (bridgeExePath.isEmpty()) {
        throw Exception("Could not locate the bridge executable path");
    } else {
        app().log().debug(QString("Bridge executable path: %1").arg(QDir::toNativeSeparators(bridgeExePath)));
    }

    qint64 const pid = qApp->applicationPid();
    QStringList const params = QStringList{"--grpc", "--parent-pid", QString::number(pid)} + args;
    app().log().info(QString("Launching bridge process with command \"%1\" %2").arg(bridgeExePath, params.join(" ")));
    overseer = std::make_unique<Overseer>(new ProcessMonitor(bridgeExePath, params, nullptr), nullptr);
    overseer->startWorker(true);
    return bridgeExePath;
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void closeBridgeApp() {
    app().grpc().quit(); // this will cause the grpc service and the bridge app to close.

    UPOverseer const &overseer = app().bridgeOverseer();
    if (overseer) {
        // A null overseer means the app was run in 'attach' mode. We're not monitoring it.
        // ReSharper disable once CppExpressionWithoutSideEffects
        overseer->wait(Overseer::maxTerminationWaitTimeMs);
    }
}

//****************************************************************************************************************************************************
/// \param[in] argv The command-line argments, including the application name at index 0.
//****************************************************************************************************************************************************
void logCommandLineInvocation(QStringList argv) {
    Log &log = app().log();
    if (argv.isEmpty()) {
        log.error("The command line is empty");
    }
    log.info("bridge-gui executable: " + argv.front());
    log.info("Command-line invocation: " + (argv.size() > 1 ? argv.last(argv.size() - 1).join(" ") : "<none>"));
}

//****************************************************************************************************************************************************
/// \param[in] argc The number of command-line arguments.
/// \param[in] argv The list of command-line arguments.
/// \return The exit code for the application.
//****************************************************************************************************************************************************
int main(int argc, char *argv[]) {
    // The application instance is needed to display system message boxes. As we may have to do it in the exception handler,
    // application instance is create outside the try/catch clause.
    if (QSysInfo::productType() != "windows") {
        QCoreApplication::setAttribute(Qt::AA_UseSoftwareOpenGL); // must be called before instantiating the BridgeApp
    }

    BridgeApp guiApp(argc, argv);
    initSentry();
    auto sentryCloser = qScopeGuard([] { sentry_close(); });

    try {
        QString const &configDir = bridgepp::userConfigDir();

        initQtApplication();
        QStringList const argvList = cliArgsToStringList(argc, argv);
        CommandLineOptions const cliOptions = parseCommandLine(argvList);
        Log &log = initLog();
        log.setLevel(cliOptions.logLevel);

        QLockFile lock(bridgepp::userCacheDir() + "/" + bridgeGUILock);
        if (!checkSingleInstance(lock)) {
            focusOtherInstance();
            return EXIT_FAILURE;
        }

#ifdef Q_OS_MACOS
        registerSecondInstanceHandler();
        setDockIconVisibleState(!cliOptions.noWindow);
#endif

        logCommandLineInvocation(argvList);

        // In attached mode, we do not intercept stderr and stdout of bridge, as we did not launch it ourselves, so we output the log to the console.
        // When not in attached mode, log entries are forwarded to bridge, which output it on stdout/stderr. bridge-gui's process monitor intercept
        // these outputs and output them on the command-line.
        log.info(QString("New Sentry reporter - id: %1.").arg(getProtectedHostname()));

        QString const &sessionID = app().sessionID();
        QString bridgeExe;
        if (!cliOptions.attach) {
            if (isBridgeRunning()) {
                throw Exception(orphanInstanceException,
                    QString(), __FUNCTION__, tailOfLatestBridgeLog(sessionID));
            }

            // before launching bridge, we remove any trailing service config file, because we need to make sure we get a newly generated one.
            FocusGRPCClient::removeServiceConfigFile(configDir);
            GRPCClient::removeServiceConfigFile(configDir);
            bridgeExe = launchBridge(cliOptions.bridgeArgs);
        }

        log.info(QString("Retrieving gRPC service configuration from '%1'").arg(QDir::toNativeSeparators(grpcServerConfigPath(configDir))));
        app().backend().init(GRPCClient::waitAndRetrieveServiceConfig(sessionID, configDir,
            cliOptions.attach ? 0 : grpcServiceConfigWaitDelayMs, app().bridgeMonitor()));
        if (!cliOptions.attach) {
            GRPCClient::removeServiceConfigFile(configDir);
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
        QQuickWindow::setSceneGraphBackend((app().settings().useSoftwareRenderer() || cliOptions.useSoftwareRenderer) ? "software" : "rhi");
        log.info(QString("Qt Quick renderer: %1").arg(QQuickWindow::sceneGraphBackend()));

        QQmlApplicationEngine engine;
        // Set up clipboard
        engine.rootContext()->setContextProperty("clipboard", new ClipboardProxy(QGuiApplication::clipboard()));
        std::unique_ptr<QQmlComponent> rootComponent(createRootQmlComponent(engine));
        std::unique_ptr<QObject> rootObject(rootComponent->create(engine.rootContext()));
        if (!rootObject) {
            throw Exception("Could not create QML root object.");
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
                    bridgeExited = true; // clazy:exclude=lambda-in-connect
                    qGuiApp->exit(returnCode);
                });
            }
        }

        int result = 0;
        if (!startError) {
            // we succeeded in launching bridge, so we can be set as mainExecutable.
            QString const mainexec = argvList[0];
            app().grpc().setMainExecutable(mainexec);
            QStringList args = cliOptions.bridgeGuiArgs;
            args.append(waitFlag);
            args.append(mainexec);
            if (!bridgeExe.isEmpty()) {
                args.append(waitFlag);
                args.append(bridgeExe);
            }
            app().setLauncherArgs(cliOptions.launcher, args);
            result = QGuiApplication::exec();
        }

        QObject::disconnect(connection);
        app().grpc().stopEventStreamReader();
        if (!app().backend().waitForEventStreamReaderToFinish(Overseer::maxTerminationWaitTimeMs)) {
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
    } catch (Exception const &e) {
        QString message = e.qwhat();
        if (e.showSupportLink()) {
            message += R"(<br/><br/>If the issue persists, please contact our <a href="https://proton.me/support/contact">customer support</a>.)";
        }
        QMessageBox::critical(nullptr, "Error", message);

        if (e.qwhat() != orphanInstanceException) {
            sentry_uuid_s const uuid = reportSentryException("Exception occurred during main", e);
            QTextStream(stderr) << "reportID: " << QByteArray(uuid.bytes, 16).toHex() << " Captured exception :"
                                << e.detailedWhat() << "\n";
        }

        return EXIT_FAILURE;
    }
}
