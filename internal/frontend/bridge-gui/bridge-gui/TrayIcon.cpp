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


#include "TrayIcon.h"
#include "QMLBackend.h"
#include <bridgepp/Exception/Exception.h>
#include <bridgepp/BridgeUtils.h>


using namespace bridgepp;


namespace {


QColor const normalColor(30, 168, 133); /// The normal state color.
QColor const errorColor(220, 50, 81); ///< The error state color.
QColor const warnColor(255, 153, 0); ///< The warn state color.
QColor const updateColor(35, 158, 206); ///< The warn state color.
QColor const greyColor(112, 109, 107); ///< The grey color.
qint64 const iconRefreshTimerIntervalMs = 1000; ///< The interval for the refresh timer when switching DPI / screen config, in milliseconds.
qint64 const iconRefreshDurationSecs = 10; ///< The total number of seconds during which we periodically refresh the icon after a DPI change.


//****************************************************************************************************************************************************
/// \brief Create a single resolution icon from an image. Throw an exception in case of failure.
//****************************************************************************************************************************************************
QIcon loadIconFromImage(QString const &path) {
    QPixmap const pixmap(path);
    if (pixmap.isNull()) {
        throw Exception(QString("Could not create an icon from an image '%1'.").arg(path));
    }
    return QIcon(pixmap);
}


//****************************************************************************************************************************************************
/// \brief Generate an icon from a SVG renderer (a.k.a. path).
///
/// \param[in] renderer The SVG renderer.
/// \param[in] color The color to use in case the SVG path is to be used as a mask.
/// \return The icon.
//****************************************************************************************************************************************************
QIcon loadIconFromSVGRenderer(QSvgRenderer &renderer, QColor const &color = QColor()) {
    if (!renderer.isValid()) {
        return QIcon();
    }
    QIcon icon;
    qint32 size = 256;

    while (size >= 16) {
        QPixmap pixmap(size, size);
        pixmap.fill(QColor(0, 0, 0, 0));
        QPainter painter(&pixmap);
        renderer.render(&painter);
        if (color.isValid()) {
            painter.setCompositionMode(QPainter::CompositionMode_SourceIn);
            painter.fillRect(pixmap.rect(), color);
        }
        painter.end();
        icon.addPixmap(pixmap);
        size /= 2;
    }

    return icon;
}


//****************************************************************************************************************************************************
/// \brief Load a multi-resolution icon from a SVG file. The image is assumed to be square. SVG is rasterized in 256, 128, 64, 32 and 16px.
///
/// Note: QPixmap can load SVG files directly, but our SVG file are defined in small shape size and QPixmap will rasterize them a very low resolution
/// by default (eg. 16x16), which is insufficient for some uses. As a consequence, we manually generate a multi-resolution icon that render smoothly
/// at any acceptable resolution for an icon.
///
/// \param[in] path The path of the SVG file.
/// \return The icon.
//****************************************************************************************************************************************************
QIcon loadIconFromSVG(QString const &path, QColor const &color = QColor()) {
    QSvgRenderer renderer(path);
    QIcon const icon = loadIconFromSVGRenderer(renderer, color);
    if (icon.isNull()) {
        Exception(QString("Could not create an icon from a vector image '%1'.").arg(path));
    }
    return icon;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
QIcon loadIcon(QString const &path) {
    if (path.endsWith(".svg", Qt::CaseInsensitive)) {
        return loadIconFromSVG(path);
    }
    return loadIconFromImage(path);
}


//****************************************************************************************************************************************************
/// \brief Retrieve the color associated with a tray icon state.
///
/// \param[in] state The state.
/// \return The color associated with a given tray icon state.
//****************************************************************************************************************************************************
QColor stateColor(TrayIcon::State state) {
    switch (state) {
    case TrayIcon::State::Normal:
        return normalColor;
    case TrayIcon::State::Error:
        return errorColor;
    case TrayIcon::State::Warn:
        return warnColor;
    case TrayIcon::State::Update:
        return updateColor;
    default:
        app().log().error(QString("Unknown tray icon state %1.").arg(static_cast<qint32>(state)));
        return normalColor;
    }
}


//****************************************************************************************************************************************************
/// \brief Return the text identifying a state in resource file names.
///
/// \param[in] state The state.
/// \param[in] The text identifying the state in resource file names.
//****************************************************************************************************************************************************
QString stateText(TrayIcon::State state) {
    switch (state) {
    case TrayIcon::State::Normal:
        return "norm";
    case TrayIcon::State::Error:
        return "error";
    case TrayIcon::State::Warn:
        return "warn";
    case TrayIcon::State::Update:
        return "update";
    default:
        app().log().error(QString("Unknown tray icon state %1.").arg(static_cast<qint32>(state)));
        return "norm";
    }
}


//****************************************************************************************************************************************************
/// \brief converts a QML resource path to Qt resource path.
/// QML resource paths are a bit different from qt resource paths
/// \param[in] path The resource path.
/// \return
//****************************************************************************************************************************************************
QString qmlResourcePathToQt(QString const &path) {
    QString result = path;
    result.replace(QRegularExpression(R"(^\.\/)"), ":/qml/");
    return result;
}

} // anonymous namespace


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TrayIcon::TrayIcon()
    : QSystemTrayIcon()
    , menu_(new QMenu)
    , notificationErrorIcon_(loadIconFromSVG(":/qml/icons/ic-alert.svg")) {
    this->generateDotIcons();
    this->setContextMenu(menu_.get());
    this->setToolTip(PROJECT_FULL_NAME);

    connect(menu_.get(), &QMenu::aboutToShow, this, &TrayIcon::onMenuAboutToShow);
    connect(this, &TrayIcon::selectUser, &app().backend(), [](QString const& userID, bool forceShowWindow) {
        app().backend().selectUser(userID, forceShowWindow);
    });
    connect(this, &TrayIcon::activated, this, &TrayIcon::onActivated);
    // some OSes/Desktop managers will automatically show main window when clicked, but not all, so we do it manually.
    if (!onLinux()) { // we disable this on linux because of a Qt bug that causes the signal to be emitted for other apps (GODT-2750)
        connect(this, &TrayIcon::messageClicked, []() { app().backend().showMainWindow("tray icon popup notification clicked"); });
    }

    this->setIcon();
    this->show();

    // TrayIcon does not expose its screen, so we connect relevant screen events to our DPI change handler.
    for (QScreen *screen: QGuiApplication::screens()) {
        connect(screen, &QScreen::logicalDotsPerInchChanged, this, &TrayIcon::handleDPIChange);
    }
    connect(qApp, &QApplication::screenAdded, [&](QScreen *screen) { connect(screen, &QScreen::logicalDotsPerInchChanged, this, &TrayIcon::handleDPIChange); });
    connect(qApp, &QApplication::primaryScreenChanged, [&](QScreen *screen) { connect(screen, &QScreen::logicalDotsPerInchChanged, this, &TrayIcon::handleDPIChange); });

    iconRefreshTimer_.setSingleShot(false);
    iconRefreshTimer_.setInterval(iconRefreshTimerIntervalMs);
    connect(&iconRefreshTimer_, &QTimer::timeout, this, &TrayIcon::onIconRefreshTimer);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::onMenuAboutToShow() {
    this->refreshContextMenu();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::onUserClicked() {
    try {
        auto action = dynamic_cast<QAction *>(this->sender());
        if (!action) {
            throw Exception("Could not retrieve context menu action.");
        }

        QString const &userID = action->data().toString();
        if (userID.isNull()) {
            throw Exception("Could not retrieve context menu's selected user.");
        }

        app().backend().selectUser(userID, true, "tray menu user clicked");
    } catch (Exception const &e) {
        app().log().error(e.qwhat());
    }
}


//****************************************************************************************************************************************************
/// \param[in] reason The icon activation reason.
//****************************************************************************************************************************************************
void TrayIcon::onActivated(QSystemTrayIcon::ActivationReason reason) {
    if ((QSystemTrayIcon::Trigger == reason) && !onMacOS()) {
        app().backend().showMainWindow("tray icon activated");
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::handleDPIChange() {
    this->setIcon();

    // Windows forces us to apply a hack. Tray icon does not redraw by itself, so we use the Qt signal that detects screen and DPI changes.
    // But the moment we get the signal the DPI change is not yet in effect. so redrawing now will have no effect, and we don't really
    // know when we can safely redraw. So we will redraw the icon every second for some time.
    iconRefreshDeadline_ = QDateTime::currentDateTime().addSecs(iconRefreshDurationSecs);
    if (!iconRefreshTimer_.isActive()) {
        iconRefreshTimer_.start();
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::setIcon() {
    QString const style = onMacOS() ? "mono" : "color";
    QString const text = stateText(state_);

    QIcon icon = loadIconFromImage(QString(":/qml/icons/systray-%1-%2.png").arg(style, text));
    icon.setIsMask(true);
    QSystemTrayIcon::setIcon(icon);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::onIconRefreshTimer() {
    this->setIcon();
    if (QDateTime::currentDateTime() > iconRefreshDeadline_) {
        iconRefreshTimer_.stop();
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::generateDotIcons() {
    QSvgRenderer dotSVG(QString(":/qml/icons/ic-dot.svg"));

    struct IconColor {
        QIcon &icon;
        QColor color;
    };
    for (auto pair: QList<IconColor> {{ greenDot_, normalColor }, { greyDot_, greyColor }, { orangeDot_, warnColor }}) {
        pair.icon = loadIconFromSVGRenderer(dotSVG, pair.color);
        if (pair.icon.isNull()) {
            throw Exception("Could not generate dot icon from vector file.");
        }
    }
}


//****************************************************************************************************************************************************
/// \param[in] state The state.
/// \param[in] stateString A string describing the state.
/// \param[in] statusIconPath The status icon path.
/// \param[in] statusIconColor The color for the status icon in hex.
//****************************************************************************************************************************************************
void TrayIcon::setState(TrayIcon::State state, QString const &stateString, QString const &statusIconPath) {
    stateString_ = stateString;
    state_ = state;
    this->setIcon();
    this->generateStatusIcon(statusIconPath, stateColor(state));
}


//****************************************************************************************************************************************************
/// \param[in] title The title.
/// \param[in] message The message.
//****************************************************************************************************************************************************
void TrayIcon::showErrorPopupNotification(QString const &title, QString const &message) {
    this->showMessage(title, message, notificationErrorIcon_);
}

//****************************************************************************************************************************************************
/// Used only by user notifications received from the event loop
/// \param[in] title The title.
/// \param[in] subtitle The subtitle.
//****************************************************************************************************************************************************
void TrayIcon::showUserNotification(QString const &title, QString const &subtitle) {
    this->showMessage(title, subtitle, QSystemTrayIcon::NoIcon);
}


//****************************************************************************************************************************************************
/// \param[in] svgPath The path of the SVG file for the icon.
/// \param[in] color The color to apply to the icon.
//****************************************************************************************************************************************************
void TrayIcon::generateStatusIcon(QString const &svgPath, QColor const &color) {
    // We use the SVG path as pixmap mask and fill it with the appropriate color
    statusIcon_ = loadIconFromSVG(qmlResourcePathToQt(svgPath), color);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TrayIcon::refreshContextMenu() {
    if (!menu_) {
        app().log().error("Native tray icon context menu is null.");
        return;
    }

    bool const internetOn = app().backend().isInternetOn();
    menu_->clear();
    menu_->addAction(statusIcon_, stateString_, []() {app().backend().showMainWindow("tray menu status clicked");});
    menu_->addSeparator();
    QKeySequence noShortcut;
    UserList const &users = app().backend().users();
    qint32 const userCount = users.count();
    bool const onMac = onMacOS();
    for (qint32 i = 0; i < userCount; i++) {
        User const &user = *users.get(i);
        UserState const state = user.state();
        auto action = new QAction(user.primaryEmailOrUsername());
        if (internetOn) {
            action->setIcon((UserState::Connected == state) ? greenDot_ : (UserState::Locked == state ? orangeDot_ : greyDot_));
        }
        action->setData(user.id());
        connect(action, &QAction::triggered, this, &TrayIcon::onUserClicked);
        if ((i < 10) && onMac) {
            action->setShortcut(QKeySequence(QString("Ctrl+%1").arg((i + 1) % 10)));
        }
        menu_->addAction(action);
    }
    if (userCount) {
        menu_->addSeparator();
    }

    menu_->addAction(tr("&Open Bridge"), onMac ? QKeySequence("Ctrl+O") : noShortcut, []() {
        app().backend().showMainWindow("tray menu 'open bridge' clicked");
    });
    menu_->addAction(tr("&Help"), onMac ? QKeySequence("Ctrl+F1") : noShortcut, []() {
        app().backend().showHelp("tray menu 'Help' clicked");
    });
    menu_->addAction(tr("&Settings"), onMac ? QKeySequence("Ctrl+,") : noShortcut, []() {
        app().backend().showSettings("tray menu 'Settings' clicked");
    });
    menu_->addSeparator();
    menu_->addAction(tr("&Quit Bridge"), onMac ? QKeySequence("Ctrl+Q") : noShortcut, &app().backend(), &QMLBackend::quit);
}
