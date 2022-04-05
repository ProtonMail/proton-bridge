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

package pmapi

import (
	"context"
)

type pmapiContextKey string

const (
	retryContextKey = pmapiContextKey("retry")
	retryDisabled   = "disabled"

	authRefreshContextKey = pmapiContextKey("authRefresh")
	authRefreshDisabled   = "disabled"
)

func ContextWithoutRetry(parent context.Context) context.Context {
	return context.WithValue(parent, retryContextKey, retryDisabled)
}

func isRetryDisabled(ctx context.Context) bool {
	if v := ctx.Value(retryContextKey); v != nil {
		return v == retryDisabled
	}
	return false
}

func ContextWithoutAuthRefresh(parent context.Context) context.Context {
	return context.WithValue(parent, authRefreshContextKey, authRefreshDisabled)
}

func isAuthRefreshDisabled(ctx context.Context) bool {
	if v := ctx.Value(authRefreshContextKey); v != nil {
		return v == authRefreshDisabled
	}
	return false
}
