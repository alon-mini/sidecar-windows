//go:build unix

package workspace

import "syscall"

// manifestLockType returns the flock flag for shared or exclusive non-blocking lock.
func manifestLockType(exclusive bool) int {
	if exclusive {
		return syscall.LOCK_EX | syscall.LOCK_NB
	}
	return syscall.LOCK_SH | syscall.LOCK_NB
}

// manifestFlock applies the given flock operation on the file descriptor.
func manifestFlock(fd uintptr, lockType int) error {
	return syscall.Flock(int(fd), lockType)
}

// manifestUnlock releases the flock on the file descriptor.
func manifestUnlock(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_UN)
}

// isLockBusy returns true if the error indicates the lock is held by another process.
func isLockBusy(err error) bool {
	return err == syscall.EWOULDBLOCK || err == syscall.EAGAIN
}
