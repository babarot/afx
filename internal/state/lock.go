package state

import (
	"os"
	"syscall"
)

// fileLock provides file-based locking to prevent concurrent access to state.json.
type fileLock struct {
	path string
	f    *os.File
}

func newFileLock(path string) *fileLock {
	return &fileLock{path: path + ".lock"}
}

func (l *fileLock) lock() error {
	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		// Cannot create lock file (e.g. non-existent directory in tests)
		return err
	}
	l.f = f
	// Use LOCK_NB (non-blocking) to avoid hanging if another process holds the lock
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		f.Close()
		l.f = nil
		return err
	}
	return nil
}

func (l *fileLock) unlock() error {
	if l.f == nil {
		return nil
	}
	if err := syscall.Flock(int(l.f.Fd()), syscall.LOCK_UN); err != nil {
		return err
	}
	l.f.Close()
	os.Remove(l.path)
	return nil
}
