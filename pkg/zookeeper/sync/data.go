package sync

import "sync/atomic"

type lockData struct {
	lockPath  atomic.Value
	lockCount int32
}

func newLockData() *lockData {
	ld := new(lockData)
	atomic.StoreInt32(&ld.lockCount, 0)
	ld.lockPath.Store("")
	return ld
}

func (l *lockData) lockUse() bool {
	return atomic.LoadInt32(&l.lockCount) != 0 && l.lockPath.Load().(string) != ""
}

func (l *lockData) getLockCount() int32 {
	return atomic.LoadInt32(&l.lockCount)
}

func (l *lockData) incrementLockCount() {
	atomic.AddInt32(&l.lockCount, 1)
}

func (l *lockData) decrementLockCount() {
	atomic.AddInt32(&l.lockCount, -1)
}

func (l *lockData) setLockPath(path string) {
	l.lockPath.Store(path)
}

func (l *lockData) getLockPath() string {
	return l.lockPath.Load().(string)
}

func (l *lockData) cleanLockData() {
	atomic.StoreInt32(&l.lockCount, 0)
	l.lockPath.Store("")
}
