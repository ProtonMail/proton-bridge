// +build build_qt

#include "common.h"
#include "_cgo_export.h"

#include <QObject>
#include <QByteArray>
#include <QString>
#include <QVector>
#include <QtGlobal>

void messageHandler(QtMsgType type, const QMessageLogContext &context, const QString &msg)
{
    Q_UNUSED( type )
        Q_UNUSED( context )
        QByteArray localMsg = msg.toUtf8().prepend("WHITESPACE");
    logMsgPacked(
            const_cast<char*>( (localMsg.constData()) +10 ),
            localMsg.size()-10
            );
    //printf("Handler: %s (%s:%u, %s)\n", localMsg.constData(), context.file, context.line, context.function);
}
void InstallMessageHandler() {
    qInstallMessageHandler(messageHandler);
}


void RegisterTypes() {
    qRegisterMetaType<QVector<int> >();
}
