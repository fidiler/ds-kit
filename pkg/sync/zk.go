package sync

import (
	"errors"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

type ZKSync struct {
	conn *zk.Conn
	lock *zkLock
}

func NewZKSync(conn *zk.Conn) *ZKSync {
	v := new(ZKSync)
	v.conn = conn
	return v
}

func (zs *ZKSync) Lock() error {
	return zs.lock.lock(zs.conn)
}

func (zs *ZKSync) TryLock(timeout int) error {
	return zs.lock.tryLock(zs.conn, timeout)
}

func (zs *ZKSync) UnLock() error {
	return zs.lock.unLock(zs.conn)
}

type zkLock struct {
	path string
}

func (zl *zkLock) lock(conn *zk.Conn) error {
	exist, _, ch, err := conn.ExistsW(zl.path)
	if err != nil {
		return err
	}

	if !exist {
		return zl.createNode(conn)
	}

	for {
		event := <-ch
		switch event.Type {
		case zk.EventNodeDeleted: // lock is release

			// created lock error, retry
			if err = zl.createNode(conn); err != nil && err != zk.ErrNodeExists {
				return err
			}

			// lock is already created
			continue
		}
	}
}

func (zl *zkLock) createNode(conn *zk.Conn) error {
	_, err := conn.Create(zl.path, []byte{}, zk.FlagEphemeral, nil)
	return err
}

func (zl *zkLock) tryLock(conn *zk.Conn, timeout int) error {
	exist, _, ch, err := conn.ExistsW(zl.path)
	if err != nil {
		return err
	}

	if !exist {
		return zl.createNode(conn)
	}

	tick := time.Tick(time.Duration(timeout) * time.Second)

	for {
		select {
		case event := <-ch:
			// lock is release event
			if event.Type == zk.EventNodeDeleted {
				// created lock error, retry
				if err = zl.createNode(conn); err != nil && err != zk.ErrNodeExists {
					return err
				}

				// lock is already created
				continue
			}

		case <-tick:
			return errors.New("lock timeout")
		}
	}
}

func (zl *zkLock) unLock(conn *zk.Conn) error {
	return conn.Delete(zl.path, -1)
}

type zkRWLock struct {
	path string
}

func (zrw *zkRWLock) lock() {

}

func (zrw *zkRWLock) createNode(conn *zk.Conn) (string, error) {
	return conn.Create(zrw.path, []byte{}, zk.FlagEphemeral, nil)
}