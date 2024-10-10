//go:build linux

package mpris

// MediaPlayer2 dbus interface implementation

func (mh *MprisHandler) Raise() error {
	return nil
}

func (mh *MprisHandler) Quit() error {
	return nil
}

func (mh *MprisHandler) CanQuit() (bool, error) {
	return false, nil
}

func (mh *MprisHandler) CanRaise() (bool, error) {
	return false, nil
}

func (mh *MprisHandler) HasTrackList() (bool, error) {
	return false, nil
}

func (mh *MprisHandler) Identity() (string, error) {
	return mh.description, nil
}

func (mh *MprisHandler) SupportedUriSchemes() ([]string, error) {
	return []string{}, nil
}

func (mh *MprisHandler) SupportedMimeTypes() ([]string, error) {
	return []string{}, nil
}
