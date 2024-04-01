// Copyright (c) 2024 Proton AG
// This file is part of Proton Mail Bridge.
// Proton Mail Bridge is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
// Proton Mail Bridge is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
// You should have received a copy of the GNU General Public License
// along with Proton Mail Bridge. If not, see <https://www.gnu.org/licenses/>.

package hv

import (
	"errors"
	"fmt"
	"strings"

	"github.com/ProtonMail/go-proton-api"
)

// VerifyAndExtractHvRequest expects an error request as input
// determines whether the given error is a Proton human verification request; if it isn't then it returns -> nil, nil (no details, no error)
// if it is a HV req. then it tries to parse the json data and verify that the captcha method is included; if either fails -> nil, err
// if the HV request was successfully decoded and the preconditions were met it returns the hv details -> hvDetails, nil.
func VerifyAndExtractHvRequest(err error) (*proton.APIHVDetails, error) {
	if err == nil {
		return nil, nil
	}

	var protonErr *proton.APIError
	if errors.As(err, &protonErr) && protonErr.IsHVError() {
		hvDetails, hvErr := protonErr.GetHVDetails()
		if hvErr != nil {
			return nil, fmt.Errorf("received HV request, but can't decode HV details")
		}
		return hvDetails, nil
	}

	return nil, nil
}

func IsHvRequest(err error) bool {
	if err == nil {
		return false
	}

	var protonErr *proton.APIError
	if errors.As(err, &protonErr) && protonErr.IsHVError() {
		return true
	}

	return false
}

func FormatHvURL(details *proton.APIHVDetails) string {
	return fmt.Sprintf("https://verify.proton.me/?methods=%v&token=%v",
		strings.Join(details.Methods, ","),
		details.Token)
}
