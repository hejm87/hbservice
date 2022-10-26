package gateway_define

type GatewayListen struct {
	Iface		string
	Port		int
}

type GatewayFlowLimit struct {
	InputLimit			int		// 输入流量限制（单位：MB)
	OutputLimit			int		// 输出流量限制（单位：MB）
}

type GatewayCommon struct {
	MaxUserCount		int		// 最大连接数
	UserHeartbeatSec	int		// 心跳周期
	MergePacketSec		int		// 合包周期
}

type ReportInfo struct {
	Url					string		// 上报url
	Period				int			// 上报周期（单位：秒）
}

type GatewayConfig struct {
	Listen				GatewayListen
	FlowLimit			GatewayFlowLimit
	Common				GatewayCommon
	Report				ReportInfo
}

