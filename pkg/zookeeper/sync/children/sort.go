package children

import (
	"sort"
)

func Sort(children []string, lockName string, f func(completeLockName, lockName string) string) {
	collectionSort := &collectionSort{
		child:               children,
		lockName:            lockName,
		compareLockNameFunc: f,
	}

	sort.Sort(collectionSort)
}

type collectionSort struct {
	child               []string
	lockName            string
	compareLockNameFunc func(completeLockName, lockName string) string
}

func (c *collectionSort) Len() int {
	return len(c.child)
}

func (c *collectionSort) Less(i, j int) bool {
	return c.compareLockNameFunc(c.child[i], c.lockName) < c.compareLockNameFunc(c.child[j], c.lockName)
}

func (c *collectionSort) Swap(i, j int) {
	c.child[i], c.child[j] = c.child[j], c.child[i]
}