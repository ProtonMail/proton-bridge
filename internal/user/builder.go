package user

import (
	"context"

	"github.com/ProtonMail/gluon/imap"
	"github.com/ProtonMail/gopenpgp/v2/crypto"
	"github.com/ProtonMail/proton-bridge/v2/internal/pool"
	"github.com/ProtonMail/proton-bridge/v2/pkg/message"
	"gitlab.protontech.ch/go/liteapi"
)

type request struct {
	messageID string
	addrKR    *crypto.KeyRing
}

type fetcher interface {
	GetMessage(context.Context, string) (liteapi.Message, error)
	GetAttachment(context.Context, string) ([]byte, error)
}

func newBuilder(f fetcher, msgWorkers, attWorkers int) *pool.Pool[request, *imap.MessageCreated] {
	attPool := pool.New(attWorkers, func(ctx context.Context, attID string) ([]byte, error) {
		return f.GetAttachment(ctx, attID)
	})

	msgPool := pool.New(msgWorkers, func(ctx context.Context, req request) (*imap.MessageCreated, error) {
		msg, err := f.GetMessage(ctx, req.messageID)
		if err != nil {
			return nil, err
		}

		var attIDs []string

		for _, att := range msg.Attachments {
			attIDs = append(attIDs, att.ID)
		}

		attData, err := attPool.ProcessAll(ctx, attIDs)
		if err != nil {
			return nil, err
		}

		literal, err := message.BuildRFC822(req.addrKR, msg, attData, message.JobOptions{
			IgnoreDecryptionErrors: true, // Whether to ignore decryption errors and create a "custom message" instead.
			SanitizeDate:           true, // Whether to replace all dates before 1970 with RFC822's birthdate.
			AddInternalID:          true, // Whether to include MessageID as X-Pm-Internal-Id.
			AddExternalID:          true, // Whether to include ExternalID as X-Pm-External-Id.
			AddMessageDate:         true, // Whether to include message time as X-Pm-Date.
			AddMessageIDReference:  true, // Whether to include the MessageID in References.
		})
		if err != nil {
			return nil, err
		}

		return getMessageCreatedUpdate(msg, literal)
	})

	return msgPool
}
