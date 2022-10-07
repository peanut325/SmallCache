package smallcache

import (
	"fmt"
	pb "SmallCache/smallcache/smallcachepb"
	"SmallCache/smallcache/singleflight"
	"log"
	"sync"
)

//                               是
// 接收 key --> 检查是否被缓存 -----> 返回缓存值 ⑴
// |  否                         是
// |-----> 是否应当从远程节点获取 -----> 与远程节点交互 --> 返回缓存值 ⑵
// 			|  否
//			|-----> 调用`回调函数`，获取值并添加到缓存 --> 返回缓存值 ⑶

// Group负责与用户交互，并且控制缓存值的存储和获取的流程
// 是一个缓存命名空间，例如:成绩为score、学生信息为info、课程名称为course
type Group struct {
	name      string
	getter    Getter			// 缓存未命中时回调获取数据源
	mainCache cache				// 一开始实现的并发缓存
	peers     PeerPicker		// 获取远程节点的接口
	loader *singleflight.Group	// 通过singleflight.Group确保多个请求都只会获取到一个缓存，防止缓存击穿
}

// 通过key来加载数据
type Getter interface {
	Get(key string) ([]byte, error)
}

// 通过一个函数来实现Getter
// 函数式接口:https://geektutu.com/post/7days-golang-q1.html
// 类似于Java中的Lamda表达式
type GetterFunc func(key string) ([]byte, error)

// Get实现Getter接口的功能
func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

var (
	mu     sync.RWMutex
	groups = make(map[string]*Group)
)

// 构造方法
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: cache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}
	groups[name] = g
	return g
}

// 根据name返回group
func GetGroup(name string) *Group {
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

// 根据key从cache中获取值
func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("[GeeCache] hit")
		return v, nil
	}

	return g.load(key)
}

// 将实现了 PeerPicker 接口的 HTTPPool 注入到 Group 中
func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

// 如果是非本地节点，那么就调用
func (g *Group) load(key string) (value ByteView, err error) {
	// 无论并发调用者的数量如何，每个键都只获取一次(本地或远程)
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			// 使用PickPeer()选择节点，如果非本机节点，那么就调用getFromPeer()远程获取，如果本机处理失败，调用getLocally()
			if peer, ok := g.peers.PickPeer(key); ok {
				if value, err = g.getFromPeer(peer, key); err == nil {
					return value, nil
				}
				log.Println("[GeeCache] Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return viewi.(ByteView), nil
	}
	
	return
}

// 添加到缓存中操作
func (g *Group) populateCache(key string, value ByteView) {
	g.mainCache.add(key, value)
}

// 回调获取缓存，并添加到缓存中
func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	value := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

// 使用实现了 PeerGetter 接口的 httpGetter 从访问远程节点，获取缓存值
func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	req := &pb.Request{
		Group: g.name,
		Key:   key,
	}
	res := &pb.Response{}
	err := peer.Get(req, res)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: res.Value}, nil
}
