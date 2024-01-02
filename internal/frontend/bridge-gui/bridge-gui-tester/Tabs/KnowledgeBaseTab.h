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


#ifndef BRIDGE_GUI_TESTER_KNOWLEDGE_BASE_TAB_H
#define BRIDGE_GUI_TESTER_KNOWLEDGE_BASE_TAB_H


#include "ui_KnowledgeBaseTab.h"
#include <bridgepp/GRPC/GRPCClient.h>


//****************************************************************************************************************************************************
/// \brief Knowledge base table.
//****************************************************************************************************************************************************
class KnowledgeBaseTab: public QWidget {
public: // member functions.
    explicit KnowledgeBaseTab(QWidget *parent = nullptr); ///< Default constructor.
    KnowledgeBaseTab(KnowledgeBaseTab const&) = delete; ///< Disabled copy-constructor.
    KnowledgeBaseTab(KnowledgeBaseTab&&) = delete; ///< Disabled assignment copy-constructor.
    ~KnowledgeBaseTab() override = default; ///< Destructor.
    KnowledgeBaseTab& operator=(KnowledgeBaseTab const&) = delete; ///< Disabled assignment operator.
    KnowledgeBaseTab& operator=(KnowledgeBaseTab&&) = delete; ///< Disabled move assignment operator.
    QList<bridgepp::KnowledgeBaseSuggestion> getSuggestions() const; ///< Returns the suggestions.

public slots:
    void requestKnowledgeBaseSuggestions(QString const &userInput) const; ///< Slot for the 'RequestKnowledgeBaseSuggestions' gRPC call.

private slots:
    void sendKnowledgeBaseSuggestions() const; ///< Send a KnowledgeBaseSuggestions event.
    void updateGuiState(); ///< Update the GUI state.

private: // data members
    Ui::KnowledgeBaseTab ui_ {}; ///< The UI for the widget.
};



#endif //BRIDGE_GUI_TESTER_KNOWLEDGE_BASE_TAB_H
