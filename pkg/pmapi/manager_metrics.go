package pmapi

import (
	"context"
	"errors"
)

func (m *manager) SendSimpleMetric(context.Context, string, string, string) error {
	// FIXME(conman): Implement.
	return errors.New("not implemented")
}
