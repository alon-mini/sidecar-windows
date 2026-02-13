//go:build windows

package config

import "golang.org/x/sys/windows"

// fileLock acquires an exclusive lock on the file descriptor using LockFileEx.
func fileLock(fd uintptr) error {
	ol := new(windows.Overlapped)
	return windows.LockFileEx(
		windows.Handle(fd),
		windows.LOCKFILE_EXCLUSIVE_LOCK,
		0, // reserved
		1, // lock 1 byte
		0, // high bits of length
		ol,
	)
}

// fileUnlock releases the lock on the file descriptor using UnlockFileEx.
func fileUnlock(fd uintptr) error {
	ol := new(windows.Overlapped)
	return windows.UnlockFileEx(
		windows.Handle(fd),
		0, // reserved
		1, // unlock 1 byte
		0, // high bits of length
		ol,
	)
}
