package pmapi

import "context"

func (m *manager) testPing(ctx context.Context) error {
	if _, err := m.r(ctx).Get("/tests/ping"); err != nil {
		return err
	}

	return nil
}
