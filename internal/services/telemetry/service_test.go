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

package telemetry

import (
	"context"
	"errors"
	"testing"

	"github.com/ProtonMail/gluon/async"
	"github.com/ProtonMail/go-proton-api"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/orderedtasks"
	"github.com/ProtonMail/proton-bridge/v3/internal/services/userevents"
	"github.com/stretchr/testify/require"
)

const TestUserID = "MyUserID"

type mockSettingsGetter struct {
	settings proton.UserSettings
	err      error
}

func (m *mockSettingsGetter) GetUserSettings(_ context.Context) (proton.UserSettings, error) {
	return m.settings, m.err
}

func TestService_Unitialised(t *testing.T) {
	mockSettingsGetter := &mockSettingsGetter{
		settings: proton.UserSettings{Telemetry: proton.SettingEnabled},
		err:      errors.New("cannot get user settings"),
	}

	service := NewService(
		TestUserID,
		mockSettingsGetter,
		&userevents.NoOpSubscribable{},
	)

	require.False(t, service.isInitialised)
	require.False(t, service.isTelemetryEnabled)

	ctx := context.Background()
	group := orderedtasks.NewOrderedCancelGroup(async.NoopPanicHandler{})
	defer group.CancelAndWait()

	service.Start(ctx, group)
	require.False(t, service.IsTelemetryEnabled(ctx))
	require.False(t, service.isInitialised)
	require.False(t, service.isTelemetryEnabled)

	mockSettingsGetter.err = nil
	require.True(t, service.IsTelemetryEnabled(ctx))
	require.True(t, service.isInitialised)
	require.True(t, service.isTelemetryEnabled)
}

func TestService_OnUserSettingsEvent(t *testing.T) {
	mockSettingsGetter := &mockSettingsGetter{}

	service := NewService(
		TestUserID,
		mockSettingsGetter,
		&userevents.NoOpSubscribable{},
	)
	require.False(t, service.isInitialised)

	ctx := context.Background()
	group := orderedtasks.NewOrderedCancelGroup(async.NoopPanicHandler{})
	defer group.CancelAndWait()

	service.Start(ctx, group)
	require.True(t, service.isInitialised)
	require.False(t, service.IsTelemetryEnabled(ctx))

	require.NoError(t, service.HandleUserSettingsEvent(context.Background(), &proton.UserSettings{Telemetry: proton.SettingEnabled}))
	require.True(t, service.IsTelemetryEnabled(ctx))

	require.NoError(t, service.HandleUserSettingsEvent(context.Background(), &proton.UserSettings{Telemetry: proton.SettingDisabled}))
	require.False(t, service.IsTelemetryEnabled(ctx))
}

func TestService_Unitialised_OnUserSettingsEvent(t *testing.T) {
	mockSettingsGetter := &mockSettingsGetter{
		settings: proton.UserSettings{Telemetry: proton.SettingEnabled},
		err:      errors.New("cannot get user settings"),
	}

	service := NewService(
		TestUserID,
		mockSettingsGetter,
		&userevents.NoOpSubscribable{},
	)

	require.False(t, service.isInitialised)
	require.False(t, service.isTelemetryEnabled)

	ctx := context.Background()
	group := orderedtasks.NewOrderedCancelGroup(async.NoopPanicHandler{})
	defer group.CancelAndWait()

	service.Start(ctx, group)
	require.False(t, service.IsTelemetryEnabled(ctx))
	require.False(t, service.isInitialised)
	require.False(t, service.isTelemetryEnabled)

	require.NoError(t, service.HandleUserSettingsEvent(context.Background(), &proton.UserSettings{Telemetry: proton.SettingEnabled}))
	require.True(t, service.IsTelemetryEnabled(ctx))
	require.True(t, service.isInitialised)
	require.True(t, service.isTelemetryEnabled)
}
