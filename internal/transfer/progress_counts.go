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

package transfer

// ProgressCounts holds counts counted by Progress.
type ProgressCounts struct {
	Failed,
	Skipped,
	Imported,
	Exported,
	Added,
	Total uint
}

// Progress returns ratio between processed messages (fully imported, skipped
// and failed ones) and total number of messages as percentage (0 - 1).
func (c *ProgressCounts) Progress() float32 {
	progressed := c.Imported + c.Skipped + c.Failed
	return float32(progressed) / float32(c.Total)
}
