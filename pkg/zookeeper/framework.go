package zookeeper

import "github.com/samuel/go-zookeeper/zk"

type Framework struct {
	Conn *zk.Conn // zookeeper connect
	ACL  []zk.ACL
}

func NewFramework(conn *zk.Conn, acl []zk.ACL) *Framework {
	return &Framework{
		Conn: conn,
		ACL:  acl,
	}
}
