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


#ifndef BRIDGE_GUI_USER_H
#define BRIDGE_GUI_USER_H


namespace bridgepp
{


typedef std::shared_ptr<class User> SPUser; ///< Type definition for shared pointer to user.


//****************************************************************************************************************************************************
/// \brief User class.
//****************************************************************************************************************************************************
class User : public QObject
{
Q_OBJECT
public: // static member function
    static SPUser newUser(QObject *parent); ///< Create a new user

public: // member functions.
    User(User const &) = delete; ///< Disabled copy-constructor.
    User(User &&) = delete; ///< Disabled assignment copy-constructor.
    ~User() override = default; ///< Destructor.
    User &operator=(User const &) = delete; ///< Disabled assignment operator.
    User &operator=(User &&) = delete; ///< Disabled move assignment operator.
    void update(User const &user); ///< Update the user.

public slots:
    // slots for QML generated calls
    void toggleSplitMode(bool makeItActive);            //    _ func(makeItActive bool) `slot:"toggleSplitMode"`
    void logout();                                      //    _ func()                  `slot:"logout"`
    void remove();                                      //    _ func()                  `slot:"remove"`
    void configureAppleMail(QString const &address);    //    _ func(address string)    `slot:"configureAppleMail"`
    void emitToggleSplitModeFinished();                 // slot for external signals

signals: // signal used to forward QML event received in the above slots
    void toggleSplitModeForUser(QString const &userID, bool makeItActive);
    void logoutUser(QString const &userID);
    void removeUser(QString const &userID);
    void configureAppleMailForUser(QString const &userID, QString const &address);


public:
    Q_PROPERTY(QString id READ id WRITE setID NOTIFY idChanged)                                                 //    _ string ID
    Q_PROPERTY(QString username READ username WRITE setUsername NOTIFY usernameChanged)                         //    _ string   `property:"username"`
    Q_PROPERTY(QString password READ password WRITE setPassword NOTIFY passwordChanged)                         //    _ string   `property:"password"`
    Q_PROPERTY(QStringList addresses READ addresses WRITE setAddresses NOTIFY addressesChanged)                 //    _ []string `property:"addresses"`
    Q_PROPERTY(QString avatarText READ avatarText WRITE setAvatarText NOTIFY avatarTextChanged)                 //    _ string   `property:"avatarText"`
    Q_PROPERTY(bool loggedIn READ loggedIn WRITE setLoggedIn NOTIFY loggedInChanged)                            //    _ bool     `property:"loggedIn"`
    Q_PROPERTY(bool splitMode READ splitMode WRITE setSplitMode NOTIFY splitModeChanged)                        //    _ bool     `property:"splitMode"`
    Q_PROPERTY(bool setupGuideSeen READ setupGuideSeen WRITE setSetupGuideSeen NOTIFY setupGuideSeenChanged)    //    _ bool     `property:"setupGuideSeen"`
    Q_PROPERTY(float usedBytes READ usedBytes WRITE setUsedBytes NOTIFY usedBytesChanged)                       //    _ float32  `property:"usedBytes"`
    Q_PROPERTY(float totalBytes READ totalBytes WRITE setTotalBytes NOTIFY totalBytesChanged)                   //    _ float32  `property:"totalBytes"`

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
    bool loggedIn() const;
    void setLoggedIn(bool loggedIn);
    bool splitMode() const;
    void setSplitMode(bool splitMode);
    bool setupGuideSeen() const;
    void setSetupGuideSeen(bool setupGuideSeen);
    float usedBytes() const;
    void setUsedBytes(float usedBytes);
    float totalBytes() const;
    void setTotalBytes(float totalBytes);

signals:
    // signals used for Qt properties
    void idChanged(QString const &id);
    void usernameChanged(QString const &username);
    void passwordChanged(QString const &);
    void addressesChanged(QStringList const &);
    void avatarTextChanged(QString const &avatarText);
    void loggedInChanged(bool loggedIn);
    void splitModeChanged(bool splitMode);
    void setupGuideSeenChanged(bool seen);
    void usedBytesChanged(float byteCount);
    void totalBytesChanged(float byteCount);

    void toggleSplitModeFinished();

private: // member functions.
    User(QObject *parent); ///< Default constructor.

private: // data members.
    QString id_; ///< The userID.
    QString username_; ///< The username
    QString password_; ///< The IMAP password of the user.
    QStringList addresses_; ///< The email address list of the user.
    QString avatarText_; ///< The avatar text (i.e. initials of the user)
    bool loggedIn_ { true }; ///< Is the user logged in.
    bool splitMode_ { false }; ///< Is split mode active.
    bool setupGuideSeen_ { false }; ///< Has the setup guide been seen.
    float usedBytes_ { 0.0f }; ///< The storage used by the user.
    float totalBytes_ { 1.0f }; ///< The storage quota of the user.
};


} // namespace bridgepp


#endif // BRIDGE_GUI_USER_H
