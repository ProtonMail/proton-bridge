// Copyright (c) 2026 Proton AG
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

package parentid

import (
	"context"
	"errors"
	"testing"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/stretchr/testify/require"
)

// fakeClient is a tiny in-memory stub for the narrow Client interface.
//
// Internal lookups (filter.ID set) resolve from the byID map.
// External lookups (filter.ExternalID set) resolve from the byExternalID map.
// addressIDSeen records the AddressID arg passed by Find, so address-mode
// behaviour can be asserted without mocking the broader APIClient surface.
type fakeClient struct {
	byID         map[string][]proton.MessageMetadata
	byExternalID map[string][]proton.MessageMetadata
	err          error

	addressIDSeen []string
}

func (f *fakeClient) GetMessageMetadata(_ context.Context, filter proton.MessageFilter) ([]proton.MessageMetadata, error) {
	f.addressIDSeen = append(f.addressIDSeen, filter.AddressID)
	if f.err != nil {
		return nil, f.err
	}
	if len(filter.ID) > 0 {
		return f.byID[filter.ID[0]], nil
	}
	if filter.ExternalID != "" {
		return f.byExternalID[filter.ExternalID], nil
	}
	return nil, nil
}

const internalDomain = "@" + message.InternalIDDomain

// References reach Find as bare message-id values (no angle brackets) —
// pkg/message's parser strips them before populating Message.References /
// Message.InReplyTo.
func internalRef(id string) string {
	return id + internalDomain
}

func draft(id string) proton.MessageMetadata {
	// A draft carries neither Sent nor Received; IsDraft() relies on the
	// Flags being unset on those bits (see go-proton-api).
	return proton.MessageMetadata{ID: id}
}

func sent(id string) proton.MessageMetadata {
	return proton.MessageMetadata{ID: id, Flags: proton.MessageFlagSent}
}

func received(id string) proton.MessageMetadata {
	return proton.MessageMetadata{ID: id, Flags: proton.MessageFlagReceived}
}

func TestFind_NoReferences(t *testing.T) {
	f := &fakeClient{}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined, nil)

	require.NoError(t, err)
	require.Equal(t, "", parentID)
	require.Empty(t, drafts)
	// No references → no API calls.
	require.Empty(t, f.addressIDSeen)
}

func TestFind_InternalReference_ResolvesToReceivedParent(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"msg-internal-1": {received("msg-internal-1")},
		},
	}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{internalRef("msg-internal-1")},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-internal-1", parentID)
	require.Empty(t, drafts)
}

func TestFind_InternalReference_DraftRecordedNotPicked(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"draft-1": {draft("draft-1")},
		},
	}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{internalRef("draft-1")},
	)

	require.NoError(t, err)
	require.Equal(t, "", parentID, "draft must not be picked as parent (GODT-2935 rule)")
	require.Equal(t, []string{"draft-1"}, drafts)
}

func TestFind_InternalReferences_MixedDraftAndReceived(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"draft-a":   {draft("draft-a")},
			"msg-b":     {received("msg-b")},
			"draft-c":   {draft("draft-c")},
		},
	}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{internalRef("draft-a"), internalRef("msg-b"), internalRef("draft-c")},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-b", parentID, "the non-draft internal reference must win")
	require.ElementsMatch(t, []string{"draft-a", "draft-c"}, drafts)
}

func TestFind_ExternalReference_ResolvesToSent(t *testing.T) {
	f := &fakeClient{
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {sent("msg-ext-1")},
		},
	}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{"ext@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-ext-1", parentID)
	require.Empty(t, drafts)
}

func TestFind_ExternalReference_ResolvesToReceived(t *testing.T) {
	f := &fakeClient{
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {received("msg-ext-2")},
		},
	}
	parentID, _, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{"ext@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-ext-2", parentID)
}

func TestFind_ExternalReference_DraftIgnored(t *testing.T) {
	f := &fakeClient{
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {draft("draft-x")},
		},
	}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{"ext@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "", parentID, "an external draft must not be picked as parent")
	require.Empty(t, drafts, "external drafts are not tracked for deletion")
}

func TestFind_ExternalReference_OnlyLastReferenceIsLookedUp(t *testing.T) {
	f := &fakeClient{
		byExternalID: map[string][]proton.MessageMetadata{
			"first@example.com":  {sent("msg-first")},
			"latest@example.com": {sent("msg-latest")},
		},
	}
	parentID, _, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{"first@example.com", "middle@example.com", "latest@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-latest", parentID, "Find resolves only the last external reference per the chain")
}

func TestFind_ExternalReference_MultipleResults_PicksSentOne(t *testing.T) {
	f := &fakeClient{
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {
				draft("draft-1"),
				sent("msg-sent"),
				received("msg-received"),
			},
		},
	}
	parentID, _, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{"ext@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-sent", parentID, "with multiple results, the Sent message wins")
}

func TestFind_InternalParentFound_DoesNotFallBackToExternal(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"msg-internal": {received("msg-internal")},
		},
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {sent("msg-external")},
		},
	}
	parentID, _, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{internalRef("msg-internal"), "ext@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-internal", parentID, "internal parent short-circuits the external lookup")
}

func TestFind_InternalDraftOnly_FallsBackToExternal(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"draft-1": {draft("draft-1")},
		},
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {sent("msg-external")},
		},
	}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{internalRef("draft-1"), "ext@example.com"},
	)

	require.NoError(t, err)
	require.Equal(t, "msg-external", parentID, "no internal parent → external lookup runs")
	require.Equal(t, []string{"draft-1"}, drafts, "internal draft is still recorded for deletion")
}

func TestFind_AddressModeSplit_PassesAddrIDToAPI(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"msg-1": {received("msg-1")},
		},
		byExternalID: map[string][]proton.MessageMetadata{
			"ext@example.com": {sent("msg-2")},
		},
	}
	_, _, err := Find(context.Background(), f, "addr-XYZ", usertypes.AddressModeSplit,
		[]string{internalRef("msg-1"), "ext@example.com"},
	)

	require.NoError(t, err)
	// Only the internal lookup runs (parent already found), so we expect one
	// API call with the split-mode address.
	require.Equal(t, []string{"addr-XYZ"}, f.addressIDSeen)
}

func TestFind_AddressModeCombined_OmitsAddrID(t *testing.T) {
	f := &fakeClient{
		byID: map[string][]proton.MessageMetadata{
			"msg-1": {received("msg-1")},
		},
	}
	_, _, err := Find(context.Background(), f, "addr-XYZ", usertypes.AddressModeCombined,
		[]string{internalRef("msg-1")},
	)

	require.NoError(t, err)
	require.Equal(t, []string{""}, f.addressIDSeen, "combined mode must not scope the lookup to a single address")
}

func TestFind_APIErrorPropagates_InternalLookup(t *testing.T) {
	f := &fakeClient{err: errors.New("api down")}
	parentID, drafts, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{internalRef("msg-1")},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get message metadata")
	require.Equal(t, "", parentID)
	require.Empty(t, drafts)
}

func TestFind_APIErrorPropagates_ExternalLookup(t *testing.T) {
	f := &fakeClient{
		// Internal lookup succeeds (empty result), external lookup hits the error.
		byID: map[string][]proton.MessageMetadata{},
		err:  errors.New("api down"),
	}
	_, _, err := Find(context.Background(), f, "addr-1", usertypes.AddressModeCombined,
		[]string{"ext@example.com"},
	)

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to get message metadata")
}
