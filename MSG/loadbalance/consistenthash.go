package loadbalance

import (
	"errors"
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

// 1. 计算存储结点「服务器」哈希值，将其存储空间抽象成一个环
// 2. 对数据「URL、IP」进行哈希计算，按「顺时针方向」将其映射到距离最近的结点上
type ConsistentHashBalance struct {
	// Hash函数，支持用户自定义，默认使用 crc32.ChecksumIEEE
	// 运算效率、散列均匀
	hash Hash
	// 服务器结点 hash值列表,从小到达排序
	hashKeys UInt32Slice
	// 服务器结点 hash值与真实地址的映射表
	hashMap map[uint32]string
	/*
		由于map是无序的，额外通过一个 排序的hashKeys 通过二分查找来确定选择哪个结点
	*/
	// 虚拟结点倍数
	// 解决平衡性问题
	replicas int

	// 由于是并发的，map不支持并发，因此需要加锁
	mux sync.RWMutex
}

type Hash func(data []byte) uint32

type UInt32Slice []uint32

func (s UInt32Slice) Len() int {
	return len(s)
}

func (s UInt32Slice) Less(i, j int) bool {
	return s[i] < s[j]
}
func (s UInt32Slice) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func NewConsistentHashBalance(replicas int, fn Hash) *ConsistentHashBalance {
	ch := &ConsistentHashBalance{
		hash:     fn,
		hashMap:  map[uint32]string{},
		replicas: replicas,
	}
	if ch.hash == nil {
		ch.hash = crc32.ChecksumIEEE
	}
	return ch
}

// 添加服务器结点地址
// 对每一个真实结点 arr，对应创建c.replicas个虚拟结点
// 用c.hash计算虚拟结点hash值,添加到环上
// 最后，对hashKeys进行排序
func (c *ConsistentHashBalance) Add(servers ...string) error {
	if len(servers) == 0 {
		return errors.New("servers length at least one!")
	}
	c.mux.Lock()
	defer c.mux.Unlock()
	for _, addr := range servers {
		//  虚拟结点 c.replicas
		for i := 0; i < c.replicas; i++ {
			hash := c.hash([]byte(strconv.Itoa(i) + addr))
			c.hashKeys = append(c.hashKeys, hash)
			c.hashMap[hash] = addr
		}
	}
	// 必须实现sort.Interface接口
	sort.Sort(c.hashKeys) // 排序-->后续通过二分进行查找
	return nil
}

// 获取hash后的服务器结点地址
// 可能会穿越起点
// 实现步骤：
// 1. 计算key哈希值
// 2. 通过二分查找「最优的服务器姐弟」
func (c *ConsistentHashBalance) Get(key string) (string, error) {
	l := len(c.hashKeys)
	if l == 0 {
		return "", errors.New("node list is empty!")
	}
	// 计算key的hash值
	hash := c.hash([]byte(key))
	c.mux.RLock() // 加读锁
	defer c.mux.RUnlock()
	idx := sort.Search(l, func(i int) bool {
		return c.hashKeys[i] >= hash
	})
	if idx == l { // 查找结果大于服务器结点的最大索引
		idx = 0
	}

	return c.hashMap[c.hashKeys[idx]], nil
}
