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
qint64 const iconRefreshDurationSecs = 10; ///< The total number of seconds during wich we periodically refresh the icon after a DPI change.


//****************************************************************************************************************************************************
/// \brief Create a single resolution icon from an image. Throw an exception in case of failure.
//****************************************************************************************************************************************************
QIcon loadIconFromImage(QString const &path) {
    QPixmap const pixmap(path);
    if (pixmap.isNull()) {
        throw Exception(QString("Could create icon from image '%1'.").arg(path));
    }
    return QIcon(pixmap);
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


} // anonymous namespace


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TrayIcon::TrayIcon()
    : QSystemTrayIcon()
    , menu_(new QMenu) {

    this->generateDotIcons();
    this->setContextMenu(menu_.get());

    connect(menu_.get(), &QMenu::aboutToShow, this, &TrayIcon::onMenuAboutToShow);
    connect(this, &TrayIcon::selectUser, &app().backend(), &QMLBackend::selectUser);
    connect(this, &TrayIcon::activated, this, &TrayIcon::onActivated);

    this->show();
    this->setState(State::Normal, QString(), QString());

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

        emit selectUser(userID);
    } catch (Exception const &e) {
        app().log().error(e.qwhat());
    }
}


//****************************************************************************************************************************************************
/// \param[in] reason The icon activation reason.
//****************************************************************************************************************************************************
void TrayIcon::onActivated(QSystemTrayIcon::ActivationReason reason) {
    if ((QSystemTrayIcon::Trigger == reason) && !onMacOS()) {
        app().backend().showMainWindow();
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
    QPixmap dotSVG(":/qml/icons/ic-dot.svg");
    struct IconColor {
        QIcon &icon;
        QColor color;
    };
    for (auto pair: QList<IconColor> {{ greenDot_, normalColor }, { greyDot_, greyColor }, { orangeDot_, warnColor }}) {
        QPixmap p = dotSVG;
        QPainter painter(&p);
        painter.setCompositionMode(QPainter::CompositionMode_SourceIn);
        painter.fillRect(p.rect(), pair.color);
        painter.end();
        pair.icon = QIcon(p);
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
/// \param[in] svgPath The path of the SVG file for the icon.
/// \param[in] color The color to apply to the icon.
//****************************************************************************************************************************************************
void TrayIcon::generateStatusIcon(QString const &svgPath, QColor const &color) {
    // We use the SVG path as pixmap mask and fill it with the appropriate color
    QString resourcePath = svgPath;
    resourcePath.replace(QRegularExpression(R"(^\.\/)"), ":/qml/"); // QML resource path are a bit different from the Qt resources path.
    QPixmap pixmap(resourcePath);
    QPainter painter(&pixmap);
    painter.setCompositionMode(QPainter::CompositionMode_SourceIn);
    painter.fillRect(pixmap.rect(), color);
    painter.end();
    statusIcon_ = QIcon(pixmap);
}


//**********************************************************************************************************************
//
//**********************************************************************************************************************
void TrayIcon::refreshContextMenu() {
    if (!menu_) {
        app().log().error("Native tray icon context menu is null.");
        return;
    }

    menu_->clear();
    menu_->addAction(statusIcon_, stateString_, &app().backend(), &QMLBackend::showMainWindow);
    menu_->addSeparator();
    UserList const &users = app().backend().users();
    qint32 const userCount = users.count();
    for (qint32 i = 0; i < userCount; i++) {
        User const &user = *users.get(i);
        UserState const state = user.state();
        auto action = new QAction(user.primaryEmailOrUsername());
        action->setIcon((UserState::Connected == state) ? greenDot_ : (UserState::Locked == state ? orangeDot_ : greyDot_));
        action->setData(user.id());
        connect(action, &QAction::triggered, this, &TrayIcon::onUserClicked);
        if (i < 10) {
            action->setShortcut(QKeySequence(QString("Ctrl+%1").arg((i + 1) % 10)));
        }
        menu_->addAction(action);
    }
    if (userCount) {
        menu_->addSeparator();
    }
    menu_->addAction(tr("&Open Bridge"), QKeySequence("Ctrl+O"), &app().backend(), &QMLBackend::showMainWindow);
    menu_->addAction(tr("&Help"), QKeySequence("Ctrl+F1"), &app().backend(), &QMLBackend::showHelp);
    menu_->addAction(tr("&Settings"), QKeySequence("Ctrl+,"), &app().backend(), &QMLBackend::showSettings);
    menu_->addSeparator();
    menu_->addAction(tr("&Quit Bridge"), QKeySequence("Ctrl+Q"), &app().backend(), &QMLBackend::quit);
}
