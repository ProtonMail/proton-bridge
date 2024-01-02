// Copyright (c) 2024 Proton AG
//
// This file is part of Proton Mail Bridge.
//
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge.  If not, see <https://www.gnu.org/licenses/>.

package syncservice

import (
	"os"

	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
)

const Kilobyte = uint64(1024)
const Megabyte = 1024 * Kilobyte
const Gigabyte = 1024 * Megabyte

func toMB(v uint64) float64 {
	return float64(v) / float64(Megabyte)
}

type syncLimits struct {
	MaxDownloadRequestMem uint64
	MinDownloadRequestMem uint64
	MaxMessageBuildingMem uint64
	MinMessageBuildingMem uint64
	MaxSyncMemory         uint64
	MaxParallelDownloads  int
	DownloadRequestMem    uint64
	MessageBuildMem       uint64
}

func newSyncLimits(maxSyncMemory uint64) syncLimits {
	limits := syncLimits{
		// There's no point in using more than 128MB of download data per stage, after that we reach a point of diminishing
		// returns as we can't keep the pipeline fed fast enough.
		MaxDownloadRequestMem: 128 * Megabyte,

		// Any lower than this and we may fail to download messages.
		MinDownloadRequestMem: 40 * Megabyte,

		// This value can be increased to your hearts content. The more system memory the user has, the more messages
		// we can build in parallel.
		MaxMessageBuildingMem: 128 * Megabyte,
		MinMessageBuildingMem: 64 * Megabyte,

		// Maximum recommend value for parallel downloads by the API team.
		MaxParallelDownloads: 32,

		MaxSyncMemory: maxSyncMemory,
	}

	if _, ok := os.LookupEnv("BRIDGE_SYNC_FORCE_MINIMUM_SPEC"); ok {
		logrus.Warn("Sync specs forced to minimum")
		limits.MaxDownloadRequestMem = 50 * Megabyte
		limits.MaxMessageBuildingMem = 80 * Megabyte
		limits.MaxParallelDownloads = 2
		limits.MaxSyncMemory = 800 * Megabyte
	}

	// Expected mem usage for this whole process should be the sum of MaxMessageBuildingMem and MaxDownloadRequestMem
	// times x due to pipeline and all additional memory used by network requests and compression+io.

	totalMemory := memory.TotalMemory()

	if limits.MaxSyncMemory >= totalMemory/2 {
		logrus.Warnf("Requested max sync memory of %v MB is greater than half of system memory (%v MB), forcing to half of system memory",
			toMB(limits.MaxSyncMemory), toMB(totalMemory/2))
		limits.MaxSyncMemory = totalMemory / 2
	}

	if limits.MaxSyncMemory < 800*Megabyte {
		logrus.Warnf("Requested max sync memory of %v MB, but minimum recommended is 800 MB, forcing max syncMemory to 800MB", toMB(limits.MaxSyncMemory))
		limits.MaxSyncMemory = 800 * Megabyte
	}

	logrus.Debugf("Total System Memory: %v", toMB(totalMemory))

	// If less than 2GB available try and limit max memory to 512 MB
	switch {
	case limits.MaxSyncMemory < 2*Gigabyte:
		if limits.MaxSyncMemory < 800*Megabyte {
			logrus.Warnf("System has less than 800MB of memory, you may experience issues sycing large mailboxes")
		}
		limits.DownloadRequestMem = limits.MinDownloadRequestMem
		limits.MessageBuildMem = limits.MinMessageBuildingMem
	case limits.MaxSyncMemory == 2*Gigabyte:
		// Increasing the max download capacity has very little effect on sync speed. We could increase the download
		// memory but the user would see less sync notifications. A smaller value here leads to more frequent
		// updates. Additionally, most of sync time is spent in the message building.
		limits.DownloadRequestMem = limits.MaxDownloadRequestMem
		// Currently limited so that if a user has multiple accounts active it also doesn't cause excessive memory usage.
		limits.MessageBuildMem = limits.MaxMessageBuildingMem
	default:
		// Divide by 8 as download stage and build stage will use aprox. 4x the specified memory.
		remainingMemory := (limits.MaxSyncMemory - 2*Gigabyte) / 8
		limits.DownloadRequestMem = limits.MaxDownloadRequestMem + remainingMemory
		limits.MessageBuildMem = limits.MaxMessageBuildingMem + remainingMemory
	}

	logrus.Debugf("Max memory usage for sync Download=%vMB Building=%vMB Predicted Max Total=%vMB",
		toMB(limits.DownloadRequestMem),
		toMB(limits.MessageBuildMem),
		toMB((limits.MessageBuildMem*4)+(limits.DownloadRequestMem*4)),
	)

	return limits
}
