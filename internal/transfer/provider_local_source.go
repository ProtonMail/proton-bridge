// Copyright (c) 2021 Proton Technologies AG
//
// This file is part of ProtonMail Bridge.
//
// ProtonMail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// ProtonMail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with ProtonMail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package transfer

import (
	"sync"
)

// TransferTo exports messages based on rules to channel.
func (p *LocalProvider) TransferTo(rules transferRules, progress *Progress, ch chan<- Message) {
	log.Info("Started transfer from EML and MBOX to channel")
	defer log.Info("Finished transfer from EML and MBOX to channel")

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		p.emlProvider.TransferTo(rules, progress, ch)
	}()
	go func() {
		defer wg.Done()
		p.mboxProvider.TransferTo(rules, progress, ch)
	}()

	wg.Wait()
}
