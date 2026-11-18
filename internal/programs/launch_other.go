//go:build !windows

package programs

func startDetached(string) error { return errNotSupported }

func openInExplorer(string) error { return errNotSupported }
