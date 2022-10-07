package singleflight

import "sync"

// 代表正在进行中，或者已经结束了的请求
type call struct {
	wg  sync.WaitGroup
	val interface{}
	err error
}

// 管理不同的key的请求(call)
type Group struct {
	mu sync.Mutex       // 保证m不被并发读写而加上锁
	m  map[string]*call // lazily initialized
}

// 无论do被调用多少次，fn函数只会被执行一次
func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error) {
	g.mu.Lock()
	// lazy load
	if g.m == nil {
		g.m = make(map[string]*call)
	}

	// 如果有请求正在执行，那么就会等待他处理完毕之后一起返回结果
	if c, ok := g.m[key]; ok {
		g.mu.Unlock()
		c.wg.Wait()	// 阻塞等待直到锁被释放
		return c.val, c.err
	}

	// 如果没有则创建call，加入map
	c := new(call)
	c.wg.Add(1)	// 加锁
	g.m[key] = c
	g.mu.Unlock()

	c.val, c.err = fn()	// 调用fn，发送请求
	c.wg.Done()	// 请求结束

	g.mu.Lock()
	delete(g.m, key)
	g.mu.Unlock()

	return c.val, c.err
}
