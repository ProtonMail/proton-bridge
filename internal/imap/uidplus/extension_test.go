// Copyright (c) 2022 Proton AG
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
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package uidplus

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// uidValidity is constant and global for bridge IMAP.
const uidValidity = 66

type testResponseData struct {
	sourceList, targetList     []int
	expCopyInfo, expAppendInfo string
}

func (td *testResponseData) getOrdSeqFromList(seqList []int) *OrderedSeq {
	set := &OrderedSeq{}
	for _, seq := range seqList {
		set.Add(uint32(seq))
	}
	return set
}

func (td *testResponseData) testCopyAndAppendResponses(tb testing.TB) {
	sourceSeq := td.getOrdSeqFromList(td.sourceList)
	targetSeq := td.getOrdSeqFromList(td.targetList)

	gotCopyResp := getStatusResponseCopy(uidValidity, sourceSeq, targetSeq)
	assert.Equal(tb, td.expCopyInfo, gotCopyResp.Info, "source: %v\ntarget: %v", td.sourceList, td.targetList)

	gotAppendResp := getStatusResponseAppend(uidValidity, targetSeq)
	assert.Equal(tb, td.expAppendInfo, gotAppendResp.Info, "source: %v\ntarget: %v", td.sourceList, td.targetList)
}

func TestStatusResponseInfo(t *testing.T) {
	testData := []*testResponseData{
		{ // Dynamic range must never be returned  e.g 4:*  (explicitly true if you OrderedSeq used instead of imap.SeqSet).
			sourceList:    []int{4, 5, 6},
			targetList:    []int{1, 2, 3},
			expCopyInfo:   "[" + copyuid + " 66 4:6 1:3] " + copySuccess,
			expAppendInfo: "[" + appenduid + " 66 1:3] " + appendSucess,
		},
		{ // Ranges can be used only for consecutive strictly rising sequence.
			sourceList:    []int{6, 7, 8, 9, 10, 1, 3, 5, 10, 11, 20, 21, 30, 31},
			targetList:    []int{1, 2, 3, 4, 50, 8, 7, 6, 12, 13, 22, 23, 32, 33},
			expCopyInfo:   "[" + copyuid + " 66 6:10,1,3,5,10:11,20:21,30:31 1:4,50,8,7,6,12:13,22:23,32:33] " + copySuccess,
			expAppendInfo: "[" + appenduid + " 66 1:4,50,8,7,6,12:13,22:23,32:33] " + appendSucess,
		},
		{ // Keep order (cannot use sequence set because 3,2,1 equals 1,2,3 equals 1:3 equals 3:1).
			sourceList:    []int{4, 5, 8},
			targetList:    []int{3, 2, 1},
			expCopyInfo:   "[" + copyuid + " 66 4:5,8 3,2,1] " + copySuccess,
			expAppendInfo: "[" + appenduid + " 66 3,2,1] " + appendSucess,
		},
		{ // Incorrect count of source and target uids is wrong and we should not report it.
			sourceList:    []int{1},
			targetList:    []int{1, 2, 3},
			expCopyInfo:   copySuccess,
			expAppendInfo: "[" + appenduid + " 66 1:3] " + appendSucess,
		},
		{
			sourceList:    []int{1, 2, 3},
			targetList:    []int{1},
			expCopyInfo:   copySuccess,
			expAppendInfo: "[" + appenduid + " 66 1] " + appendSucess,
		},
		{ // One item should be always interpreted as one number (don't use imap.SeqSet because 1:1 means 1).
			sourceList:    []int{1},
			targetList:    []int{1},
			expCopyInfo:   "[" + copyuid + " 66 1 1] " + copySuccess,
			expAppendInfo: "[" + appenduid + " 66 1] " + appendSucess,
		},
		{ // No UID is wrong we should not report it.
			sourceList:    []int{1},
			targetList:    []int{},
			expCopyInfo:   copySuccess,
			expAppendInfo: appendSucess,
		},
		{ // Duplicates should be reported as list.
			sourceList:    []int{1, 1, 1},
			targetList:    []int{6, 6, 6},
			expCopyInfo:   "[" + copyuid + " 66 1,1,1 6,6,6] " + copySuccess,
			expAppendInfo: "[" + appenduid + " 66 6,6,6] " + appendSucess,
		},
	}

	for _, td := range testData {
		td.testCopyAndAppendResponses(t)
	}
}
