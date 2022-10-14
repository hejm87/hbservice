package mservice_define

const (
	SERVICE_ETCD_CONFIG_PATH		string = "/service/server/config"
)

type EtcdConfig struct {
	Addrs			[]string
	Username		string
	Password		string
}

type ObjClientConfig struct {
	HeartbeatTs				int		// 心跳发送间隔
	RetryConnectSec			int		// 重连时长
	CallTimeoutSec			int		// 同步调用超时（单位：秒）
	InvalidSeriesFailCount	int		// 连续失败N次为失效连接
}

type CommonConfig struct {
}

type RedisConfig struct {
	Addr				string
	Password			string
	DbIndex				int
}

type ServiceInfo struct {
	Name					string		// 服务名
	ListenHost				string		// 监听地址
	ListenPort				int			// 监听端口
	KeepAliveTtl			int			// 服务上报etcd的ttl
	OpenPerformanceMonitor	bool		// 是否开启性能监控
}

type MServiceConfig struct {
	EtcdCfg					EtcdConfig
	ObjClientCfg			ObjClientConfig
	RedisCfg				RedisConfig
	Service					ServiceInfo
}