package driver

import (
	"fmt"
	"github.com/ds-kit/pkg/zookeeper"
	"github.com/ds-kit/pkg/zookeeper/sync/children"
	"github.com/samuel/go-zookeeper/zk"
	"strings"
)

type Standard struct {
}

func (s *Standard) CreateLock(framework *zookeeper.Framework, parentPath, path string) (string, error) {
	for {
		path, err := framework.Conn.CreateProtectedEphemeralSequential(path, []byte{}, framework.ACL)
		if err == zk.ErrNoNode {
			// create parent
			parts := strings.Split(parentPath, "/")
			pth := ""
			for _, p := range parts[1:] {
				var exists bool
				pth += "/" + p
				exists, _, err = framework.Conn.Exists(pth)
				if err != nil {
					return "", err
				}

				if exists == true {
					continue
				}
				_, err = framework.Conn.Create(pth, []byte{}, 0, framework.ACL)
				if err != nil && err != zk.ErrNodeExists {
					return "", err
				}
			}
		} else if err != nil {
			return "", err

		} else {
			return path, err
		}
	}
}

func (s *Standard) DeleteLock(framework *zookeeper.Framework, path string) error{
	return framework.Conn.Delete(path, -1)
}

func (s *Standard) FilterChildrenCollection(child []string, completePath, pathIndexName string, maxLease int) *children.Collection {
	collection := new(children.Collection)
	pathIndex := collection.ChildIndexOf(child, pathIndexName)
	if pathIndex == -1 {
		panic(fmt.Sprintf("sequential path not found: %s", completePath))
	}

	collection.IsFirstChild = pathIndex < maxLease
	if !collection.IsFirstChild {
		collection.WatchPath = child[pathIndex-maxLease]
	}

	return collection
}

func (s *Standard) LockCompareName(completeLockName, lockName string) string {
	lastIndex := strings.LastIndex(completeLockName, lockName)
	if lastIndex >= 0 {
		lastIndex += len(lockName)

		if lastIndex <= len(completeLockName) {
			return completeLockName[lastIndex:]
		}

		return ""
	}

	return completeLockName
}
