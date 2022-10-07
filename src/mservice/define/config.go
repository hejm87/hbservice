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

type ServiceInfo struct {
	Name				string		// 服务名
	ListenPort			int
}

type MServiceConfig struct {
	EtcdCfg					EtcdConfig
	ObjClientCfg			ObjClientConfig
	Service					ServiceInfo
}