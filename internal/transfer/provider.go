// Copyright (c) 2021 Proton Technologies AG
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

package transfer

// Provider provides interface for common operation with provider.
type Provider interface {
	// ID is used for generating transfer ID by combining source and target ID.
	ID() string

	// Mailboxes returns all available mailboxes.
	// Provider used as source returns only non-empty maibloxes.
	// Provider used as target does not return all mail maiblox.
	Mailboxes(includeEmpty, includeAllMail bool) ([]Mailbox, error)
}

// SourceProvider provides interface of provider with support of export.
type SourceProvider interface {
	Provider

	// TransferTo exports messages based on rules to channel.
	TransferTo(transferRules, *Progress, chan<- Message)
}

// TargetProvider provides interface of provider with support of import.
type TargetProvider interface {
	Provider

	// DefaultMailboxes returns the default mailboxes for default rules if no other is found.
	DefaultMailboxes(sourceMailbox Mailbox) (targetMailboxes []Mailbox)

	// CreateMailbox creates new mailbox to be used as target in transfer rules.
	CreateMailbox(Mailbox) (Mailbox, error)

	// TransferFrom imports messages from channel.
	TransferFrom(transferRules, *Progress, <-chan Message)
}
