//go:build windows

package api

import (
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// Native Win32 file/folder dialogs via syscall. This replaces the old
// approach of spawning powershell.exe for every dialog, which was slow
// (~300ms+ per process launch) and fragile (string-injection risk).

var (
	shell32  = syscall.NewLazyDLL("shell32.dll")
	comdlg32 = syscall.NewLazyDLL("comdlg32.dll")
	user32   = syscall.NewLazyDLL("user32.dll")
	ole32    = syscall.NewLazyDLL("ole32.dll")

	procSHBrowseForFolderW   = shell32.NewProc("SHBrowseForFolderW")
	procSHGetPathFromIDListW = shell32.NewProc("SHGetPathFromIDListW")
	procCoTaskMemFree        = ole32.NewProc("CoTaskMemFree")
	procCoInitializeEx       = ole32.NewProc("CoInitializeEx")
	procCoUninitialize       = ole32.NewProc("CoUninitialize")
	procGetOpenFileNameW     = comdlg32.NewProc("GetOpenFileNameW")
	procGetSaveFileNameW     = comdlg32.NewProc("GetSaveFileNameW")
	procSendMessageW         = user32.NewProc("SendMessageW")
)

const (
	ofnFileMustExist    = 0x00001000
	ofnPathMustExist    = 0x00000800
	ofnOverwritePrompt  = 0x00000002
	ofnHideReadOnly     = 0x00000004
	ofnNoChangeDir      = 0x00000008

	bifReturnOnlyFSDirs = 0x00000001
	bifNewDialogStyle   = 0x00000040

	bffmInitialized   = 1
	bffmSetSelectionW = 0x04CB // WM_USER (0x0400) + 203
	coinitApartment   = 0x2

	// maxPathBuf is large enough for long paths (> MAX_PATH) since the
	// common dialogs copy the result into this buffer.
	maxPathBuf = 32768
)

// openFileNameW mirrors the Windows OPENFILENAMEW structure. Field order
// and types match the SDK definition so unsafe.Sizeof yields the value
// the API expects (116 bytes on 32-bit, 152 on 64-bit).
type openFileNameW struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
	pvReserved        uintptr
	dwReserved        uint32
	flagsEx           uint32
}

// browseInfoW mirrors the Windows BROWSEINFOW structure.
type browseInfoW struct {
	hwndOwner      uintptr
	pidlRoot       uintptr
	pszDisplayName *uint16
	lpszTitle      *uint16
	ulFlags        uint32
	lpfn           uintptr
	lParam         uintptr
	iImage         int32
}

// makeFilter converts a .NET-style filter ("All (*.*)|*.*") into the
// double-NUL-terminated UTF-16 buffer that the common dialogs expect.
func makeFilter(filter string) []uint16 {
	if filter == "" {
		filter = "All files (*.*)|*.*"
	}
	s := strings.ReplaceAll(filter, "|", "\x00") + "\x00\x00"
	return utf16.Encode([]rune(s))
}

// browseCallbackProc is invoked by SHBrowseForFolderW. On
// BFFM_INITIALIZED it sends BFFM_SETSELECTIONW so the dialog opens at
// lpData (the initial directory pointer we passed via BROWSEINFOW.lParam).
func browseCallbackProc(hwnd uintptr, uMsg uint32, lParam uintptr, lpData uintptr) int {
	if uMsg == bffmInitialized && lpData != 0 {
		procSendMessageW.Call(hwnd, bffmSetSelectionW, 1, lpData)
	}
	return 0
}

func runFolderDialog(title, initialDir string) (string, error) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	displayName := make([]uint16, maxPathBuf)

	var initDirPtr *uint16
	if initialDir != "" {
		initDirPtr, _ = syscall.UTF16PtrFromString(initialDir)
	}

	bi := browseInfoW{
		pszDisplayName: &displayName[0],
		lpszTitle:      titlePtr,
		ulFlags:        bifReturnOnlyFSDirs | bifNewDialogStyle,
		lpfn:           syscall.NewCallback(browseCallbackProc),
		lParam:         uintptr(unsafe.Pointer(initDirPtr)),
	}

	// The new-style browse dialog requires COM to be initialised.
	procCoInitializeEx.Call(0, coinitApartment)
	defer procCoUninitialize.Call()

	pidl, _, _ := procSHBrowseForFolderW.Call(uintptr(unsafe.Pointer(&bi)))
	if pidl == 0 {
		return "", nil // user cancelled
	}
	defer procCoTaskMemFree.Call(pidl)

	path := make([]uint16, maxPathBuf)
	ok, _, _ := procSHGetPathFromIDListW.Call(pidl, uintptr(unsafe.Pointer(&path[0])))
	if ok == 0 {
		return "", nil
	}
	return syscall.UTF16ToString(path), nil
}

func runFileDialog(title, initialDir, filter string, save bool) (string, error) {
	filterW := makeFilter(filter)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	var initDirPtr *uint16
	if initialDir != "" {
		initDirPtr, _ = syscall.UTF16PtrFromString(initialDir)
	}

	fileBuf := make([]uint16, maxPathBuf)

	ofn := openFileNameW{
		lStructSize:     uint32(unsafe.Sizeof(openFileNameW{})),
		lpstrFilter:     &filterW[0],
		nFilterIndex:    1,
		lpstrFile:       &fileBuf[0],
		nMaxFile:        maxPathBuf,
		lpstrInitialDir: initDirPtr,
		lpstrTitle:      titlePtr,
	}

	var proc *syscall.LazyProc
	if save {
		ofn.flags = ofnOverwritePrompt | ofnPathMustExist | ofnHideReadOnly | ofnNoChangeDir
		proc = procGetSaveFileNameW
	} else {
		ofn.flags = ofnFileMustExist | ofnPathMustExist | ofnHideReadOnly | ofnNoChangeDir
		proc = procGetOpenFileNameW
	}

	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", nil // user cancelled or error
	}
	return syscall.UTF16ToString(fileBuf), nil
}

func runSaveDialog(title, initialDir, filter, defaultName string) (string, error) {
	filterW := makeFilter(filter)
	titlePtr, _ := syscall.UTF16PtrFromString(title)

	var initDirPtr *uint16
	if initialDir != "" {
		initDirPtr, _ = syscall.UTF16PtrFromString(initialDir)
	}

	fileBuf := make([]uint16, maxPathBuf)
	if defaultName != "" {
		copy(fileBuf, utf16.Encode([]rune(defaultName)))
	}

	ofn := openFileNameW{
		lStructSize:     uint32(unsafe.Sizeof(openFileNameW{})),
		lpstrFilter:     &filterW[0],
		nFilterIndex:    1,
		lpstrFile:       &fileBuf[0],
		nMaxFile:        maxPathBuf,
		lpstrInitialDir: initDirPtr,
		lpstrTitle:      titlePtr,
		flags:           ofnOverwritePrompt | ofnPathMustExist | ofnHideReadOnly | ofnNoChangeDir,
	}

	r, _, _ := procGetSaveFileNameW.Call(uintptr(unsafe.Pointer(&ofn)))
	if r == 0 {
		return "", nil // user cancelled or error
	}
	return syscall.UTF16ToString(fileBuf), nil
}
