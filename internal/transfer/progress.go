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

import (
	"crypto/sha256"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Progress maintains progress between import, export and user interface.
// Import and export update progress about processing messages and progress
// informs user interface, vice versa action (such as pause or resume) from
// user interface is passed down to import and export.
type Progress struct {
	log  *logrus.Entry
	lock sync.RWMutex

	updateCh        chan struct{}
	messageCounts   map[string]uint
	messageStatuses map[string]*MessageStatus
	pauseReason     string
	isStopped       bool
	fatalError      error
	fileReport      *fileReport
}

func newProgress(log *logrus.Entry, fileReport *fileReport) Progress {
	return Progress{
		log: log,

		updateCh:        make(chan struct{}),
		messageCounts:   map[string]uint{},
		messageStatuses: map[string]*MessageStatus{},
		fileReport:      fileReport,
	}
}

// update is helper to notify listener for updates.
func (p *Progress) update() {
	if p.updateCh == nil {
		// If the progress was ended by fatal instead finish, we ignore error.
		if p.fatalError != nil {
			return
		}
		panic("update should not be called after finish was called")
	}

	// In case no one listens for an update, do not block the progress.
	select {
	case p.updateCh <- struct{}{}:
	case <-time.After(100 * time.Millisecond):
	}
}

// start should be called before anything starts.
func (p *Progress) start() {
	p.lock.Lock()
	defer p.lock.Unlock()
}

// finish should be called as the last call once everything is done.
func (p *Progress) finish() {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.cleanUpdateCh()
}

// fatal should be called once there is error with no possible continuation.
func (p *Progress) fatal(err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.isStopped = true
	p.fatalError = err
	p.cleanUpdateCh()
}

func (p *Progress) cleanUpdateCh() {
	if p.updateCh == nil {
		// If the progress was ended by fatal instead finish, we ignore error.
		if p.fatalError != nil {
			return
		}
		panic("update should not be called after finish was called")
	}

	close(p.updateCh)
	p.updateCh = nil
}

func (p *Progress) updateCount(mailbox string, count uint) {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	log.WithField("mailbox", mailbox).WithField("count", count).Debug("Mailbox count updated")
	p.messageCounts[mailbox] = count
}

// addMessage should be called as soon as there is ID of the message.
func (p *Progress) addMessage(messageID string, rule *Rule) {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	p.log.WithField("id", messageID).Trace("Message added")
	p.messageStatuses[messageID] = &MessageStatus{
		eventTime: time.Now(),
		rule:      rule,
		SourceID:  messageID,
	}
}

// messageExported should be called right before message is exported.
func (p *Progress) messageExported(messageID string, body []byte, err error) {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	p.log.WithField("id", messageID).WithError(err).Debug("Message exported")
	status := p.messageStatuses[messageID]
	status.exportErr = err
	if err == nil {
		status.exported = true
	}

	if len(body) > 0 {
		status.bodyHash = fmt.Sprintf("%x", sha256.Sum256(body))

		if header, err := getMessageHeader(body); err != nil {
			p.log.WithField("id", messageID).WithError(err).Warning("Failed to parse headers for reporting")
		} else {
			status.setDetailsFromHeader(header)
		}
	}

	// If export failed, no other step will be done with message and we can log it to the report file.
	if err != nil {
		p.logMessage(messageID)
	}
}

// messageImported should be called right after message is imported.
func (p *Progress) messageImported(messageID, importID string, err error) {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	p.log.WithField("id", messageID).WithError(err).Debug("Message imported")
	p.messageStatuses[messageID].targetID = importID
	p.messageStatuses[messageID].importErr = err
	if err == nil {
		p.messageStatuses[messageID].imported = true
	}

	// Import is the last step, now we can log the result to the report file.
	p.logMessage(messageID)
}

// logMessage writes message status to log file.
func (p *Progress) logMessage(messageID string) {
	if p.fileReport == nil {
		return
	}
	p.fileReport.writeMessageStatus(p.messageStatuses[messageID])
}

// callWrap calls the callback and in case of problem it pause the process.
// Then it waits for user action to fix it and click on continue or abort.
func (p *Progress) callWrap(callback func() error) {
	for {
		if p.shouldStop() {
			break
		}

		err := callback()
		if err == nil {
			break
		}

		p.Pause(err.Error())
	}
}

// shouldStop is utility for providers to automatically wait during pause
// and returned value determines whether the process shouls be fully stopped.
func (p *Progress) shouldStop() bool {
	for p.IsPaused() {
		time.Sleep(time.Second)
	}
	return p.IsStopped()
}

// GetUpdateChannel returns channel notifying any update from import or export.
func (p *Progress) GetUpdateChannel() chan struct{} {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.updateCh
}

// Pause pauses the progress.
func (p *Progress) Pause(reason string) {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	p.log.Info("Progress paused")
	p.pauseReason = reason
}

// Resume resumes the progress.
func (p *Progress) Resume() {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	p.log.Info("Progress resumed")
	p.pauseReason = ""
}

// IsPaused returns whether progress is paused.
func (p *Progress) IsPaused() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.pauseReason != ""
}

// PauseReason returns pause reason.
func (p *Progress) PauseReason() string {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.pauseReason
}

// Stop stops the process.
func (p *Progress) Stop() {
	p.lock.Lock()
	defer p.update()
	defer p.lock.Unlock()

	p.log.Info("Progress stopped")
	p.isStopped = true
	p.pauseReason = "" // Clear pause to run paused code and stop it.
}

// IsStopped returns whether progress is stopped.
func (p *Progress) IsStopped() bool {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.isStopped
}

// GetFatalError returns fatal error (progress failed and did not finish).
func (p *Progress) GetFatalError() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	return p.fatalError
}

// GetFailedMessages returns statuses of failed messages.
func (p *Progress) GetFailedMessages() []*MessageStatus {
	p.lock.Lock()
	defer p.lock.Unlock()

	// Include lost messages in the process only when transfer is done.
	includeMissing := p.updateCh == nil

	statuses := []*MessageStatus{}
	for _, status := range p.messageStatuses {
		if status.hasError(includeMissing) {
			statuses = append(statuses, status)
		}
	}
	return statuses
}

// GetCounts returns counts of exported and imported messages.
func (p *Progress) GetCounts() (failed, imported, exported, added, total uint) {
	p.lock.Lock()
	defer p.lock.Unlock()

	// Include lost messages in the process only when transfer is done.
	includeMissing := p.updateCh == nil

	for _, mailboxCount := range p.messageCounts {
		total += mailboxCount
	}
	for _, status := range p.messageStatuses {
		added++
		if status.exported {
			exported++
		}
		if status.imported {
			imported++
		}
		if status.hasError(includeMissing) {
			failed++
		}
	}
	return
}

// GenerateBugReport generates similar file to import log except private information.
func (p *Progress) GenerateBugReport() []byte {
	bugReport := bugReport{}
	for _, status := range p.messageStatuses {
		bugReport.writeMessageStatus(status)
	}
	return bugReport.getData()
}
