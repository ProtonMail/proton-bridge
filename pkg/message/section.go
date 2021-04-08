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

package message

import (
	"bufio"
	"bytes"
	"io"
	"io/ioutil"
	"net/textproto"
	"strconv"
	"strings"

	pmmime "github.com/ProtonMail/proton-bridge/pkg/mime"
	"github.com/emersion/go-imap"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack/v5"
)

type SectionInfo struct {
	Header                    textproto.MIMEHeader
	Start, BSize, Size, Lines int
	reader                    io.Reader
}

// Read and count.
func (si *SectionInfo) Read(p []byte) (n int, err error) {
	n, err = si.reader.Read(p)
	si.Size += n
	si.Lines += bytes.Count(p, []byte("\n"))
	return
}

type boundaryReader struct {
	reader *bufio.Reader

	closed, first bool
	skipped       int

	nl               []byte // "\r\n" or "\n" (set after seeing first boundary line)
	nlDashBoundary   []byte // nl + "--boundary"
	dashBoundaryDash []byte // "--boundary--"
	dashBoundary     []byte // "--boundary"
}

func newBoundaryReader(r *bufio.Reader, boundary string) (br *boundaryReader, err error) {
	b := []byte("\r\n--" + boundary + "--")
	br = &boundaryReader{
		reader:           r,
		closed:           false,
		first:            true,
		nl:               b[:2],
		nlDashBoundary:   b[:len(b)-2],
		dashBoundaryDash: b[2:],
		dashBoundary:     b[2 : len(b)-2],
	}
	err = br.WriteNextPartTo(nil)
	return
}

func skipLWSPChar(b []byte) []byte {
	for len(b) > 0 && (b[0] == ' ' || b[0] == '\t') {
		b = b[1:]
	}
	return b
}

func (br *boundaryReader) isFinalBoundary(line []byte) bool {
	if !bytes.HasPrefix(line, br.dashBoundaryDash) {
		return false
	}
	rest := line[len(br.dashBoundaryDash):]
	rest = skipLWSPChar(rest)
	return len(rest) == 0 || bytes.Equal(rest, br.nl)
}

func (br *boundaryReader) isBoundaryDelimiterLine(line []byte) (ret bool) {
	if !bytes.HasPrefix(line, br.dashBoundary) {
		return false
	}
	rest := line[len(br.dashBoundary):]
	rest = skipLWSPChar(rest)

	if br.first && len(rest) == 1 && rest[0] == '\n' {
		br.nl = br.nl[1:]
		br.nlDashBoundary = br.nlDashBoundary[1:]
	}
	return bytes.Equal(rest, br.nl)
}

func (br *boundaryReader) WriteNextPartTo(part io.Writer) (err error) {
	if br.closed {
		return io.EOF
	}

	var line, slice []byte
	br.skipped = 0

	for {
		slice, err = br.reader.ReadSlice('\n')
		line = append(line, slice...)
		if err == bufio.ErrBufferFull {
			continue
		}

		br.skipped += len(line)

		if err == io.EOF && br.isFinalBoundary(line) {
			err = nil
			br.closed = true
			return
		}

		if err != nil {
			return
		}

		if br.isBoundaryDelimiterLine(line) {
			br.first = false
			return
		}

		if br.isFinalBoundary(line) {
			br.closed = true
			return
		}

		if part != nil {
			if _, err = part.Write(line); err != nil {
				return
			}
		}

		line = []byte{}
	}
}

type BodyStructure map[string]*SectionInfo

func NewBodyStructure(reader io.Reader) (structure *BodyStructure, err error) {
	structure = &BodyStructure{}
	err = structure.Parse(reader)
	return
}

func DeserializeBodyStructure(raw []byte) (*BodyStructure, error) {
	bs := &BodyStructure{}
	err := msgpack.Unmarshal(raw, bs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot deserialize bodystructure")
	}
	return bs, err
}

func (bs *BodyStructure) Serialize() ([]byte, error) {
	data, err := msgpack.Marshal(bs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot serialize bodystructure")
	}
	return data, nil
}

func (bs *BodyStructure) Parse(r io.Reader) error {
	return bs.parseAllChildSections(r, []int{}, 0)
}

func (bs *BodyStructure) parseAllChildSections(r io.Reader, currentPath []int, start int) (err error) { //nolint[funlen]
	info := &SectionInfo{
		Start:  start,
		Size:   0,
		BSize:  0,
		Lines:  0,
		reader: r,
	}

	bufInfo := bufio.NewReader(info)
	tp := textproto.NewReader(bufInfo)

	if info.Header, err = tp.ReadMIMEHeader(); err != nil {
		return
	}

	bodyInfo := &SectionInfo{reader: tp.R}
	bodyReader := bufio.NewReader(bodyInfo)

	mediaType, params, _ := pmmime.ParseMediaType(info.Header.Get("Content-Type"))

	// If multipart, call getAllParts, else read to count lines.
	if (strings.HasPrefix(mediaType, "multipart/") || mediaType == rfc822Message) && params["boundary"] != "" {
		newPath := append(currentPath, 1)

		var br *boundaryReader
		br, err = newBoundaryReader(bodyReader, params["boundary"])
		// New reader seeks first boundary.
		if err != nil {
			// Return also EOF.
			return
		}

		for err == nil {
			start += br.skipped
			part := &bytes.Buffer{}
			err = br.WriteNextPartTo(part)
			if err != nil {
				break
			}
			err = bs.parseAllChildSections(part, newPath, start)
			part.Reset()
			newPath[len(newPath)-1]++
		}
		br.reader = nil

		if err == io.EOF {
			err = nil
		}
		if err != nil {
			return
		}
	} else {
		// Count length.
		_, _ = bodyReader.WriteTo(ioutil.Discard)
	}

	// Clear all buffers.
	bodyReader = nil //nolint[wastedassign] just to be sure we clear garbage collector
	bodyInfo.reader = nil
	tp.R = nil
	tp = nil      //nolint[wastedassign] just to be sure we clear garbage collector
	bufInfo = nil //nolint[ineffassign] just to be sure we clear garbage collector
	info.reader = nil

	// Store boundaries.
	info.BSize = bodyInfo.Size
	path := stringPathFromInts(currentPath)
	(*bs)[path] = info

	// Fix start of subsections.
	newPath := append(currentPath, 1)
	shift := info.Size - info.BSize
	subInfo, err := bs.getInfo(newPath)

	// If it has subparts.
	for err == nil {
		subInfo.Start += shift

		// Level down.
		subInfo, err = bs.getInfo(append(newPath, 1))
		if err == nil {
			newPath = append(newPath, 1)
			continue
		}

		// Next.
		newPath[len(newPath)-1]++
		subInfo, err = bs.getInfo(newPath)
		if err == nil {
			continue
		}

		// Level up.
		for {
			newPath = newPath[:len(newPath)-1]
			if len(newPath) > 0 {
				newPath[len(newPath)-1]++
				subInfo, err = bs.getInfo(newPath)
				if err != nil {
					err = nil
					continue
				}
			}
			break
		}

		// The end.
		if len(newPath) == 0 {
			break
		}
	}

	return nil
}

func stringPathFromInts(ints []int) (ret string) {
	for i, n := range ints {
		if i != 0 {
			ret += "."
		}
		ret += strconv.Itoa(n)
	}
	return
}

func (bs *BodyStructure) hasInfo(sectionPath []int) bool {
	_, err := bs.getInfo(sectionPath)
	return err == nil
}

func (bs *BodyStructure) getInfo(sectionPath []int) (sectionInfo *SectionInfo, err error) {
	path := stringPathFromInts(sectionPath)
	sectionInfo, ok := (*bs)[path]
	if !ok {
		err = errors.New("wrong section " + path)
	}
	return
}

func (bs *BodyStructure) GetSection(wholeMail io.ReadSeeker, sectionPath []int) (section []byte, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	if _, err = wholeMail.Seek(int64(info.Start), io.SeekStart); err != nil {
		return
	}
	section = make([]byte, info.Size)
	_, err = wholeMail.Read(section)
	return
}

func (bs *BodyStructure) GetSectionContent(wholeMail io.ReadSeeker, sectionPath []int) (section []byte, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	if _, err = wholeMail.Seek(int64(info.Start+info.Size-info.BSize), io.SeekStart); err != nil {
		return
	}
	section = make([]byte, info.BSize)
	_, err = wholeMail.Read(section)
	return

	/* This is slow:
	sectionBuf, err := bs.GetSection(wholeMail, sectionPath)
	if err != nil {
		return
	}

	tp := textproto.NewReader(bufio.NewReader(buf))
	if _, err = tp.ReadMIMEHeader(); err != nil {
		return err
	}

	sectionBuf = &bytes.Buffer{}
	_, err = io.Copy(sectionBuf, tp.R)
	return
	*/
}

func (bs *BodyStructure) GetSectionHeader(sectionPath []int) (header textproto.MIMEHeader, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	header = info.Header
	return
}

func (bs *BodyStructure) IMAPBodyStructure(currentPart []int) (imapBS *imap.BodyStructure, err error) {
	var info *SectionInfo
	if info, err = bs.getInfo(currentPart); err != nil {
		return
	}

	mediaType, params, _ := pmmime.ParseMediaType(info.Header.Get("Content-Type"))

	mediaTypeSep := strings.Split(mediaType, "/")

	// If it is empty or missing it will not crash.
	mediaTypeSep = append(mediaTypeSep, "")

	imapBS = &imap.BodyStructure{
		MIMEType:    mediaTypeSep[0],
		MIMESubType: mediaTypeSep[1],
		Params:      params,
		Size:        uint32(info.BSize),
		Lines:       uint32(info.Lines),
	}

	if val := info.Header.Get("Content-ID"); val != "" {
		imapBS.Id = val
	}

	if val := info.Header.Get("Content-Transfer-Encoding"); val != "" {
		imapBS.Encoding = val
	}

	if val := info.Header.Get("Content-Description"); val != "" {
		imapBS.Description = val
	}

	if val := info.Header.Get("Content-Disposition"); val != "" {
		imapBS.Disposition = val
	}

	nextPart := append(currentPart, 1)
	for {
		if !bs.hasInfo(nextPart) {
			break
		}
		var subStruct *imap.BodyStructure
		subStruct, err = bs.IMAPBodyStructure(nextPart)
		if err != nil {
			return
		}
		if imapBS.Parts == nil {
			imapBS.Parts = []*imap.BodyStructure{}
		}
		imapBS.Parts = append(imapBS.Parts, subStruct)
		nextPart[len(nextPart)-1]++
	}

	return imapBS, nil
}
