package gateway_define

type GatewayListen struct {
	Interface	string
	Port		int
}

type GatewayFlowLimit struct {
	InputLimit			int		// 输入流量限制
	OutputLimit			int		// 输出流量限制
}

type GatewayCommon struct {
	MaxUserCount		int		// 最大连接数
	UserHeartbeatSec	int		// 心跳周期
	MergePacketSec		int		// 合包周期
}

type GatewayConfig struct {
	Listen				GatewayListen
	FlowLimit			GatewayFlowLimit
	Common				GatewayCommon
}

