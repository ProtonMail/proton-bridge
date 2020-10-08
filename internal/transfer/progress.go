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
type Progress struct { //nolint[maligned]
	log  *logrus.Entry
	lock sync.Locker

	updateCh        chan struct{}
	messageCounted  bool
	messageCounts   map[string]uint
	messageStatuses map[string]*MessageStatus
	pauseReason     string
	isStopped       bool
	fatalError      error
	fileReport      *fileReport
}

func newProgress(log *logrus.Entry, fileReport *fileReport) Progress {
	return Progress{
		log:  log,
		lock: &sync.Mutex{},

		updateCh:        make(chan struct{}),
		messageCounts:   map[string]uint{},
		messageStatuses: map[string]*MessageStatus{},
		fileReport:      fileReport,
	}
}

// update is helper to notify listener for updates.
func (p *Progress) update() {
	if p.updateCh == nil {
		return
	}

	// In case no one listens for an update, do not block the whole progress.
	go func() {
		defer func() {
			// updateCh can be closed at the end of progress which is fine.
			if r := recover(); r != nil {
				log.WithField("r", r).Warn("Failed to send update")
			}
		}()

		select {
		case p.updateCh <- struct{}{}:
		case <-time.After(5 * time.Millisecond):
		}
	}()
}

// finish should be called as the last call once everything is done.
func (p *Progress) finish() {
	p.lock.Lock()
	defer p.lock.Unlock()

	log.Debug("Progress finished")
	p.cleanUpdateCh()
}

// fatal should be called once there is error with no possible continuation.
func (p *Progress) fatal(err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	log.WithError(err).Error("Progress finished")
	p.setStop()
	p.fatalError = err
	p.cleanUpdateCh()
}

func (p *Progress) cleanUpdateCh() {
	if p.updateCh == nil {
		return
	}

	close(p.updateCh)
	p.updateCh = nil
}

func (p *Progress) countsFinal() {
	p.lock.Lock()
	defer p.lock.Unlock()
	defer p.update()

	log.Info("Estimating count finished")
	p.messageCounted = true
}

func (p *Progress) updateCount(mailbox string, count uint) {
	p.lock.Lock()
	defer p.lock.Unlock()
	defer p.update()

	log.WithField("mailbox", mailbox).WithField("count", count).Debug("Mailbox count updated")
	p.messageCounts[mailbox] = count
}

// addMessage should be called as soon as there is ID of the message.
func (p *Progress) addMessage(messageID string, sourceNames, targetNames []string) {
	p.lock.Lock()
	defer p.lock.Unlock()
	defer p.update()

	p.log.WithField("id", messageID).Trace("Message added")
	p.messageStatuses[messageID] = &MessageStatus{
		eventTime:   time.Now(),
		sourceNames: sourceNames,
		SourceID:    messageID,
		targetNames: targetNames,
	}
}

// messageExported should be called right before message is exported.
func (p *Progress) messageExported(messageID string, body []byte, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()
	defer p.update()

	log := p.log.WithField("id", messageID)
	if err != nil {
		log = log.WithError(err)
	}
	log.Debug("Message exported")

	status := p.messageStatuses[messageID]
	status.exportErr = err
	if err == nil {
		status.exported = true
	}

	if len(body) > 0 {
		status.bodyHash = fmt.Sprintf("%x", sha256.Sum256(body))

		if header, err := getMessageHeader(body); err != nil {
			log.WithError(err).Warning("Failed to parse headers for reporting")
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
	defer p.lock.Unlock()
	defer p.update()

	log := p.log.WithField("id", messageID)
	if err != nil {
		log = log.WithError(err)
	}
	log.Debug("Message imported")

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
// Every function doing I/O should be wrapped by this function to provide
// stopping and pausing functionality.
func (p *Progress) callWrap(callback func() error) {
	for {
		if p.shouldStop() {
			break
		}

		err := callback()
		if err == nil {
			break
		}

		p.Pause("paused due to " + err.Error())
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
	defer p.lock.Unlock()
	defer p.update()

	p.log.Info("Progress paused")
	p.pauseReason = reason
}

// Resume resumes the progress.
func (p *Progress) Resume() {
	p.lock.Lock()
	defer p.lock.Unlock()
	defer p.update()

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
	defer p.lock.Unlock()
	defer p.update()

	p.log.Info("Progress stopped")
	p.setStop()

	// Once progress is stopped, some calls might be in progress. Results from
	// those calls are irrelevant so we can close update channel sooner to not
	// propagate any progress to user interface anymore.
	p.cleanUpdateCh()
}

func (p *Progress) setStop() {
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

	// Return counts only once total is estimated or the process already
	// ended (for a case when it ended quickly to report it correctly).
	if p.updateCh != nil && !p.messageCounted {
		return
	}

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

// FileReport returns path to generated defailed file report.
func (p *Progress) FileReport() string {
	if p.fileReport == nil {
		return ""
	}
	return p.fileReport.path
}
