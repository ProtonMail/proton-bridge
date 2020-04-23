package pmapi

func newTestClientManager(cfg *ClientConfig) *ClientManager {
	cm := NewClientManager(cfg)

	go func() {
		for range cm.authUpdates {
		}
	}()

	return cm
}
