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
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gluon/rfc822"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/gopenpgp/v2/constants"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	imapservice "github.com/ProtonMail/proton-bridge/v3/internal/services/imapservice"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"github.com/ProtonMail/proton-bridge/v3/internal/vault"
	bmessage "github.com/ProtonMail/proton-bridge/v3/pkg/message"
	"github.com/bradenaw/juniper/xmaps"
	"github.com/bradenaw/juniper/xslices"
	"github.com/emersion/go-message"
	"github.com/sirupsen/logrus"
)

type DiagnosticMetadata struct {
	MessageIDs       []string
	Metadata         []proton.MessageMetadata
	FailedMessageIDs xmaps.Set[string]
}

type AccountMailboxMap map[string][]DiagMailboxMessage

type DiagMailboxMessage struct {
	AddressID string
	UserID    string
	ID        string
	Flags     imap.FlagSet
}

func (apm DiagnosticMetadata) BuildMailboxToMessageMap(ctx context.Context, user *User) (map[string]AccountMailboxMap, error) {
	apiAddrs, err := user.identityService.GetAddresses(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get addresses: %w", err)
	}

	apiLabels, err := user.imapService.GetLabels(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get labels: %w", err)
	}

	result := make(map[string]AccountMailboxMap)

	mode := user.GetAddressMode()
	primaryAddrID, err := usertypes.GetPrimaryAddr(apiAddrs)
	if err != nil {
		return nil, fmt.Errorf("failed to get primary addr for user: %w", err)
	}

	getAccount := func(addrID string) (AccountMailboxMap, bool) {
		if mode == vault.CombinedMode {
			addrID = primaryAddrID.ID
		}

		addr := apiAddrs[addrID]
		if addr.Status != proton.AddressStatusEnabled {
			return nil, false
		}

		v, ok := result[addr.Email]
		if !ok {
			result[addr.Email] = make(AccountMailboxMap)
			v = result[addr.Email]
		}

		return v, true
	}

	for _, metadata := range apm.Metadata {
		for _, label := range metadata.LabelIDs {
			details, ok := apiLabels[label]
			if !ok {
				logrus.Warnf("User %v has message with unknown label '%v'", user.Name(), label)
				continue
			}

			if !imapservice.WantLabel(details) {
				continue
			}

			account, enabled := getAccount(metadata.AddressID)
			if !enabled {
				continue
			}

			var mboxName string
			if details.Type == proton.LabelTypeSystem {
				mboxName = details.Name
			} else {
				mboxName = strings.Join(imapservice.GetMailboxName(details), "/")
			}

			mboxMessage := DiagMailboxMessage{
				UserID:    user.ID(),
				ID:        metadata.ID,
				AddressID: metadata.AddressID,
				Flags:     imapservice.BuildFlagSetFromMessageMetadata(metadata),
			}

			if v, ok := account[mboxName]; ok {
				account[mboxName] = append(v, mboxMessage)
			} else {
				account[mboxName] = []DiagMailboxMessage{mboxMessage}
			}
		}
	}
	return result, nil
}

func (user *User) GetDiagnosticMetadata(ctx context.Context) (DiagnosticMetadata, error) {
	failedMessages, err := user.imapService.GetSyncFailedMessageIDs(ctx)
	if err != nil {
		return DiagnosticMetadata{}, err
	}

	messageIDs, err := user.client.GetAllMessageIDs(ctx, "")
	if err != nil {
		return DiagnosticMetadata{}, err
	}

	meta := make([]proton.MessageMetadata, 0, len(messageIDs))

	for _, m := range xslices.Chunk(messageIDs, 100) {
		metadata, err := user.client.GetMessageMetadataPage(ctx, 0, len(m), proton.MessageFilter{ID: m})
		if err != nil {
			return DiagnosticMetadata{}, err
		}

		meta = append(meta, metadata...)
	}

	return DiagnosticMetadata{
		MessageIDs:       messageIDs,
		Metadata:         meta,
		FailedMessageIDs: xmaps.SetFromSlice(failedMessages),
	}, nil
}

func (user *User) DebugDownloadMessages(
	ctx context.Context,
	path string,
	msgs map[string]DiagMailboxMessage,
	progressCB func(string, int, int),
) error {
	total := len(msgs)
	userID := user.ID()

	apiUser, err := user.identityService.GetAPIUser(ctx)
	if err != nil {
		return fmt.Errorf("failed to get api user: %w", err)
	}

	apiAddrs, err := user.identityService.GetAddresses(ctx)
	if err != nil {
		return fmt.Errorf("failed to get address: %w", err)
	}

	counter := 1
	for _, msg := range msgs {
		if progressCB != nil {
			progressCB(userID, counter, total)
			counter++
		}

		msgDir := filepath.Join(path, msg.ID)
		if err := os.MkdirAll(msgDir, 0o700); err != nil {
			return fmt.Errorf("failed to create directory '%v':%w", msgDir, err)
		}

		message, err := user.client.GetFullMessage(ctx, msg.ID, usertypes.NewProtonAPIScheduler(user.panicHandler), proton.NewDefaultAttachmentAllocator())
		if err != nil {
			return fmt.Errorf("failed to download message '%v':%w", msg.ID, err)
		}

		if err := writeMetadata(msgDir, message.Message); err != nil {
			return err
		}

		if err := usertypes.WithAddrKR(apiUser, apiAddrs[msg.AddressID], user.vault.KeyPass(), func(_, addrKR *crypto.KeyRing) error {
			switch {
			case len(message.Attachments) > 0:
				return decodeMultipartMessage(msgDir, addrKR, message.Message, message.AttData)

			case message.MIMEType == "multipart/mixed":
				return decodePGPMessage(msgDir, addrKR, message.Message)

			default:
				return decodeSimpleMessage(msgDir, addrKR, message.Message)
			}
		}); err != nil {
			return err
		}
	}
	return nil
}

func TryBuildDebugMessage(path string) error {
	meta, err := loadDebugMetadata(path)
	if err != nil {
		return fmt.Errorf("failed to load metadata: %w", err)
	}

	body, bodyDecrypted, err := loadDebugBody(path)
	if err != nil {
		return fmt.Errorf("failed to load body: %w", err)
	}

	var da []bmessage.DecryptedAttachment
	if len(meta.Attachments) != 0 {
		d, err := loadAttachments(path, &meta)
		if err != nil {
			return err
		}
		da = d
	}

	decryptedMessage := bmessage.DecryptedMessage{
		Msg: proton.Message{
			MessageMetadata: meta.MessageMetadata,
			Header:          meta.Header,
			ParsedHeaders:   meta.ParsedHeaders,
			Body:            "",
			MIMEType:        meta.MIMEType,
			Attachments:     nil,
		},
		Body:        bytes.Buffer{},
		BodyErr:     nil,
		Attachments: da,
	}

	if bodyDecrypted {
		decryptedMessage.Body.Write(body)
	} else {
		decryptedMessage.Msg.Body = string(body)
		decryptedMessage.BodyErr = fmt.Errorf("body did not decrypt")
	}

	var rfc822Message bytes.Buffer
	if err := bmessage.BuildRFC822Into(nil, &decryptedMessage, defaultMessageJobOpts(), &rfc822Message); err != nil {
		return fmt.Errorf("failed to build message: %w", err)
	}

	return nil
}

func getBodyName(path string) string {
	return filepath.Join(path, "body.txt")
}

func getBodyNameFailed(path string) string {
	return filepath.Join(path, "body_failed.txt")
}

func getBodyNamePGP(path string) string {
	return filepath.Join(path, "body.pgp")
}

func getMetadataPath(path string) string {
	return filepath.Join(path, "metadata.json")
}

func getAttachmentPathSuccess(path, id, name string) string {
	return filepath.Join(path, fmt.Sprintf("attachment_%v_%v", id, name))
}

func getAttachmentPathFailure(path, id string) string {
	return filepath.Join(path, fmt.Sprintf("attachment_%v_failed.pgp", id))
}

func decodeMultipartMessage(outPath string, kr *crypto.KeyRing, msg proton.Message, attData [][]byte) error {
	for idx, attachment := range msg.Attachments {
		if err := decodeAttachment(outPath, kr, attachment, attData[idx]); err != nil {
			return fmt.Errorf("failed to decode attachment %v of message %v: %w", attachment.ID, msg.ID, err)
		}
	}

	return decodeSimpleMessage(outPath, kr, msg)
}

func decodePGPMessage(outPath string, kr *crypto.KeyRing, msg proton.Message) error {
	var decrypted bytes.Buffer
	decrypted.Grow(len(msg.Body))

	if err := msg.DecryptInto(kr, &decrypted); err != nil {
		logrus.Warnf("Failed to decrypt pgp message %v, storing as is: %v", msg.ID, err)
		bodyPath := getBodyNamePGP(outPath)
		if err := os.WriteFile(bodyPath, []byte(msg.Body), 0o600); err != nil {
			return fmt.Errorf("failed to write pgp body to '%v': %w", bodyPath, err)
		}
	}

	bodyPath := getBodyName(outPath)

	if err := os.WriteFile(bodyPath, decrypted.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write pgp body to '%v': %w", bodyPath, err)
	}

	return nil
}

func decodeSimpleMessage(outPath string, kr *crypto.KeyRing, msg proton.Message) error {
	var decrypted bytes.Buffer
	decrypted.Grow(len(msg.Body))

	if err := msg.DecryptInto(kr, &decrypted); err != nil {
		logrus.Warnf("Failed to decrypt simple message %v, will try again as attachment : %v", msg.ID, err)
		return writeCustomTextPart(getBodyNameFailed(outPath), msg, err)
	}

	bodyPath := getBodyName(outPath)

	if err := os.WriteFile(bodyPath, decrypted.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write simple body to '%v': %w", bodyPath, err)
	}

	return nil
}

type DebugMetadata struct {
	proton.MessageMetadata
	Header        string
	ParsedHeaders proton.Headers
	MIMEType      rfc822.MIMEType
	Attachments   []proton.Attachment
}

func writeMetadata(outPath string, msg proton.Message) error {
	metadata := DebugMetadata{
		MessageMetadata: msg.MessageMetadata,
		Header:          msg.Header,
		ParsedHeaders:   msg.ParsedHeaders,
		MIMEType:        msg.MIMEType,
		Attachments:     msg.Attachments,
	}

	j, err := json.MarshalIndent(metadata, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode json for message %v: %w", msg.ID, err)
	}

	metaPath := getMetadataPath(outPath)

	if err := os.WriteFile(metaPath, j, 0o600); err != nil {
		return fmt.Errorf("failed to write metadata to '%v': %w", metaPath, err)
	}

	return nil
}

func decodeAttachment(outPath string, kr *crypto.KeyRing,
	att proton.Attachment,
	attData []byte) error {
	kps, err := base64.StdEncoding.DecodeString(att.KeyPackets)
	if err != nil {
		return err
	}

	// Use io.Multi
	attachmentReader := io.MultiReader(bytes.NewReader(kps), bytes.NewReader(attData))

	stream, err := kr.DecryptStream(attachmentReader, nil, crypto.GetUnixTime())
	if err != nil {
		logrus.
			WithField("attID", att.ID).
			WithError(err).
			Warn("Attachment decryption failed - construct")

		var pgpMessageBuffer bytes.Buffer
		pgpMessageBuffer.Grow(len(kps) + len(attData))
		pgpMessageBuffer.Write(kps)
		pgpMessageBuffer.Write(attData)

		return writeCustomAttachmentPart(getAttachmentPathFailure(outPath, att.ID), att, &crypto.PGPMessage{Data: pgpMessageBuffer.Bytes()}, err)
	}

	var decryptBuffer bytes.Buffer
	decryptBuffer.Grow(len(kps) + len(attData))

	if _, err := decryptBuffer.ReadFrom(stream); err != nil {
		logrus.
			WithField("attID", att.ID).
			WithError(err).
			Warn("Attachment decryption failed - stream")

		var pgpMessageBuffer bytes.Buffer
		pgpMessageBuffer.Grow(len(kps) + len(attData))
		pgpMessageBuffer.Write(kps)
		pgpMessageBuffer.Write(attData)

		return writeCustomAttachmentPart(getAttachmentPathFailure(outPath, att.ID), att, &crypto.PGPMessage{Data: pgpMessageBuffer.Bytes()}, err)
	}

	attachmentPath := getAttachmentPathSuccess(outPath, att.ID, att.Name)

	if err := os.WriteFile(attachmentPath, decryptBuffer.Bytes(), 0o600); err != nil {
		return fmt.Errorf("failed to write attachment %v to '%v': %w", att.ID, attachmentPath, err)
	}

	return nil
}

func writeCustomTextPart(
	outPath string,
	msg proton.Message,
	decError error,
) error {
	enc, err := crypto.NewPGPMessageFromArmored(msg.Body)
	if err != nil {
		return err
	}

	arm, err := enc.GetArmoredWithCustomHeaders(
		fmt.Sprintf("This message could not be decrypted: %v", decError),
		constants.ArmorHeaderVersion,
	)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outPath, []byte(arm), 0o600); err != nil {
		return fmt.Errorf("failed to write custom message %v data to '%v': %w", msg.ID, outPath, err)
	}

	return nil
}

// writeCustomAttachmentPart writes an armored-PGP data part for an attachment that couldn't be decrypted.
func writeCustomAttachmentPart(
	outPath string,
	att proton.Attachment,
	msg *crypto.PGPMessage,
	decError error,
) error {
	arm, err := msg.GetArmoredWithCustomHeaders(
		fmt.Sprintf("This attachment could not be decrypted: %v", decError),
		constants.ArmorHeaderVersion,
	)
	if err != nil {
		return err
	}

	filename := mime.QEncoding.Encode("utf-8", att.Name+".pgp")

	var hdr message.Header

	hdr.SetContentType("application/octet-stream", map[string]string{"name": filename})
	hdr.SetContentDisposition(string(att.Disposition), map[string]string{"filename": filename})

	if err := os.WriteFile(outPath, []byte(arm), 0o600); err != nil {
		return fmt.Errorf("failed to write custom attachment %v part  to '%v': %w", att.ID, outPath, err)
	}

	return nil
}

func loadDebugMetadata(dir string) (DebugMetadata, error) {
	metadataPath := getMetadataPath(dir)
	b, err := os.ReadFile(metadataPath) //nolint:gosec
	if err != nil {
		return DebugMetadata{}, err
	}

	var m DebugMetadata

	if err := json.Unmarshal(b, &m); err != nil {
		return DebugMetadata{}, err
	}

	return m, nil
}

func loadDebugBody(dir string) ([]byte, bool, error) {
	if b, err := os.ReadFile(getBodyName(dir)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, false, err
		}
	} else {
		return b, true, nil
	}

	if b, err := os.ReadFile(getBodyNameFailed(dir)); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return nil, false, err
		}
	} else {
		return b, false, nil
	}

	return nil, false, fmt.Errorf("body is either pgp message, which we can't handle or is missing")
}

func loadAttachments(dir string, meta *DebugMetadata) ([]bmessage.DecryptedAttachment, error) {
	attDecrypted := make([]bmessage.DecryptedAttachment, 0, len(meta.Attachments))

	for _, a := range meta.Attachments {
		data, err := os.ReadFile(getAttachmentPathSuccess(dir, a.ID, a.Name))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("attachment (%v,%v) must have failed to decrypt, we can't do anything since we need the user's keyring", a.ID, a.Name)
			}

			return nil, fmt.Errorf("failed to load attachment (%v,%v): %w", a.ID, a.Name, err)
		}

		da := bmessage.DecryptedAttachment{
			Packet:    nil,
			Encrypted: nil,
			Data:      bytes.Buffer{},
			Err:       nil,
		}

		da.Data.Write(data)

		attDecrypted = append(attDecrypted, da)
	}

	return attDecrypted, nil
}

func defaultMessageJobOpts() bmessage.JobOptions {
	return bmessage.JobOptions{
		IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
		SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
		AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
		AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
		AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
		AddMessageIDReference:  true, // Whether to include the MessageID in References.
	}
}
