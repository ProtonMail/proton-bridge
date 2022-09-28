package bridge

import (
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/clientconfig"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
)

func (bridge *Bridge) ConfigureAppleMail(userID, address string) error {
	user, ok := bridge.users[userID]
	if !ok {
		return ErrNoSuchUser
	}

	// TODO: Handle split mode!
	if address == "" {
		address = user.Emails()[0]
	}

	// If configuring apple mail for Catalina or newer, users should use SSL.
	if useragent.IsCatalinaOrNewer() && !bridge.vault.GetSMTPSSL() {
		if err := bridge.SetSMTPSSL(true); err != nil {
			return err
		}
	}

	return (&clientconfig.AppleMail{}).Configure(
		constants.Host,
		bridge.vault.GetIMAPPort(),
		bridge.vault.GetSMTPPort(),
		bridge.vault.GetIMAPSSL(),
		bridge.vault.GetSMTPSSL(),
		address,
		strings.Join(user.Emails(), ","),
		user.BridgePass(),
	)
}
