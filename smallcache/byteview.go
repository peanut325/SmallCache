package smallcache

// 只读的数据结构保存缓存值，实现Value接口
type ByteView struct {
	b []byte	// byte能表示任意数据类型
}

// 缓存对象必须实现Value接口，返回缓存对象的长度
func (v ByteView) Len() int {
	return len(v.b)
}

// b是只读的，返回一个拷贝，防止缓存值被外部就该
func (v ByteView) ByteSlice() []byte {
	return cloneBytes(v.b)
}

// 转化为字符串
func (v ByteView) String() string {
	return string(v.b)
}

// 拷贝数组方法
func cloneBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	return c
}