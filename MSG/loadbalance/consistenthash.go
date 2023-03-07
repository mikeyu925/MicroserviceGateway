package loadbalance

// 1. 计算存储结点「服务器」哈希值，将其存储空间抽象成一个环
// 2. 对数据「URL、IP」进行哈希计算，按「顺时针方向」将其映射到距离最近的结点上
type ConsistentHashBalance struct {
	hash Hash // Hash函数，支持用户自定义，默认使用 crc32.ChecksumIEEE

}

type Hash func(data []byte) uint32{

}