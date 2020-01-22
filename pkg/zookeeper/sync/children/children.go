package children

type Collection struct {
	WatchPath    string
	IsFirstChild bool
}

func (c *Collection) ChildIndexOf(children []string, pathIndexName string) int {
	for i, n := 0, len(children); i < n; i++ {
		if children[i] == pathIndexName {
			return i
		}
	}
	return -1
}