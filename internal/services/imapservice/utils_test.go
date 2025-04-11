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
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message/parser"
	"github.com/stretchr/testify/require"
)

type testConnector struct {
	addressMode        usertypes.AddressMode
	primaryAddress     proton.Address
	senderAddress      proton.Address
	imapAddress        proton.Address
	senderAddressError error
}

func (t *testConnector) getSenderProtonAddress(_ *parser.Parser) (proton.Address, error) {
	return t.senderAddress, t.senderAddressError
}

func (t *testConnector) getAddress(_ string) (proton.Address, bool) {
	return t.imapAddress, true
}

func (t *testConnector) getPrimaryAddress() (proton.Address, error) {
	return t.primaryAddress, nil
}

func (t *testConnector) getAddressMode() usertypes.AddressMode {
	return t.addressMode
}

func (t *testConnector) logError(_ error, _ string) {
}

func Test_GetImportAddress_SplitMode(t *testing.T) {
	primaryAddress := proton.Address{
		ID:      "1",
		Email:   "primary@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
		Status:  proton.AddressStatusEnabled,
	}

	imapAddressProton := proton.Address{
		ID:      "2",
		Email:   "imap@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
	}

	testConn := &testConnector{
		addressMode:    usertypes.AddressModeSplit,
		primaryAddress: primaryAddress,
		imapAddress:    imapAddressProton,
	}

	// Import address is internal, we're creating a draft.
	// Expected: returned address is internal.
	addr, err := getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)

	// Import address is internal, we're attempting to import a message.
	// Expected: returned address is internal.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)

	imapAddressBYOE := proton.Address{
		ID:      "3",
		Email:   "byoe@external.com",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeExternal,
	}

	// IMAP address is BYOE, we're creating a draft
	// Expected: returned address is BYOE.
	testConn.imapAddress = imapAddressBYOE
	addr, err = getImportAddress(nil, true, imapAddressBYOE.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressBYOE.ID, addr.ID)
	require.Equal(t, imapAddressBYOE.Email, addr.Email)
	// IMAP address is BYOE, we're importing a message
	// Expected: returned address is BYOE.
	addr, err = getImportAddress(nil, false, imapAddressBYOE.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressBYOE.ID, addr.ID)
	require.Equal(t, imapAddressBYOE.Email, addr.Email)

	imapAddressExternal := proton.Address{
		ID:      "4",
		Email:   "external@external.com",
		Send:    false,
		Receive: false,
		Type:    proton.AddressTypeExternal,
	}

	// IMAP address is external, we're creating a draft.
	// Expected: returned address is primary.
	testConn.imapAddress = imapAddressExternal
	addr, err = getImportAddress(nil, true, imapAddressExternal.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, primaryAddress.ID, addr.ID)
	require.Equal(t, primaryAddress.Email, addr.Email)
	// IMAP address is external, we're trying to import.
	// Expected: returned address is primary.
	addr, err = getImportAddress(nil, false, imapAddressExternal.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, primaryAddress.ID, addr.ID)
	require.Equal(t, primaryAddress.Email, addr.Email)
}

func Test_GetImportAddress_CombinedMode_ProtonAddresses(t *testing.T) {
	primaryAddress := proton.Address{
		ID:      "1",
		Email:   "primary@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
		Status:  proton.AddressStatusEnabled,
	}

	imapAddressProton := proton.Address{
		ID:      "2",
		Email:   "imap@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
	}

	senderAddress := proton.Address{
		ID:      "3",
		Email:   "sender@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
		Status:  proton.AddressStatusEnabled,
	}

	testConn := &testConnector{
		addressMode:    usertypes.AddressModeCombined,
		primaryAddress: primaryAddress,
		imapAddress:    imapAddressProton,
		senderAddress:  senderAddress,
	}

	// Both the sender address and the imap address are the same. We're creating a draft.
	// Expected: IMAP address is returned.
	testConn.senderAddress = imapAddressProton
	testConn.imapAddress = imapAddressProton
	addr, err := getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)
	// Both the sender address and the imap address are the same. We're trying to import
	// Expected: IMAP address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)

	// Sender address and imap address are different. Sender address is enabled and has sending enabled.
	// We're creating a draft.
	// Expected: Sender address is returned.
	testConn.senderAddress = senderAddress
	testConn.imapAddress = imapAddressProton
	addr, err = getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, senderAddress.ID, addr.ID)
	require.Equal(t, senderAddress.Email, addr.Email)
	// Sender address and imap address are different. Sender address is enabled and has sending enabled.
	// We're importing a message.
	// Expected: Sender address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, senderAddress.ID, addr.ID)
	require.Equal(t, senderAddress.Email, addr.Email)

	// Sender address and imap address are different. Sender address is disabled, but has sending enabled.
	// We're creating a draft message.
	// Expected: IMAP address is returned.
	senderAddress.Status = proton.AddressStatusDisabled
	testConn.senderAddress = senderAddress
	addr, err = getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)
	// Sender address and imap address are different. Sender address is disabled, but has sending enabled.
	// We're importing a message.
	// Expected: IMAP address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, senderAddress.ID, addr.ID)
	require.Equal(t, senderAddress.Email, addr.Email)

	// Sender address and imap address are different. Sender address is enabled, but has sending disabled.
	// We're creating a draft.
	// Expected: IMAP address is returned.
	senderAddress.Status = proton.AddressStatusEnabled
	senderAddress.Send = false
	testConn.senderAddress = senderAddress
	addr, err = getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)
	// Sender address and imap address are different. Sender address is enabled, but has sending disabled.
	// We're importing a message.
	// Expected: IMAP address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, senderAddress.ID, addr.ID)
	require.Equal(t, senderAddress.Email, addr.Email)

	// Sender address and imap address are different. But sender address is not an associated proton address.
	// We're creating a draft.
	// Expected: Sender address is returned.
	testConn.senderAddressError = errors.New("sender address is not associated with the account")
	addr, err = getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)
	// Sender address and imap address are different. But sender address is not an associated proton address.
	// We're importing a message.
	// Expected: Sender address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)
}

func Test_GetImportAddress_CombinedMode_ExternalAddresses(t *testing.T) {
	primaryAddress := proton.Address{
		ID:      "1",
		Email:   "primary@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
		Status:  proton.AddressStatusEnabled,
	}

	imapAddressProton := proton.Address{
		ID:      "2",
		Email:   "imap@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
	}

	senderAddressExternal := proton.Address{
		ID:      "3",
		Email:   "sender@external.me",
		Send:    false,
		Receive: false,
		Type:    proton.AddressTypeExternal,
		Status:  proton.AddressStatusEnabled,
	}

	senderAddressExternalSecondary := proton.Address{
		ID:      "4",
		Email:   "sender2@external.me",
		Send:    false,
		Receive: false,
		Type:    proton.AddressTypeExternal,
		Status:  proton.AddressStatusEnabled,
	}

	testConn := &testConnector{
		addressMode:    usertypes.AddressModeCombined,
		primaryAddress: primaryAddress,
		imapAddress:    imapAddressProton,
		senderAddress:  senderAddressExternal,
	}

	// Sender address is external, and we're creating a draft.
	// Expected: IMAP address is returned.
	testConn.senderAddress = senderAddressExternal
	testConn.imapAddress = imapAddressProton
	addr, err := getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)
	// Sender address is external, and we're importing a message.
	// Expected: IMAP address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, imapAddressProton.ID, addr.ID)
	require.Equal(t, imapAddressProton.Email, addr.Email)

	// Sender and IMAP address are external, and we're trying to import.
	// Expected: Primary address is returned.
	testConn.imapAddress = senderAddressExternalSecondary
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, primaryAddress.ID, addr.ID)
	require.Equal(t, primaryAddress.Email, addr.Email)
	// Sender and IMAP address are external, and we're trying to create a draft.
	// Expected: Primary address is returned.
	addr, err = getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, primaryAddress.ID, addr.ID)
	require.Equal(t, primaryAddress.Email, addr.Email)
}

func Test_GetImportAddress_CombinedMode_BYOEAddresses(t *testing.T) {
	primaryAddress := proton.Address{
		ID:      "1",
		Email:   "primary@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
		Status:  proton.AddressStatusEnabled,
	}

	imapAddressProton := proton.Address{
		ID:      "2",
		Email:   "imap@proton.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeOriginal,
	}

	senderAddressBYOE := proton.Address{
		ID:      "3",
		Email:   "sender@external.me",
		Send:    true,
		Receive: true,
		Type:    proton.AddressTypeExternal,
		Status:  proton.AddressStatusEnabled,
	}

	testConn := &testConnector{
		addressMode:    usertypes.AddressModeCombined,
		primaryAddress: primaryAddress,
		imapAddress:    imapAddressProton,
		senderAddress:  senderAddressBYOE,
	}

	// Sender address is BYOE, and we're creating a draft.
	// Expected: BYOE address is returned.
	testConn.senderAddress = senderAddressBYOE
	testConn.imapAddress = imapAddressProton
	addr, err := getImportAddress(nil, true, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, senderAddressBYOE.ID, addr.ID)
	require.Equal(t, senderAddressBYOE.Email, addr.Email)

	// Sender address is BYOE, and we're importing a message.
	// Expected: BYOE address is returned.
	addr, err = getImportAddress(nil, false, imapAddressProton.ID, testConn)
	require.NoError(t, err)
	require.Equal(t, senderAddressBYOE.ID, addr.ID)
	require.Equal(t, senderAddressBYOE.Email, addr.Email)
}
