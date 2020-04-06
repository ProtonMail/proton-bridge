// Copyright (c) 2020 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package pmapi

// ConversationsCount have same structure as MessagesCount.
type ConversationsCount MessagesCount

// ConversationsCountsRes holds response from server.
type ConversationsCountsRes struct {
	Res

	Counts []*ConversationsCount
}

// Conversation contains one body and multiple metadata.
type Conversation struct{}

// CountConversations counts conversations by label.
func (c *client) CountConversations(addressID string) (counts []*ConversationsCount, err error) {
	reqURL := "/conversations/count"
	if addressID != "" {
		reqURL += ("?AddressID=" + addressID)
	}
	req, err := c.NewRequest("GET", reqURL, nil)
	if err != nil {
		return
	}

	var res ConversationsCountsRes
	if err = c.DoJSON(req, &res); err != nil {
		return
	}

	counts, err = res.Counts, res.Err()
	return
}
