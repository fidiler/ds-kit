package sync

import (
	"fmt"
	"github.com/dslock/pkg/zookeeper"
	"github.com/dslock/pkg/zookeeper/sync/children"
	"github.com/dslock/pkg/zookeeper/sync/driver"
	"github.com/dslock/pkg/zookeeper/util"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"sync/atomic"
	"time"
)

var LOCK_NAME = "lock-"

var gid uint32

var once sync.Once


type DSMutex struct {
	Path       string
	ParentPath string
	LockPath   string
	Standard   *driver.Standard
	framework  *zookeeper.Framework

	gdata sync.Map
	mu    sync.Mutex
}

func NewDSMutex(framework *zookeeper.Framework, standardDriver *driver.Standard, path string) (*DSMutex, error) {
	once.Do(func() {
		atomic.StoreUint32(&gid, 1)
	})

	atomic.AddUint32(&gid, 1)

	if err := util.ValidatePath(path); err != nil {
		return nil, err
	}

	mu := &DSMutex{
		framework:  framework,
		Standard:   standardDriver,
		Path:       util.MakePath(path, LOCK_NAME),
		ParentPath: path,
	}

}

func (d *DSMutex) Acquire() error {
	if err := d.attemptLock(-1); err != nil {
		return err
	}

	return nil
}

func (d *DSMutex) AcquireTimeout(milliWaitTime int64) error {
	if err := d.attemptLock(milliWaitTime); err != nil {
		return err
	}

	return nil
}

func (d *DSMutex) attemptLock(milliWaitTime int64) error {
	path, err := d.Standard.CreateLock(d.framework, d.ParentPath, d.Path)
	if err != nil {
		return err
	}

	startTime := time.Now()
	attempted, err := d.tryAttemptLock(path, startTime, milliWaitTime)
	if err != nil {
		return err
	}

	if attempted {
		d.LockPath = path
	}

	return nil
}

func (d *DSMutex) tryAttemptLock(path string, startTime time.Time, milliWait int64) (bool, error) {
	// TODO zookeeper not start case
	for {
		// get sorted child for lock
		child, err := d.GetSortedChild()
		if err != nil {
			return false, err

		}

		filteredChildren := d.Standard.FilterChildrenCollection(child, path, path[len(d.ParentPath)+1:], 1)
		// the current child is first in zookeeper
		if filteredChildren.IsFirstChild {
			return true, nil
		}

		// to watch previous node
		_, _, eventCh, err := d.framework.Conn.GetW(d.ParentPath + "/" + filteredChildren.WatchPath)
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
	child, _, err := d.framework.Conn.Children(d.ParentPath)
	if err != nil {
		return nil, err
	}

	children.Sort(child, LOCK_NAME, d.Standard.LockCompareName)

	return child, nil
}

func (d *DSMutex) Release() error {
	return d.releaseLock()
}

func (d *DSMutex) releaseLock() error {
	if d.LockPath != "" {
		err := d.Standard.DeleteLock(d.framework, d.LockPath)
		if err != nil {
			return err
		}
	}

	d.LockPath = ""
	return nil
}
