package grpc

/*
func (s *Service) checkUpdate() {
	version, err := s.updater.Check()
	if err != nil {
		s.log.WithError(err).Error("An error occurred while checking for updates")
		s.SetVersion(updater.VersionInfo{})
		return
	}
	s.SetVersion(version)
}

func (s *Service) updateForce() {
	s.updateCheckMutex.Lock()
	defer s.updateCheckMutex.Unlock()
	s.checkUpdate()
	_ = s.SendEvent(NewUpdateForceEvent(s.newVersionInfo.Version.String()))
}

func (s *Service) checkUpdateAndNotify(isReqFromUser bool) {
	s.updateCheckMutex.Lock()
	defer func() {
		s.updateCheckMutex.Unlock()
		_ = s.SendEvent(NewUpdateCheckFinishedEvent())
	}()

	s.checkUpdate()
	version := s.newVersionInfo
	if version.Version.String() == "" {
		if isReqFromUser {
			_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
		}
		return
	}
	if !s.updater.IsUpdateApplicable(s.newVersionInfo) {
		s.log.Info("No need to update")
		if isReqFromUser {
			_ = s.SendEvent(NewUpdateIsLatestVersionEvent())
		}
	} else if isReqFromUser {
		s.NotifyManualUpdate(s.newVersionInfo, s.updater.CanInstall(s.newVersionInfo))
	}
}

func (s *Service) installUpdate() {
	s.updateCheckMutex.Lock()
	defer s.updateCheckMutex.Unlock()

	if !s.updater.CanInstall(s.newVersionInfo) {
		s.log.Warning("Skipping update installation, current version too old")
		_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
		return
	}

	if err := s.updater.InstallUpdate(s.newVersionInfo); err != nil {
		if errors.Cause(err) == updater.ErrDownloadVerify {
			s.log.WithError(err).Warning("Skipping update installation due to temporary error")
		} else {
			s.log.WithError(err).Error("The update couldn't be installed")
			_ = s.SendEvent(NewUpdateErrorEvent(UpdateErrorType_UPDATE_MANUAL_ERROR))
		}
		return
	}

	_ = s.SendEvent(NewUpdateSilentRestartNeededEvent())
}
*/
