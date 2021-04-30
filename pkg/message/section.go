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

// BodyStructure is used to parse an email into MIME sections and then generate
// body structure for IMAP server.
type BodyStructure map[string]*SectionInfo

// SectionInfo is used to hold data about parts of each section.
type SectionInfo struct {
	Header                    textproto.MIMEHeader
	Start, BSize, Size, Lines int
	reader                    io.Reader
}

// Read will also count the final size of section.
func (si *SectionInfo) Read(p []byte) (n int, err error) {
	n, err = si.reader.Read(p)
	si.Size += n
	si.Lines += bytes.Count(p, []byte("\n"))
	return
}

func NewBodyStructure(reader io.Reader) (structure *BodyStructure, err error) {
	structure = &BodyStructure{}
	err = structure.Parse(reader)
	return
}

// DeserializeBodyStructure will create new structure from msgpack bytes.
func DeserializeBodyStructure(raw []byte) (*BodyStructure, error) {
	bs := &BodyStructure{}
	err := msgpack.Unmarshal(raw, bs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot deserialize bodystructure")
	}
	return bs, err
}

// Serialize will write msgpack bytes.
func (bs *BodyStructure) Serialize() ([]byte, error) {
	data, err := msgpack.Marshal(bs)
	if err != nil {
		return nil, errors.Wrap(err, "cannot serialize bodystructure")
	}
	return data, nil
}

// Parse will read the mail and create all body structures.
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
	if (strings.HasPrefix(mediaType, "multipart/") || mediaType == "message/rfc822") && params["boundary"] != "" {
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
			err = br.writeNextPartTo(part)
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

// GetSection returns bytes of section including MIME header.
func (bs *BodyStructure) GetSection(wholeMail io.ReadSeeker, sectionPath []int) (section []byte, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	return goToOffsetAndReadNBytes(wholeMail, info.Start, info.Size)
}

// GetSectionContent returns bytes of section content (excluding MIME header).
func (bs *BodyStructure) GetSectionContent(wholeMail io.ReadSeeker, sectionPath []int) (section []byte, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	return goToOffsetAndReadNBytes(wholeMail, info.Start+info.Size-info.BSize, info.BSize)
}

// GetMailHeader returns the main header of mail.
func (bs *BodyStructure) GetMailHeader() (header textproto.MIMEHeader, err error) {
	return bs.GetSectionHeader([]int{})
}

// GetMailHeaderBytes returns the bytes with main mail header.
// Warning: It can contain extra lines or multipart comment.
func (bs *BodyStructure) GetMailHeaderBytes(wholeMail io.ReadSeeker) (header []byte, err error) {
	return bs.GetSectionHeaderBytes(wholeMail, []int{})
}

func goToOffsetAndReadNBytes(wholeMail io.ReadSeeker, offset, length int) ([]byte, error) {
	if length < 1 {
		return nil, errors.New("requested non positive length")
	}
	if offset > 0 {
		if _, err := wholeMail.Seek(int64(offset), io.SeekStart); err != nil {
			return nil, err
		}
	}
	out := make([]byte, length)
	_, err := wholeMail.Read(out)
	return out, err
}

// GetSectionHeader returns the mime header of specified section.
func (bs *BodyStructure) GetSectionHeader(sectionPath []int) (header textproto.MIMEHeader, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	header = info.Header
	return
}

func (bs *BodyStructure) GetSectionHeaderBytes(wholeMail io.ReadSeeker, sectionPath []int) (header []byte, err error) {
	info, err := bs.getInfo(sectionPath)
	if err != nil {
		return
	}
	headerLength := info.Size - info.BSize
	return goToOffsetAndReadNBytes(wholeMail, info.Start, headerLength)
}

// IMAPBodyStructure will prepare imap bodystructure recurently for given part.
// Use empty path to create whole email structure.
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
