package pusher_define

type PusherConfig struct {
	UserCacheSec			int
	UserCacheSize			int
	BroadcastUidCount		int		// 超过设定阈值则广播所有网关
	PushChannelCount		int		// 后台推送channel缓冲大小
	RetryClientConnectSec	int		// 客户端重连周期
}