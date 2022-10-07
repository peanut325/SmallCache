package lru

import "container/list"

// LRU缓存实现，线程不安全
type Cache struct {
	maxBytes int64							// 允许使用的最大内存
	nbytes   int64							// 当前已使用的内存
	ll       *list.List						// 双向链表实现
	cache    map[string]*list.Element		// 键是字符串，值是双向链表对应节点的指针
	OnEvicted func(key string, value Value)	// 当Entry被清除的时候执行，可以自定义
}

// 节点类型
type entry struct {
	key   string
	value Value
}

// 为了通用性，我们允许值是实现了 Value 接口的任意类型，该接口只包含了一个方法 Len() int，用于返回值所占用的内存大小
type Value interface {
	Len() int
}

// 构造函数
func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes:  maxBytes,
		ll:        list.New(),
		cache:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}
}

// 添加节点
func (c *Cache) Add(key string, value Value) {
	if ele, ok := c.cache[key]; ok {
		// 存在节点，修改即可
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	} else {
		// 不存在进行加入
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}

	// 超过最大大小，此时删除节点
	for c.maxBytes != 0 && c.maxBytes < c.nbytes {
		c.RemoveOldest()
	}
}

// 获取节点
func (c *Cache) Get(key string) (value Value, ok bool) {
	if ele, ok := c.cache[key]; ok {
		// 获取了之后需要移动到头节点
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true
	}

	return
}

// 删除节点
func (c *Cache) RemoveOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())

		// 调用删除的回调函数
		if c.OnEvicted != nil {
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

// 计算值的长度
func (c *Cache) Len() int {
	return c.ll.Len()
}