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


#ifndef BRIDGE_QT6_USER_H
#define BRIDGE_QT6_USER_H


#include "Log.h"


//****************************************************************************************************************************************************
/// \brief User class.
//****************************************************************************************************************************************************
class User : public QObject
{
Q_OBJECT
public: // member functions.
    explicit User(QObject *parent = nullptr); ///< Default constructor.
    User(User const &) = delete; ///< Disabled copy-constructor.
    User(User &&) = delete; ///< Disabled assignment copy-constructor.
    ~User() override = default; ///< Destructor.
    User &operator=(User const &) = delete; ///< Disabled assignment operator.
    User &operator=(User &&) = delete; ///< Disabled move assignment operator.
    void update(User const &user); ///< Update the user
public slots:
    // slots for QML generated calls
    void toggleSplitMode(bool makeItActive);            //    _ func(makeItActive bool) `slot:"toggleSplitMode"`
    void logout();                                      //    _ func()                  `slot:"logout"`
    void remove();                                      //    _ func()                  `slot:"remove"`
    void configureAppleMail(QString const &address);    //    _ func(address string)    `slot:"configureAppleMail"`

    // slots for external signals
    void emitToggleSplitModeFinished();

public:
    Q_PROPERTY(QString username MEMBER username_ NOTIFY usernameChanged)                //    _ string   `property:"username"`
    Q_PROPERTY(QString avatarText MEMBER avatarText_ NOTIFY avatarTextChanged)          //    _ string   `property:"avatarText"`
    Q_PROPERTY(bool loggedIn MEMBER loggedIn_ NOTIFY loggedInChanged)                   //    _ bool     `property:"loggedIn"`
    Q_PROPERTY(bool splitMode MEMBER splitMode_ NOTIFY splitModeChanged)                //    _ bool     `property:"splitMode"`
    Q_PROPERTY(bool setupGuideSeen MEMBER setupGuideSeen_ NOTIFY setupGuideSeenChanged) //    _ bool     `property:"setupGuideSeen"`
    Q_PROPERTY(float usedBytes MEMBER usedBytes_ NOTIFY usedBytesChanged)               //    _ float32  `property:"usedBytes"`
    Q_PROPERTY(float totalBytes MEMBER totalBytes_ NOTIFY totalBytesChanged)            //    _ float32  `property:"totalBytes"`
    Q_PROPERTY(QString password MEMBER password_ NOTIFY passwordChanged)                //    _ string   `property:"password"`
    Q_PROPERTY(QStringList addresses MEMBER addresses_ NOTIFY addressesChanged)         //    _ []string `property:"addresses"`
    Q_PROPERTY(QString id MEMBER id_ NOTIFY idChanged)                                  //    _ string ID

signals:
    // signals used for Qt properties
    void usernameChanged(QString const &username);
    void avatarTextChanged(QString const &avatarText);
    void loggedInChanged(bool loggedIn);
    void splitModeChanged(bool splitMode);
    void setupGuideSeenChanged(bool seen);
    void usedBytesChanged(float byteCount);
    void totalBytesChanged(float byteCount);
    void passwordChanged(QString const &);
    void addressesChanged(QStringList const &);
    void idChanged(QStringList const &id);
    void toggleSplitModeFinished();

private:
    QString id_; ///< The userID.
    QString username_; ///< The username
    QString avatarText_; ///< The avatar text (i.e. initials of the user)
    bool loggedIn_{false}; ///< Is the user logged in.
    bool splitMode_{false}; ///< Is split mode active.
    bool setupGuideSeen_{false}; ///< Has the setup guide been seen.
    float usedBytes_{0.0f}; ///< The storage used by the user.
    float totalBytes_{0.0f}; ///< The storage quota of the user.
    QString password_; ///< The IMAP password of the user.
    QStringList addresses_; ///< The email address list of the user.
};


typedef std::shared_ptr<User> SPUser;


#endif // BRIDGE_QT6_USER_H
