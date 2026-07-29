package api

// Platform-neutral wrappers over the OS file/folder dialogs implemented in
// dialogs_windows.go / dialogs_other.go.

func (a *API) PickFolder(title string, initialDir string) string {
	path, err := runFolderDialog(title, initialDir)
	if err != nil {
		return ""
	}
	return path
}

func (a *API) PickOpenFile(title string, initialDir string, filter string) string {
	path, err := runFileDialog(title, initialDir, filter, false)
	if err != nil {
		return ""
	}
	return path
}

func (a *API) PickSaveFile(title string, initialDir string, filter string, defaultName string) string {
	path, err := runSaveDialog(title, initialDir, filter, defaultName)
	if err != nil {
		return ""
	}
	return path
}
