package sync

import (
	"fmt"
	"github.com/ds-kit/pkg/zookeeper"
	"github.com/ds-kit/pkg/zookeeper/sync/children"
	"github.com/ds-kit/pkg/zookeeper/sync/driver"
	"github.com/ds-kit/pkg/zookeeper/util"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"time"
)

var LOCK_NAME = "lock-"

type DSMutex struct {
	mu         sync.Mutex
	path       string
	parentPath string
	standard   *driver.Standard
	framework  *zookeeper.Framework
	lockdata   *lockData
}

func NewDSMutex(framework *zookeeper.Framework, standardDriver *driver.Standard, path string) (*DSMutex, error) {
	if err := util.ValidatePath(path); err != nil {
		return nil, err
	}

	mu := &DSMutex{
		framework:  framework,
		standard:   standardDriver,
		lockdata:   newLockData(),
		path:       util.MakePath(path, LOCK_NAME),
		parentPath: path,
	}

	return mu, nil
}

func (d *DSMutex) Acquire() error {
	if d.lockdata.lockUse() {
		d.lockdata.incrementLockCount()
		return nil
	}

	attempted, lockPath, err := d.attemptLock(-1)
	if err != nil {
		return err
	}

	if attempted {
		d.lockdata.incrementLockCount()
		d.lockdata.setLockPath(lockPath)
	}

	return nil
}

func (d *DSMutex) AcquireTimeout(milliWaitTime int64) error {
	if d.lockdata.lockUse() {
		d.lockdata.incrementLockCount()
		return nil
	}

	attempted, lockPath, err := d.attemptLock(milliWaitTime)
	if err != nil {
		return err
	}

	if attempted {
		d.lockdata.incrementLockCount()
		d.lockdata.setLockPath(lockPath)
	}

	return nil
}

func (d *DSMutex) attemptLock(milliWaitTime int64) (attempted bool, lockPath string, err error) {
	if lockPath, err = d.standard.CreateLock(d.framework, d.parentPath, d.path); err != nil {
		return
	}

	startTime := time.Now()
	if attempted, err = d.tryAttemptLock(lockPath, startTime, milliWaitTime); err != nil {
		return
	}

	return
}

func (d *DSMutex) tryAttemptLock(path string, startTime time.Time, milliWait int64) (bool, error) {
	// TODO zookeeper not start case
	for {
		// get sorted child for lock
		child, err := d.GetSortedChild()
		if err != nil {
			return false, err

		}

		filteredChildren := d.standard.FilterChildrenCollection(child, path, path[len(d.parentPath)+1:], 1)
		// the current child is first in zookeeper
		if filteredChildren.IsFirstChild {
			return true, nil
		}

		// to watch previous node
		_, _, eventCh, err := d.framework.Conn.GetW(d.parentPath + "/" + filteredChildren.WatchPath)
		if err == nil {
			if milliWait > 0 {
				milliWait -= time.Now().Sub(startTime).Milliseconds()
				startTime = time.Now()

				// Acquire lock is timeout
				if milliWait <= 0 {
					fmt.Println("before: ", milliWait)
					_ = d.framework.Conn.Delete(path, -1)
					return false, ErrAcquireTimeout

				} else {
					// wait previous node or Acquire lock timeout
					select {
					case <-eventCh:
						break
					case <-time.Tick(time.Millisecond * time.Duration(milliWait)):
						break
					}
				}

			} else {
				// only wait previous node
				<-eventCh
			}

		} else if err == zk.ErrNoNode {

			// Acquire lock attach timeout
			if milliWait > 0 {
				milliWait -= time.Now().Sub(startTime).Milliseconds()
				startTime = time.Now()

				// Acquire lock is timeout
				if milliWait <= 0 {
					_ = d.framework.Conn.Delete(path, -1)
					return false, ErrAcquireTimeout
				}

				// it has been deleted (lock released). to try Acquire again until timeout
			}

			// it has been deleted (lock released). to try Acquire again

		} else {
			return false, nil
		}
	}
}

func (d *DSMutex) GetSortedChild() ([]string, error) {
	child, _, err := d.framework.Conn.Children(d.parentPath)
	if err != nil {
		return nil, err
	}

	children.Sort(child, LOCK_NAME, d.standard.LockCompareName)

	return child, nil
}

func (d *DSMutex) Release() error {
	return d.releaseLock()
}

func (d *DSMutex) releaseLock() error {
	if !d.lockdata.lockUse() {
		panic("you do not own the lock: " + d.path)
	}

	d.lockdata.decrementLockCount()

	if d.lockdata.getLockCount() > 0 {
		return nil
	}

	if d.lockdata.getLockCount() < 0 {
		panic("lock count has gone negative for lock: " + d.path)
	}

	if err := d.standard.DeleteLock(d.framework, d.lockdata.getLockPath()); err != nil {
		return err
	}

	d.lockdata.cleanLockData()

	return nil
}
