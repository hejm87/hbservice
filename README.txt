功能测试点：
1. NetRecv的超时问题测试
2. socket对端关闭的时候，服务端的表现是怎样的
3. client第二次启动，连接服务端会报错               --> 已解决
4. 关于各模块的优雅关闭问题
5. 将client端连接断开的报错改为正常断开日志
6. etcd的client是否线程安全，内部实现是否池化
7. pusher_handle.go, get_uid_addr存在并发性能问题
8. obj_client需要区分call, cast处理


微服务网关功能要求：
1. 最大连接数限制
2. 用户心跳（保活）
3. 限流
4. 合包