//go:build windows

package workspace

import (
	"golang.org/x/sys/windows"
)

// manifestLockType returns the LockFileEx flags for shared or exclusive non-blocking lock.
func manifestLockType(exclusive bool) int {
	flags := windows.LOCKFILE_FAIL_IMMEDIATELY
	if exclusive {
		flags |= windows.LOCKFILE_EXCLUSIVE_LOCK
	}
	return int(flags)
}

// manifestFlock acquires a lock on the file descriptor using LockFileEx.
func manifestFlock(fd uintptr, lockType int) error {
	ol := new(windows.Overlapped)
	return windows.LockFileEx(
		windows.Handle(fd),
		uint32(lockType),
		0, // reserved
		1, // lock 1 byte
		0, // high bits of length
		ol,
	)
}

// manifestUnlock releases the lock on the file descriptor using UnlockFileEx.
func manifestUnlock(fd uintptr) error {
	ol := new(windows.Overlapped)
	return windows.UnlockFileEx(
		windows.Handle(fd),
		0, // reserved
		1, // unlock 1 byte
		0, // high bits of length
		ol,
	)
}

// isLockBusy returns true if the error indicates the lock is held by another process.
func isLockBusy(err error) bool {
	// ERROR_LOCK_VIOLATION (33) is returned when the lock is held
	return err == windows.ERROR_LOCK_VIOLATION
}
