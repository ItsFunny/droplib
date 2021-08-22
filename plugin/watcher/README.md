# How To Use
- 创建一个动态变化的watcher
    ```buildoutcfg
    watcher := NewChannelWatcher()
    watcher.BStart()
    即可
    ```
- 创建指定某种类型的watcher
    ```buildoutcfg
    watcher:=NewForeverWatcher(wh.watchType)
    watcher.BStart()即可
    ```
# TODO
- selectn routineSize并没有严格计算,因为watcher#Size函数并没有使用
- selectn mergeRegion 因为golang select case 的特性,会触发耗时久的情况
## 内存
- 字段压缩合并
## 性能
- cache
 - 需要提供gc
 - region需要parallel
 - reflect需要concurrent
## 安全
- 确保无内存泄漏
# 测试
- 需要更多更多的随机测试