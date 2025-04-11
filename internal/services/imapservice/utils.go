// Copyright (c) 2025 Proton AG
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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imapservice

import (
	"errors"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
)

type connectorInterface interface {
	getSenderProtonAddress(p *parser.Parser) (proton.Address, error)
	getAddress(id string) (proton.Address, bool)
	getPrimaryAddress() (proton.Address, error)
	getAddressMode() usertypes.AddressMode
	logError(err error, errMsg string)
}

func (s *Connector) logError(err error, errMsg string) {
	s.log.WithError(err).Warn(errMsg)
}

func (s *Connector) getAddressMode() usertypes.AddressMode {
	return s.addressMode
}

func (s *Connector) getPrimaryAddress() (proton.Address, error) {
	return s.identityState.GetPrimaryAddress()
}

func (s *Connector) getAddress(id string) (proton.Address, bool) {
	return s.identityState.GetAddress(id)
}

func getImportAddress(p *parser.Parser, isDraft bool, id string, conn connectorInterface) (proton.Address, error) {
	// addr is primary for combined mode or active for split mode
	address, ok := conn.getAddress(id)
	if !ok {
		return proton.Address{}, errors.New("could not find account address")
	}

	// If the address is external and not BYOE - with sending enabled, then use the primary address as an import target.
	if address.Type == proton.AddressTypeExternal && !address.Send {
		var err error
		address, err = conn.getPrimaryAddress()
		if err != nil {
			return proton.Address{}, errors.New("could not get primary account address")
		}
	}

	inCombinedMode := conn.getAddressMode() == usertypes.AddressModeCombined
	if !inCombinedMode {
		return address, nil
	}

	senderAddr, err := conn.getSenderProtonAddress(p)
	if err != nil {
		if !errors.Is(err, errNoSenderAddressMatch) {
			conn.logError(err, "Could not get import address")
		}

		// We did not find a match, so we use the default address.
		return address, nil
	}

	if senderAddr.ID == address.ID {
		return address, nil
	}

	// GODT-3185 / BRIDGE-120 In combined mode, in certain cases we adapt the address used for encryption.
	// - draft with non-default address in combined mode: using sender address
	// - import with non-default address in combined mode: using sender address
	// - import with non-default disabled address in combined mode: using sender address
	isSenderAddressDisabled := (!bool(senderAddr.Send)) || (senderAddr.Status != proton.AddressStatusEnabled)
	isSenderExternalNonBYOE := senderAddr.Type == proton.AddressTypeExternal && !bool(senderAddr.Send)

	// Forbid drafts/imports for external non-BYOE addresses
	if isSenderExternalNonBYOE || (isDraft && isSenderAddressDisabled) {
		return address, nil
	}

	return senderAddr, nil
}
