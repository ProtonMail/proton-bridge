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


#include "BridgeUtils.h"
#include "Exception/Exception.h"
#include <random>


namespace bridgepp
{


namespace {


QString const configFolder = "protonmail/bridge";
QMutex rngMutex; ///< the mutex to use when accessing the rng.

QStringList const firstNames {
    "James", "John", "Robert", "Michael", "William", "David", "Richard", "Charles", "Joseph", "Thomas", "Christopher", "Daniel", "Paul", "Mark",
    "Donald", "George", "Kenneth", "Steven", "Edward", "Brian", "Ronald", "Anthony", "Kevin", "Jason", "Matthew", "Gary", "Timothy", "Jose", "Larry",
    "Jeffrey", "Frank", "Scott", "Eric", "Stephen", "Andrew", "Raymond", "Jack", "Gregory", "Joshua", "Jerry", "Dennis", "Walter", "Patrick", "Peter",
    "Harold", "Douglas", "Henry", "Carl", "Arthur", "Ryan", "Mary", "Patricia", "Barbara", "Linda", "Elizabeth", "Maria", "Jennifer", "Susan",
    "Margaret", "Dorothy", "Lisa", "Nancy", "Karen", "Betty", "Helen", "Sandra", "Donna", "Ruth", "Sharon", "Michelle", "Laura", "Sarah", "Kimberly",
    "Deborah", "Jessica", "Shirley", "Cynthia", "Angela", "Emily", "Brenda", "Amy", "Anna", "Rebecca", "Virginia", "Kathleen", "Pamela", "Martha",
    "Debra", "Amanda", "Stephanie", "Caroline", "Christine", "Marie", "Janet", "Catherine", "Frances", "Ann", "Joyce", "Diane", "Alice",
}; ///< List of common US first names. (source: census.gov)


QStringList const lastNames {
    "Smith", "Johnson", "Williams", "Brown", "Jones", "Garcia", "Miller", "Davis", "Rodriguez", "Martinez", "Hernandez", "Lopez", "Gonzalez",
    "Wilson", "Anderson", "Thomas", "Taylor", "Moore", "Jackson", "Martin", "Lee", "Perez", "Thompson", "White", "Harris", "Sanchez", "Clark",
    "Ramirez", "Lewis", "Robinson", "Walker", "Young", "Allen", "King", "Wright", "Scott", "Torres", "Nguyen", "Hill", "Flores", "Green", "Adams",
    "Nelson", "Baker", "Hall", "Rivera", "Campbell", "Mitchell", "Carter", "Roberts", "Gomez", "Phillips", "Evans", "Turner", "Diaz", "Parker",
    "Cruz", "Edwards", "Collins", "Reyes", "Stewart", "Morris", "Morales", "Murphy", "Cook", "Rogers", "Gutierrez", "Ortiz", "Morgan", "Cooper",
    "Peterson", "Bailey", "Reed", "Kelly", "Howard", "Ramos", "Kim", "Cox", "Ward", "Richardson", "Watson", "Brooks", "Chavez", "Wood", "James",
    "Bennett", "Gray", "Mendoza", "Ruiz", "Hughes", "Price", "Alvarez", "Castillo", "Sanders", "Patel", "Myers", "Long", "Ross", "Foster", "Jimenez"
}; ///< List of common US last names (source: census.gov)


//****************************************************************************************************************************************************
/// \brief Return the 64 bit Mersenne twister random number generator.
//****************************************************************************************************************************************************
std::mt19937_64& rng()
{
    // Do not use for crypto. std::random_device is not good enough.
    static std::mt19937_64 generator = std::mt19937_64(std::random_device()());
    return generator;
}


} // anonymous namespace


//****************************************************************************************************************************************************
/// \return user configuration directory used by bridge (based on Golang OS/File's UserConfigDir).
//****************************************************************************************************************************************************
QString userConfigDir()
{
    QString dir;
#ifdef Q_OS_WIN
    dir = qgetenv ("AppData");
    if (dir.isEmpty())
        throw Exception("%AppData% is not defined.");
#elif defined(Q_OS_IOS) || defined(Q_OS_DARWIN)
    dir = qgetenv ("HOME");
    if (dir.isEmpty())
        throw Exception("$HOME is not defined.");
    dir += "/Library/Application Support";
#else
    dir = qgetenv ("XDG_CONFIG_HOME");
        if (dir.isEmpty())
            dir = qgetenv ("HOME");
        if (dir.isEmpty())
            throw Exception("neither $XDG_CONFIG_HOME nor $HOME are defined");
        dir += "/.config";
#endif
    QString const folder = QDir(dir).absoluteFilePath(configFolder);
    QDir().mkpath(folder);

    return folder;
}


//****************************************************************************************************************************************************
/// \return user configuration directory used by bridge (based on Golang OS/File's UserCacheDir).
//****************************************************************************************************************************************************
QString userCacheDir()
{
    QString dir;

#ifdef Q_OS_WIN
    dir = qgetenv ("LocalAppData");
    if (dir.isEmpty())
        throw Exception("%LocalAppData% is not defined.");
#elif defined(Q_OS_IOS) || defined(Q_OS_DARWIN)
    dir = qgetenv ("HOME");
    if (dir.isEmpty())
        throw Exception("$HOME is not defined.");
    dir += "/Library/Caches";
#else
    dir = qgetenv ("XDG_CACHE_HOME");
        if (dir.isEmpty())
            dir = qgetenv ("HOME");
        if (dir.isEmpty())
            throw Exception("neither XDG_CACHE_HOME nor $HOME are defined");
        dir += "/.cache";
#endif

    QString const folder = QDir(dir).absoluteFilePath(configFolder);
    QDir().mkpath(folder);

    return folder;
}


//****************************************************************************************************************************************************
/// \return The value GOOS would return for the current platform.
//****************************************************************************************************************************************************
QString goos()
{
#if defined(Q_OS_DARWIN)
    return "darwin";
#elif defined(Q_OS_WINDOWS)
    return "windows";
#else
    return "linux";
#endif
}


//****************************************************************************************************************************************************
/// Slow, but not biased. Not for use in crypto functions though, as the RNG use std::random_device as a seed.
///
/// \return a random number in the range [0, n-1]
//****************************************************************************************************************************************************
qint64 randN(qint64 n)
{
    QMutexLocker locker(&rngMutex);
    return (n > 0) ? std::uniform_int_distribution<qint64>(0, n - 1)(rng()) : 0;
}


//****************************************************************************************************************************************************
/// \return A random first name.
//****************************************************************************************************************************************************
QString randomFirstName()
{
    return firstNames[randN(firstNames.size())];
}


//****************************************************************************************************************************************************
/// \return A random last name.
//****************************************************************************************************************************************************
QString randomLastName()
{
    return lastNames[randN(lastNames.size())];
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
SPUser randomUser()
{
    SPUser user = User::newUser(nullptr);
    user->setID(QUuid::createUuid().toString());
    QString const firstName = randomFirstName();
    QString const lastName = randomLastName();
    QString const username = QString("%1.%2").arg(firstName.toLower(), lastName.toLower());
    user->setUsername(username);
    user->setAddresses(QStringList() << (username + "@proton.me") << (username + "@protonmail.com") );
    user->setPassword(QUuid().createUuid().toString(QUuid::StringFormat::WithoutBraces).left(20));
    user->setAvatarText(firstName.left(1) + lastName.left(1));
    user->setLoggedIn(true);
    user->setSplitMode(false);
    user->setSetupGuideSeen(true);
    qint64 const totalBytes = (500 + randN(2501)) * 1000000;
    user->setUsedBytes(float(bridgepp::randN(totalBytes + 1)) * 1.05f); // we maybe slightly over quota
    user->setTotalBytes(float(totalBytes));
    return user;
}


} // namespace bridgepp
