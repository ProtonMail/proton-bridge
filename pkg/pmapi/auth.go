package pmapi

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/go-resty/resty/v2"
)

func (c *client) Auth2FA(ctx context.Context, req Auth2FAReq) error {
	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.SetBody(req).Post("/auth/2fa")
	}); err != nil {
		return err
	}

	return nil
}

func (c *client) AuthDelete(ctx context.Context) error {
	if _, err := c.do(ctx, func(r *resty.Request) (*resty.Response, error) {
		return r.Delete("/auth")
	}); err != nil {
		return err
	}

	c.uid, c.acc, c.ref, c.exp = "", "", "", time.Time{}

	// FIXME(conman): should we perhaps signal via AuthHandler that the auth was deleted?

	return nil
}

func (c *client) AuthSalt(ctx context.Context) (string, error) {
	salts, err := c.GetKeySalts(ctx)
	if err != nil {
		return "", err
	}

	if _, err := c.CurrentUser(ctx); err != nil {
		return "", err
	}

	for _, s := range salts {
		if s.ID == c.user.Keys[0].ID {
			return s.KeySalt, nil
		}
	}

	return "", errors.New("no matching salt found")
}

func (c *client) AddAuthHandler(handler AuthHandler) {
	c.authHandlers = append(c.authHandlers, handler)
}

func (c *client) authRefresh(ctx context.Context) error {
	c.authLocker.Lock()
	defer c.authLocker.Unlock()

	auth, err := c.req.authRefresh(ctx, c.uid, c.ref)
	if err != nil {
		return err
	}

	c.acc = auth.AccessToken
	c.ref = auth.RefreshToken

	for _, handler := range c.authHandlers {
		if err := handler(auth); err != nil {
			return err
		}
	}

	return nil
}

func randomString(length int) string {
	noise := make([]byte, length)

	if _, err := io.ReadFull(rand.Reader, noise); err != nil {
		panic(err)
	}

	return base64.StdEncoding.EncodeToString(noise)[:length]
}
