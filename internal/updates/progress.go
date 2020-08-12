// Copyright (c) 2020 Proton Technologies AG
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

package updates

const (
	InfoCurrentVersion = 1 + iota
	InfoDownloading
	InfoVerifying
	InfoUnpacking
	InfoUpgrading
	InfoQuitApp
	InfoRestartApp
)

type Progress struct {
	Processed   float32 // fraction of finished procedure [0.0-1.0]
	Description int     // description by code (needs to be translated anyway)
	Err         error   // occurred error
	channel     chan<- Progress
}

func (s *Progress) Update() {
	s.channel <- *s
}

func (s *Progress) UpdateDescription(description int) {
	s.Description = description
	s.Processed = 0
	s.Update()
}

func (s *Progress) UpdateProcessed(processed float32) {
	s.Processed = processed
	s.Update()
}
