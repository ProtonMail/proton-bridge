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


#ifndef BRIDGE_PP_USER_H
#define BRIDGE_PP_USER_H


namespace bridgepp {


//****************************************************************************************************************************************************
/// A wrapper QObject class around a C++ enum. The purpose of this is to be able to use this enum in both Qt and QML code.
/// See https://qml.guide/enums-in-qt-qml/ for details (we used Q_OBJECT instead of Q_GADGET as in the reference document avoid a QML warning
/// complaining about the case of the data type).
//****************************************************************************************************************************************************
class EUserState : public QObject {
Q_OBJECT
public:
    enum class State {
        SignedOut = 0,
        Locked = 1,
        Connected = 2
    };


    Q_ENUM(State)


    EUserState() = delete; ///< Default constructor.
    EUserState(EUserState const &) = delete; ///< Disabled copy-constructor.
    EUserState(EUserState &&) = delete; ///< Disabled assignment copy-constructor.
    ~EUserState() = default; ///< Destructor.
    EUserState &operator=(EUserState const &) = delete; ///< Disabled assignment operator.
    EUserState &operator=(EUserState &&) = delete; ///< Disabled move assignment operator.
};


typedef EUserState::State UserState;


typedef std::shared_ptr<class User> SPUser; ///< Type definition for shared pointer to user.


//****************************************************************************************************************************************************
/// \brief User class.
//****************************************************************************************************************************************************
class User : public QObject {

Q_OBJECT
public: // data types
    enum class ENotification {
        IMAPLoginWhileSignedOut, ///< An IMAP client tried to login while the user is signed out.
        IMAPPasswordFailure, ///< An IMAP client provided an invalid password for the user.
        IMAPLoginWhileLocked, ///< An IMAP client tried to connect while the user is locked.
    };

public: // static member function
    static SPUser newUser(QObject *parent); ///< Create a new user
    static QString stateToString(UserState state); ///< Return a string describing a user state.

public: // member functions.
    User(User const &) = delete; ///< Disabled copy-constructor.
    User(User &&) = delete; ///< Disabled assignment copy-constructor.
    ~User() override = default; ///< Destructor.
    User &operator=(User const &) = delete; ///< Disabled assignment operator.
    User &operator=(User &&) = delete; ///< Disabled move assignment operator.
    void update(User const &user); ///< Update the user.
    Q_INVOKABLE QString primaryEmailOrUsername() const; ///< Return the user primary email, or, if unknown its username.
    void startNotificationCooldownPeriod(ENotification notification, qint64 durationMSecs); ///< Start the user cooldown period for a notification.
    bool isNotificationInCooldown(ENotification notification) const; ///< Return true iff the notification is in a cooldown period.

public slots:
    // slots for QML generated calls
    void toggleSplitMode(bool makeItActive);
    void logout();
    void remove();
    void configureAppleMail(QString const &address);
    void emitToggleSplitModeFinished();                 // slot for external signals

signals: // signal used to forward QML event received in the above slots
    void toggleSplitModeForUser(QString const &userID, bool makeItActive);
    void logoutUser(QString const &userID);
    void removeUser(QString const &userID);
    void configureAppleMailForUser(QString const &userID, QString const &address);

public:
    Q_PROPERTY(QString id READ id WRITE setID NOTIFY idChanged)
    Q_PROPERTY(QString username READ username WRITE setUsername NOTIFY usernameChanged)
    Q_PROPERTY(QString password READ password WRITE setPassword NOTIFY passwordChanged)
    Q_PROPERTY(QStringList addresses READ addresses WRITE setAddresses NOTIFY addressesChanged)
    Q_PROPERTY(QString avatarText READ avatarText WRITE setAvatarText NOTIFY avatarTextChanged)
    Q_PROPERTY(UserState state READ state WRITE setState NOTIFY stateChanged)
    Q_PROPERTY(bool splitMode READ splitMode WRITE setSplitMode NOTIFY splitModeChanged)
    Q_PROPERTY(float usedBytes READ usedBytes WRITE setUsedBytes NOTIFY usedBytesChanged)
    Q_PROPERTY(float totalBytes READ totalBytes WRITE setTotalBytes NOTIFY totalBytesChanged)
    Q_PROPERTY(bool isSyncing READ isSyncing WRITE setIsSyncing NOTIFY isSyncingChanged)
    Q_PROPERTY(float syncProgress READ syncProgress WRITE setSyncProgress NOTIFY syncProgressChanged)

    QString id() const;
    void setID(QString const &id);
    QString username() const;
    void setUsername(QString const &username);
    QString password() const;
    void setPassword(QString const &password);
    QStringList addresses() const;
    void setAddresses(QStringList const &addresses);
    QString avatarText() const;
    void setAvatarText(QString const &avatarText);
    UserState state() const;
    void setState(UserState state);
    bool splitMode() const;
    void setSplitMode(bool splitMode);
    float usedBytes() const;
    void setUsedBytes(float usedBytes);
    float totalBytes() const;
    void setTotalBytes(float totalBytes);
    bool isSyncing() const;
    void setIsSyncing(bool syncing);
    float syncProgress() const;
    void setSyncProgress(float progress);

signals:
    // signals used for Qt properties
    void idChanged(QString const &id);
    void usernameChanged(QString const &username);
    void passwordChanged(QString const &);
    void addressesChanged(QStringList const &);
    void avatarTextChanged(QString const &avatarText);
    void loggedInChanged(bool loggedIn);
    void stateChanged(UserState state);
    void splitModeChanged(bool splitMode);
    void usedBytesChanged(float byteCount);
    void totalBytesChanged(float byteCount);
    void toggleSplitModeFinished();
    void isSyncingChanged(bool syncing);
    void syncProgressChanged(float syncProgress);

private: // member functions.
    User(QObject *parent); ///< Default constructor.

private: // data members.
    QMap<ENotification, QDateTime> notificationCooldownList_; ///< A list of cooldown period end time for notifications.
    QString id_; ///< The userID.
    QString username_; ///< The username
    QString password_; ///< The IMAP password of the user.
    QStringList addresses_; ///< The email address list of the user.
    QString avatarText_; ///< The avatar text (i.e. initials of the user)
    UserState state_ { UserState::SignedOut }; ///< The state of the user
    bool splitMode_ { false }; ///< Is split mode active.
    float usedBytes_ { 0.0f }; ///< The storage used by the user.
    float totalBytes_ { 1.0f }; ///< The storage quota of the user.
    bool isSyncing_ { false }; ///< Is a sync in progress for the user.
    float syncProgress_ { 0.0f }; ///< The sync progress.
};


} // namespace bridgepp


#endif // BRIDGE_PP_USER_H
