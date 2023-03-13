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


#include "Overseer.h"
#include "../Exception/Exception.h"


namespace bridgepp {


//****************************************************************************************************************************************************
/// \param[in] worker The worker.
/// \param[in] parent The parent object of the worker.
//****************************************************************************************************************************************************
Overseer::Overseer(Worker *worker, QObject *parent)
    : QObject(parent)
    , thread_(new QThread(parent))
    , worker_(worker) {
    if (!worker_) {
        throw Exception("Overseer cannot accept a nil worker.");
    }
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
Overseer::~Overseer() {
    this->releaseWorker();
}


//****************************************************************************************************************************************************
/// \param[in] autorelease Should the overseer automatically release the worker and thread when done.
//****************************************************************************************************************************************************
void Overseer::startWorker(bool autorelease) const {
    if (!worker_) {
        throw Exception("Cannot start overseer with null worker.");
    }
    if (!thread_) {
        throw Exception("Cannot start overseer with null thread.");
    }

    worker_->moveToThread(thread_);
    connect(thread_, &QThread::started, worker_, &Worker::run);
    connect(worker_, &Worker::finished, [&]() { thread_->quit(); }); // Safety, normally the thread already properly quits.
    connect(worker_, &Worker::error, [&]() { thread_->quit(); });
    connect(worker_, &Worker::cancelled, [&]() { thread_->quit(); });

    if (autorelease) {
        connect(worker_, &Worker::error, this, &Overseer::releaseWorker);
        connect(worker_, &Worker::cancelled, this, &Overseer::releaseWorker);
        connect(worker_, &Worker::finished, this, &Overseer::releaseWorker);
    }

    thread_->start();
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void Overseer::releaseWorker() {
    if (worker_) {
        worker_->deleteLater();
        worker_ = nullptr;
    }

    if (thread_) {
        if (!thread_->isFinished()) {
            thread_->quit();
            if (!thread_->wait(maxTerminationWaitTimeMs))
                thread_->terminate();
        }
        thread_->deleteLater();
        thread_ = nullptr;
    }
}


//****************************************************************************************************************************************************
/// \return true iff the worker is finished.
//****************************************************************************************************************************************************
bool Overseer::isFinished() const {
    if ((!worker_) || (!worker_->thread())) {
        return true;
    }

    return worker_->thread()->isFinished();
}


//****************************************************************************************************************************************************
/// \param timeoutMs The timeout after which the function should return false if the event stream reader is not finished. if -1 one, the function
/// never times out.
/// \return false if and only if the timeout delay was reached.
//****************************************************************************************************************************************************
bool Overseer::wait(qint32 timeoutMs) const {
    if (this->isFinished()) {
        return true;
    }

    QEventLoop loop;
    QTimer timer;
    bool inTime = true;
    if (timeoutMs >= 0) {
        connect(&timer, &QTimer::timeout, &loop, [&]() {
            loop.quit();
            inTime = false;
        });
        timer.setSingleShot(true);
        timer.start(timeoutMs);
    }

    QTimer repeatTimer; // safety timer, used if for some reason the worker does not emit finished(), error() or cancelled()
    repeatTimer.setSingleShot(false);
    repeatTimer.setInterval(100);
    connect(&repeatTimer, &QTimer::timeout, [&]() { if (this->isFinished()) { loop.quit(); }});
    repeatTimer.start();

    connect(worker_, &Worker::finished, &loop, &QEventLoop::quit);
    connect(worker_, &Worker::error, &loop, &QEventLoop::quit);
    connect(worker_, &Worker::cancelled, &loop, &QEventLoop::quit);

    loop.exec();

    return inTime;
}


//****************************************************************************************************************************************************
/// \return The worker.
//****************************************************************************************************************************************************
Worker *Overseer::worker() const {
    return worker_;
}


} // namespace bridgepp
