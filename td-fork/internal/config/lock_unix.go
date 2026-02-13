//go:build unix

package config

import "syscall"

// fileLock acquires an exclusive lock on the file descriptor.
func fileLock(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_EX)
}

// fileUnlock releases the lock on the file descriptor.
func fileUnlock(fd uintptr) error {
	return syscall.Flock(int(fd), syscall.LOCK_UN)
}
