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
	"context"
	"testing"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice/mocks"
	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func TestFixGODT3003Labels(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	log := logrus.WithField("test", "test")

	sharedLabels := newRWLabels()
	wr := sharedLabels.Write()
	wr.SetLabel("foo", proton.Label{
		ID:       "foo",
		ParentID: "bar",
		Name:     "Foo",
		Path:     []string{"bar", "Foo"},
		Color:    "",
		Type:     proton.LabelTypeFolder,
	}, "")

	wr.SetLabel("0", proton.Label{
		ID:       "0",
		ParentID: "",
		Name:     "Inbox",
		Path:     []string{"Inbox"},
		Color:    "",
		Type:     proton.LabelTypeSystem,
	}, "")

	wr.SetLabel("bar", proton.Label{
		ID:       "bar",
		ParentID: "",
		Name:     "boo",
		Path:     []string{"bar"},
		Color:    "",
		Type:     proton.LabelTypeFolder,
	}, "")

	wr.SetLabel("my_label", proton.Label{
		ID:       "my_label",
		ParentID: "",
		Name:     "MyLabel",
		Path:     []string{"MyLabel"},
		Color:    "",
		Type:     proton.LabelTypeLabel,
	}, "")

	wr.SetLabel("my_label2", proton.Label{
		ID:       "my_label2",
		ParentID: "",
		Name:     "MyLabel2",
		Path:     []string{labelPrefix, "MyLabel2"},
		Color:    "",
		Type:     proton.LabelTypeLabel,
	}, "")
	wr.Close()

	mboxs := []imap.MailboxNoAttrib{
		{
			ID:   "0",
			Name: []string{"Inbox"},
		},
		{
			ID:   "bar",
			Name: []string{"bar"},
		},
		{
			ID:   "foo",
			Name: []string{"bar", "Foo"},
		},
		{
			ID:   "my_label",
			Name: []string{"MyLabel"},
		},
		{
			ID:   "my_label2",
			Name: []string{labelPrefix, "MyLabel2"},
		},
	}

	rd := sharedLabels.Read()
	defer rd.Close()

	imapState := mocks.NewMockIMAPStateWrite(mockCtrl)

	imapState.EXPECT().PatchMailboxHierarchyWithoutTransforms(gomock.Any(), gomock.Eq(imap.MailboxID("bar")), gomock.Eq([]string{folderPrefix, "bar"}))
	imapState.EXPECT().PatchMailboxHierarchyWithoutTransforms(gomock.Any(), gomock.Eq(imap.MailboxID("foo")), gomock.Eq([]string{folderPrefix, "bar", "Foo"}))
	imapState.EXPECT().PatchMailboxHierarchyWithoutTransforms(gomock.Any(), gomock.Eq(imap.MailboxID("my_label")), gomock.Eq([]string{labelPrefix, "MyLabel"}))

	applied, err := fixGODT3003Labels(context.Background(), log, mboxs, rd, imapState)
	require.NoError(t, err)
	require.True(t, applied)
}

func TestFixGODT3003Labels_Noop(t *testing.T) {
	mockCtrl := gomock.NewController(t)

	log := logrus.WithField("test", "test")

	sharedLabels := newRWLabels()
	wr := sharedLabels.Write()
	wr.SetLabel("foo", proton.Label{
		ID:       "foo",
		ParentID: "bar",
		Name:     "Foo",
		Path:     []string{folderPrefix, "bar", "Foo"},
		Color:    "",
		Type:     proton.LabelTypeFolder,
	}, "")

	wr.SetLabel("0", proton.Label{
		ID:       "0",
		ParentID: "",
		Name:     "Inbox",
		Path:     []string{"Inbox"},
		Color:    "",
		Type:     proton.LabelTypeSystem,
	}, "")

	wr.SetLabel("bar", proton.Label{
		ID:       "bar",
		ParentID: "",
		Name:     "bar",
		Path:     []string{folderPrefix, "bar"},
		Color:    "",
		Type:     proton.LabelTypeFolder,
	}, "")

	wr.SetLabel("my_label", proton.Label{
		ID:       "my_label",
		ParentID: "",
		Name:     "MyLabel",
		Path:     []string{labelPrefix, "MyLabel"},
		Color:    "",
		Type:     proton.LabelTypeLabel,
	}, "")

	wr.SetLabel("my_label2", proton.Label{
		ID:       "my_label2",
		ParentID: "",
		Name:     "MyLabel2",
		Path:     []string{labelPrefix, "MyLabel2"},
		Color:    "",
		Type:     proton.LabelTypeLabel,
	}, "")
	wr.Close()

	mboxs := []imap.MailboxNoAttrib{
		{
			ID:   "0",
			Name: []string{"Inbox"},
		},
		{
			ID:   "bar",
			Name: []string{folderPrefix, "bar"},
		},
		{
			ID:   "foo",
			Name: []string{folderPrefix, "bar", "Foo"},
		},
		{
			ID:   "my_label",
			Name: []string{labelPrefix, "MyLabel"},
		},
		{
			ID:   "my_label2",
			Name: []string{labelPrefix, "MyLabel2"},
		},
	}

	rd := sharedLabels.Read()
	defer rd.Close()

	imapState := mocks.NewMockIMAPStateWrite(mockCtrl)
	applied, err := fixGODT3003Labels(context.Background(), log, mboxs, rd, imapState)
	require.NoError(t, err)
	require.False(t, applied)
}

func TestStripPlusAlias(t *testing.T) {
	cases := map[string]string{
		"one@three.com":     "one@three.com",
		"one+two@three.com": "one@three.com",
		"one@three+two.com": "one@three+two.com",
		"+one@three.com":    "+one@three.com",
		"@three.com":        "@three.com",
	}

	for given, want := range cases {
		require.Equal(t, want, stripPlusAlias(given), "input was %q", given)
	}
}

func TestEqualAddresse(t *testing.T) {
	cases := []struct {
		a, b string
		want bool
	}{
		{"one@three.com", "one@three.com", true},
		{"one@three.com", "one+two@three.com", true},
		{"OnE@thReE.com", "One@THree.com", true},
		{"one@three.com", "two@three.com", false},
		{"one+two@three.com", "two@three.com", false},
		{"one@three.com", "one@three+two.com", false},
	}

	for _, c := range cases {
		require.Equal(t, c.want, equalAddresses(c.a, c.b), "input was %q and %q", c.a, c.b)
	}
}
