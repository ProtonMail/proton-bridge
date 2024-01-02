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


#include "KnowledgeBaseTab.h"

#include "GRPCService.h"
#include "bridgepp/GRPC/EventFactory.h"


using namespace bridgepp;


//****************************************************************************************************************************************************
/// \param[in] parent The parent widget of the tab.
//****************************************************************************************************************************************************
KnowledgeBaseTab::KnowledgeBaseTab(QWidget* parent)
    : QWidget(parent) {
    ui_.setupUi(this);

    connect(ui_.checkSuggestion1, &QCheckBox::stateChanged, this, &KnowledgeBaseTab::updateGuiState);
    connect(ui_.checkSuggestion2, &QCheckBox::stateChanged, this, &KnowledgeBaseTab::updateGuiState);
    connect(ui_.checkSuggestion3, &QCheckBox::stateChanged, this, &KnowledgeBaseTab::updateGuiState);
    connect(ui_.buttonSend, &QCheckBox::clicked, this, &KnowledgeBaseTab::sendKnowledgeBaseSuggestions);

}


//****************************************************************************************************************************************************
/// \param[in] checkbox The check box.
/// \param[in] widgets The widgets to conditionally enable.
//****************************************************************************************************************************************************
void enableWidgetsIfChecked(QCheckBox const* checkbox, QWidgetList const& widgets) {
    bool const checked = checkbox->isChecked();
    for (QWidget *const widget: widgets) {
        widget->setEnabled(checked);
    }
}


//****************************************************************************************************************************************************
/// \return The suggestions.
//****************************************************************************************************************************************************
QList<KnowledgeBaseSuggestion> KnowledgeBaseTab::getSuggestions() const {
    QList<KnowledgeBaseSuggestion> result;
    if (ui_.checkSuggestion1->isChecked()) {
        result.push_back({ .url = ui_.editUrl1->text(), .title = ui_.editTitle1->text() });
    }
    if (ui_.checkSuggestion2->isChecked()) {
        result.push_back({ .url = ui_.editUrl2->text(), .title = ui_.editTitle2->text() });
    }
    if (ui_.checkSuggestion3->isChecked()) {
        result.push_back({ .url = ui_.editUrl3->text(), .title = ui_.editTitle3->text() });
    }
    return result;
}


//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void KnowledgeBaseTab::sendKnowledgeBaseSuggestions() const {
    app().grpc().sendEvent(newKnowledgeBaseSuggestionsEvent(this->getSuggestions()));
}

//****************************************************************************************************************************************************
//
//****************************************************************************************************************************************************
void KnowledgeBaseTab::updateGuiState() {
    enableWidgetsIfChecked(ui_.checkSuggestion1, { ui_.labelTitle1, ui_.editTitle1, ui_.labelUrl1, ui_.editUrl1});
    enableWidgetsIfChecked(ui_.checkSuggestion2, { ui_.labelTitle2, ui_.editTitle2, ui_.labelUrl2, ui_.editUrl2});
    enableWidgetsIfChecked(ui_.checkSuggestion3, { ui_.labelTitle3, ui_.editTitle3, ui_.labelUrl3, ui_.editUrl3});
}


//****************************************************************************************************************************************************
/// \param[in] userInput The user input.
//****************************************************************************************************************************************************
void KnowledgeBaseTab::requestKnowledgeBaseSuggestions(QString const& userInput) const {
    ui_.editUserInput->setPlainText(userInput);
    ui_.labelLastReceived->setText(tr("Last received: %1").arg(QDateTime::currentDateTime().toString("yyyy-MM-dd HH:mm:ss.zzz")));
}
