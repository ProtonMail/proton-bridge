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

package user

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
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

func writeMetadata(outPath string, msg proton.Message) error {
	type CustomMetadata struct {
		proton.MessageMetadata
		Header        string
		ParsedHeaders proton.Headers
		MIMEType      rfc822.MIMEType
		Attachments   []proton.Attachment
	}

	metadata := CustomMetadata{
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
