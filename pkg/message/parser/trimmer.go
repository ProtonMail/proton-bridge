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

package parser

import (
	"bytes"
	"io"
)

const endOfMail = "\r\n.\r\n"

// endOfMailTrimmer wraps a reader to trim the End-Of-Mail indicator at the end
// of the input, if present.
//
// During SMTP sending of a message, the DATA command indicates that you are
// about to send the text (or body) of the message. The message text must end
// with "\r\n.\r\n." I'm 99% sure that these 5 bytes should not be considered
// part of the message body. However, some mail servers keep them as part of
// the message, which our parser sometimes doesn't like. Therefore, we strip
// them if we find them.
type endOfMailTrimmer struct {
	r   io.Reader
	buf bytes.Buffer
}

func newEndOfMailTrimmer(r io.Reader) *endOfMailTrimmer {
	return &endOfMailTrimmer{r: r}
}

func (r *endOfMailTrimmer) Read(p []byte) (int, error) {
	_, err := io.CopyN(&r.buf, r.r, int64(len(p)+len(endOfMail)-r.buf.Len()))
	if err != nil && err != io.EOF {
		return 0, err
	}

	if err == io.EOF && bytes.HasSuffix(r.buf.Bytes(), []byte(endOfMail)) {
		r.buf.Truncate(r.buf.Len() - len(endOfMail))
	}

	return r.buf.Read(p)
}
