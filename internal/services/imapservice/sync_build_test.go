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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package imapservice

import (
	"errors"
	"testing"
	"time"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/stretchr/testify/require"
)

func TestNewFailedMessageLiteral(t *testing.T) {
	literal := newFailedMessageLiteral("abcd-efgh", time.Unix(123456789, 0), "subject", errors.New("oops"))

	header, err := rfc822.Parse(literal).ParseHeader()
	require.NoError(t, err)
	require.Equal(t, "Message failed to build", header.Get("Subject"))
	require.Equal(t, "29 Nov 73 21:33 UTC", header.Get("Date"))
	require.Equal(t, "text/plain", header.Get("Content-Type"))
	require.Equal(t, "base64", header.Get("Content-Transfer-Encoding"))

	b, err := rfc822.Parse(literal).DecodedBody()
	require.NoError(t, err)
	require.Equal(t, string(b), "Failed to build message: \nSubject:   subject\nError:     oops\nMessageID: abcd-efgh\n")

	parsed, err := imap.NewParsedMessage(literal)
	require.NoError(t, err)
	require.Equal(t, `("29 Nov 73 21:33 UTC" "Message failed to build" NIL NIL NIL NIL NIL NIL NIL NIL)`, parsed.Envelope)
	require.Equal(t, `("text" "plain" () NIL NIL "base64" 114 2)`, parsed.Body)
	require.Equal(t, `("text" "plain" () NIL NIL "base64" 114 2 NIL NIL NIL NIL)`, parsed.Structure)
}
