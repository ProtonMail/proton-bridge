package pmapi

import (
	"context"
	"errors"
)

// Report sends request as json or multipart (if has attachment).
func (m *manager) ReportBug(context.Context, ReportBugReq) error {
	// FIXME(conman): Implement.
	return errors.New("not implemented")
}
