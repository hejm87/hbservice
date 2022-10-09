package mservice_define

type EtcdConfig struct {
	Addrs			[]string
	Username		string
	Password		string
}

type ObjClientConfig struct {
	HeartbeatTs				int		// 心跳发送间隔
	CallTimeoutSec			int		// 同步调用超时（单位：秒）
	InvalidSeriesFailCount	int		// 连续失败N次为失效连接
}

type CommonConfig struct {
	OpenPerformanceMonitor	bool
}

type ServiceInfo struct {
	Name				string		// 服务名
	ListenHost			string		// 监听地址
	ListenPort			int			// 监听端口
}

type MServiceConfig struct {
	EtcdCfg					EtcdConfig
	ObjClientCfg			ObjClientConfig
	CommonCfg				CommonConfig
	Service					ServiceInfo
}