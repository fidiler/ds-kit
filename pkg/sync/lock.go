package sync

type Locker interface {
	Lock() bool
	UnLock() bool
}

type RWLocker interface {
	RLock() bool
	UnRLock() bool
	Lock() bool
	UnLock()
}
