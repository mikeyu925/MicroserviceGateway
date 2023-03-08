package zookeeper

import (
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"time"
)

// zookeeper 管理器
type ZkManager struct {
	hosts      []string // 主机列表zw
	conn       *zk.Conn // zookeeper连接
	pathPrefix string   // 路径前缀
}

func NewZkManager(hosts []string) *ZkManager {
	return &ZkManager{hosts: hosts, pathPrefix: "/gateway_servers_"}
}

func (z *ZkManager) GetConnect() error {
	conn, _, err := zk.Connect(z.hosts, 5*time.Second)
	if err != nil {
		return err
	}
	z.conn = conn
	return nil
}

func (z *ZkManager) Close() {
	z.conn.Close()
}
func (z *ZkManager) GetPathData(nodePath string) ([]byte, *zk.Stat, error) {
	return z.conn.Get(nodePath)
}

func (z *ZkManager) GetServerListByPath(path string) (list []string, err error) {
	list, _, err = z.conn.Children(path) // 获取当前路径下的所有子结点
	return
}

func (z *ZkManager) WatchServerListByPath(path string) (chan []string, chan error) {
	conn := z.conn

	snapshots := make(chan []string)
	errors := make(chan error)

	go func() {
		for {
			snapshot, _, events, err := conn.ChildrenW(path)
			if err != nil {
				errors <- err
			}
			snapshots <- snapshot
			select {
			case evt := <-events:
				if evt.Err != nil {
					errors <- evt.Err
				}
				fmt.Printf(evt.Err.Error())
			}
		}
	}()
	return snapshots, errors
}

func (z *ZkManager) RegisterServerPath(nodePath, host string) (err error) {
	ex, _, err := z.conn.Exists(nodePath)
	if err != nil {
		return err
	}
	if !ex {
		_, err = z.conn.Create(nodePath, nil, 0, zk.WorldACL(zk.PermAll))
		if err != nil {
			fmt.Println("Create error", nodePath)
			return err
		}
	}
	return nil
}
