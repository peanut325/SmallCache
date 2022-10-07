package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

// 允许替换为自定义的hash函数
type Hash func(data []byte) uint32

// Map包含所有的hash的key
type Map struct {
	hash     Hash			// Hash计算函数
	replicas int			// 虚拟节点倍数
	keys     []int			// Hash环，被排序
	hashMap  map[int]string	// key:虚拟节点hash值 value:实际节点
}

// 构造函数
func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas,
		hash:     fn,
		hashMap:  make(map[int]string),
	}

	// hash默认为crc32.ChecksumIEEE算法
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}

	return m
}

// 添加节点
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// 获取最近的hash节点
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))
	// 查找第一个key的hash的值的虚拟节点的hash，没有找到则返回长度，而不是-1
	// https://studygolang.com/articles/14087
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})

	return m.hashMap[m.keys[idx%len(m.keys)]]
}
