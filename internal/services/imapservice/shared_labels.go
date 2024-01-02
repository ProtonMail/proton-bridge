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

package imapservice

import (
	"sync"

	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/usertypes"
	"golang.org/x/exp/maps"
)

type labelMap = map[string]proton.Label

// sharedLabels holds the shared state of all the available API labels which can safely be shared among
// all IMAP states. It's written this way to prevent issues with invalid use of the locks.
type sharedLabels interface {
	Read() labelsRead
	Write() labelsWrite
}

type labelsRead interface {
	Close()
	GetLabel(id string) (proton.Label, bool)
	GetLabels() []proton.Label
}

type labelsWrite interface {
	labelsRead
	SetLabel(id string, label proton.Label)
	Delete(id string)
}

type rwLabels struct {
	lock   sync.RWMutex
	labels labelMap
}

func (r *rwLabels) Read() labelsRead {
	r.lock.RLock()
	return &rwLabelsRead{rw: r}
}

func (r *rwLabels) Write() labelsWrite {
	r.lock.Lock()
	return &rwLabelsWrite{rw: r}
}

func (r *rwLabels) getLabelUnsafe(id string) (proton.Label, bool) {
	v, ok := r.labels[id]

	return v, ok
}

func (r *rwLabels) getLabelsUnsafe() []proton.Label {
	return maps.Values(r.labels)
}

func (r *rwLabels) SetLabels(labels []proton.Label) {
	r.lock.Lock()
	defer r.lock.Unlock()

	r.labels = usertypes.GroupBy(labels, func(label proton.Label) string { return label.ID })
}

func (r *rwLabels) GetLabelMap() labelMap {
	r.lock.Lock()
	defer r.lock.Unlock()

	return maps.Clone(r.labels)
}

func newRWLabels() *rwLabels {
	return &rwLabels{
		labels: make(labelMap),
	}
}

type rwLabelsRead struct {
	rw *rwLabels
}

func (r rwLabelsRead) Close() {
	r.rw.lock.RUnlock()
}

func (r rwLabelsRead) GetLabel(id string) (proton.Label, bool) {
	return r.rw.getLabelUnsafe(id)
}

func (r rwLabelsRead) GetLabels() []proton.Label {
	return r.rw.getLabelsUnsafe()
}

type rwLabelsWrite struct {
	rw *rwLabels
}

func (r rwLabelsWrite) Close() {
	r.rw.lock.Unlock()
}

func (r rwLabelsWrite) GetLabel(id string) (proton.Label, bool) {
	return r.rw.getLabelUnsafe(id)
}

func (r rwLabelsWrite) GetLabels() []proton.Label {
	return r.rw.getLabelsUnsafe()
}

func (r rwLabelsWrite) SetLabel(id string, label proton.Label) {
	r.rw.labels[id] = label
}

func (r rwLabelsWrite) Delete(id string) {
	delete(r.rw.labels, id)
}
