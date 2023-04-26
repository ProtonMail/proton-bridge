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

// clazy:excludeall=lambda-in-connect

#include "TestWorker.h"
#include <bridgepp/Worker/Overseer.h>
#include <bridgepp/Exception/Exception.h>


using namespace bridgepp;


namespace {


qint32 dummyArgc = 1; ///< A dummy int value because QCoreApplication constructor requires a reference to it.


}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Workers::Workers()
    : testing::Test()
    , app_(dummyArgc, nullptr) {
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void Workers::SetUp() {
    Test::SetUp();

    EXPECT_NO_THROW(worker_ = new TestWorker);

    QObject::connect(worker_, &TestWorker::started, [&]() { results_.started = true; });
    QObject::connect(worker_, &TestWorker::finished, [&]() { results_.finished = true; });
    QObject::connect(worker_, &TestWorker::finished, &loop_, &QEventLoop::quit);
    QObject::connect(worker_, &TestWorker::error, [&] { results_.error = true; });
    QObject::connect(worker_, &TestWorker::error, &loop_, &QEventLoop::quit);
    QObject::connect(worker_, &TestWorker::error, [&] { results_.error = true; });
    QObject::connect(worker_, &TestWorker::error, &loop_, &QEventLoop::quit);
    QObject::connect(worker_, &TestWorker::cancelled, [&] { results_.cancelled = true; });
    QObject::connect(worker_, &TestWorker::cancelled, &loop_, &QEventLoop::quit);

    overseer_ = std::make_unique<Overseer>(worker_, nullptr);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void Workers::TearDown() {
    EXPECT_NO_FATAL_FAILURE(overseer_.reset());
    Test::TearDown();
}


//****************************************************************************************************************************************************
/// \param[in] lifetimeMs The lifetime of the worker in milliseconds.
/// \param[in] willSucceed Will the worker succeed (emit finished) or fail (emit error).
//****************************************************************************************************************************************************
TestWorker::TestWorker()
    : Worker(nullptr) {
}


//****************************************************************************************************************************************************
/// \param[in] lifetimeMs The lifetime of the worker in milliseconds.
//****************************************************************************************************************************************************
void TestWorker::setLifetime(qint64 lifetimeMs) {
    lifetimeMs_ = lifetimeMs;
}


//****************************************************************************************************************************************************
/// \param[in] willSucceed Will the worker succeed?
//****************************************************************************************************************************************************
void TestWorker::setWillSucceed(bool willSucceed) {
    willSucceed_ = willSucceed;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TestWorker::run() {
    emit started();

    QElapsedTimer timer;
    timer.start();
    while (true) {
        if (cancelled_.loadRelaxed()) {
            emit cancelled();
            return;
        }
        if (timer.elapsed() >= lifetimeMs_) {
            break;
        }
    }

    if (willSucceed_) {
        emit finished();
    } else {
        emit error(QString());
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void TestWorker::cancel() {
    cancelled_.storeRelaxed(1);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(Workers, SuccessfulWorker) {
    worker_->setLifetime(10);
    worker_->setWillSucceed(true);

    EXPECT_NO_THROW(overseer_->startWorker(false));
    EXPECT_NO_THROW(loop_.exec());

    EXPECT_TRUE(results_.started);
    EXPECT_TRUE(results_.finished);
    EXPECT_FALSE(results_.error);
    EXPECT_FALSE(results_.cancelled);

    EXPECT_TRUE(overseer_->worker() != nullptr); // overseer started without autorelease.
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(Workers, ErrorWorker) {
    worker_->setLifetime(10);
    worker_->setWillSucceed(false);

    EXPECT_NO_THROW(overseer_->startWorker(true));
    EXPECT_NO_THROW(loop_.exec());

    EXPECT_TRUE(results_.started);
    EXPECT_FALSE(results_.finished);
    EXPECT_TRUE(results_.error);
    EXPECT_FALSE(results_.cancelled);

    EXPECT_TRUE(overseer_->worker() == nullptr); // overseer started with autorelease.
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(Workers, CancelledWorker) {
    worker_->setLifetime(10000);
    worker_->setWillSucceed(true);
    EXPECT_NO_THROW(overseer_->startWorker(false));
    EXPECT_NO_THROW(QTimer::singleShot(10, [&]() { worker_->cancel(); }));

    EXPECT_NO_THROW(loop_.exec());

    EXPECT_TRUE(results_.started);
    EXPECT_FALSE(results_.finished);
    EXPECT_FALSE(results_.error);
    EXPECT_TRUE(results_.cancelled);
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
TEST_F(Workers, Wait) {
    worker_->setLifetime(10000);
    worker_->setWillSucceed(true);
    overseer_->startWorker(true);

    bool isFinished = false;
    EXPECT_NO_THROW(isFinished = overseer_->isFinished());
    EXPECT_FALSE(isFinished);

    EXPECT_NO_THROW(isFinished = overseer_->wait(10));
    EXPECT_FALSE(isFinished);

    worker_->cancel();

    EXPECT_NO_THROW(isFinished = overseer_->wait(10000));
    EXPECT_TRUE(isFinished);

    EXPECT_NO_THROW(isFinished = overseer_->isFinished());
    EXPECT_TRUE(isFinished);
}


