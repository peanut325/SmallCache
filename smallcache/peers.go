package smallcache

import pb "SmallCache/smallcache/smallcachepb"

//使用一致性哈希选择节点        是                                    是
//    |-----> 是否是远程节点 -----> HTTP 客户端访问远程节点 --> 成功？-----> 服务端返回返回值
//                    |  否                                    ↓  否
//                    |----------------------------> 回退到本地节点处理。

// 根据传入的key来获取PeerGetter
type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool)
}

// 从对应的group中查找缓存值，类似于上述的客户端
type PeerGetter interface {
	Get(in *pb.Request, out *pb.Response) error
}