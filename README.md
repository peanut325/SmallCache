# 🚀SmallCache
基于Go语言实现简单的分布式缓存。
# ✍️项目亮点
- 使用LRU算法作为缓存淘汰策略
- 实现了一致性哈希，来作为分布式获取缓存的优化策略
- 解决了并发请求导致缓存击穿的问题
- 使用Protobuf进行节点之间的通讯，提高了效率
