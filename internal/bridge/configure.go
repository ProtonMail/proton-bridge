package bridge

import (
	"fmt"
	"strings"

	"github.com/ProtonMail/proton-bridge/v2/internal/clientconfig"
	"github.com/ProtonMail/proton-bridge/v2/internal/constants"
	"github.com/ProtonMail/proton-bridge/v2/internal/user"
	"github.com/ProtonMail/proton-bridge/v2/internal/useragent"
	"github.com/ProtonMail/proton-bridge/v2/internal/vault"
)

func (bridge *Bridge) ConfigureAppleMail(userID, address string) error {
	if ok, err := bridge.users.GetErr(userID, func(user *user.User) error {
		if address == "" {
			address = user.Emails()[0]
		}

		username := address
		addresses := address

		if user.GetAddressMode() == vault.CombinedMode {
			username = user.Emails()[0]
			addresses = strings.Join(user.Emails(), ",")
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
			username,
			addresses,
			user.BridgePass(),
		)
	}); !ok {
		return ErrNoSuchUser
	} else if err != nil {
		return fmt.Errorf("failed to configure apple mail: %w", err)
	}

	return nil
}
