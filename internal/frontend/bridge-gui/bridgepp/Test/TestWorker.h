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


#ifndef BRIDGE_GUI_TEST_WORKER_H
#define BRIDGE_GUI_TEST_WORKER_H


#include <bridgepp/Worker/Overseer.h>
#include <gtest/gtest.h>


//****************************************************************************************************************************************************
/// \brief Test worker class.
///
/// This worker simply waits:
/// - For a specified amount of time and will succeed (emit finished()) or fail (emit error()) based on its parameters.
/// - to be cancelled (and will emit cancelled in that case).
//****************************************************************************************************************************************************
class TestWorker : public bridgepp::Worker {
Q_OBJECT
public: // member functions.
    TestWorker(); ///< Default constructor.
    TestWorker(TestWorker const &) = delete; ///< Disabled copy-constructor.
    TestWorker(TestWorker &&) = delete; ///< Disabled assignment copy-constructor.
    ~TestWorker() override = default; ///< Destructor.
    TestWorker &operator=(TestWorker const &) = delete; ///< Disabled assignment operator.
    TestWorker &operator=(TestWorker &&) = delete; ///< Disabled move assignment operator.
    void setLifetime(qint64 lifetimeMs); ///< Set the lifetime of the worker.
    void setWillSucceed(bool willSucceed); ///< Set if the worker will succeed.
    void run() override; ///< Run the worker.
    void cancel(); ///< Cancel the worker.

private: // data members
    qint64 lifetimeMs_ { 10 }; ///< The lifetime of the worker in milliseconds.
    bool willSucceed_ { true }; ///< Will the worker succeed?
    QAtomicInteger<char> cancelled_; ///< Has the worker been cancelled.
};


//****************************************************************************************************************************************************
/// \brief Fixture class for worker tests.
//****************************************************************************************************************************************************
class Workers : public testing::Test {
public: // member functions.
    Workers(); ///< Default constructor.
    Workers(Workers const &) = delete; ///< Disabled copy-constructor.
    Workers(Workers &&) = delete; ///< Disabled assignment copy-constructor.
    ~Workers() = default; ///< Destructor.
    Workers &operator=(Workers const &) = delete; ///< Disabled assignment operator.
    Workers &operator=(Workers &&) = delete; ///< Disabled move assignment operator.

protected: // member functions.
    void SetUp() override; ///< Setup the fixture.
    void TearDown() override; ///< Tear down the fixture.

protected: // data type
    struct Results {
        bool started { false };
        bool finished { false };
        bool error { false };
        bool cancelled { false };
    }; ///< Test results data type

protected: // data members
    QCoreApplication app_; ///< The Qt application required for event loop.
    bridgepp::UPOverseer overseer_; ///< The overseer for the worker.
    TestWorker *worker_ { nullptr }; ///< The worker.
    QEventLoop loop_; ///< The event loop.
    Results results_; ///< The test results.
};


#endif //BRIDGE_GUI_TEST_WORKER_H
