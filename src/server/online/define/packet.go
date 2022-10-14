package online_define

import (
	"hbservice/src/mservice/define"
)

const (
	// for gateway
	MS_ONLINE_LOGIN				string = "MS_ONLINE_LOGIN"
	MS_ONLINE_LOGOUT			string = "MS_ONLINE_LOGOUT"
	MS_ONLINE_HEARTBEA			string = "MS_ONLINE_HEARTBEAT"

	// for mservice
	MS_ONLINE_GET_UID_ADDR		string = "MS_ONLINE_GET_UID_ADDR"

	SERVICE_NAME				string = "online"
)

type OLLoginReq struct {
	Uid				string
	ChannelId		string
}

type OLLoginResp struct {
	Err				error
}

type OLLogoutReq struct {
	Uid				string
}

type OLLogoutResp struct {
	Err				error
}

type OLHeartbeatReq struct {
	Uid				string
	ChannelId		string
}

type OLHeartbeatResp struct {
	Err				error
}

type OLGetUidAddrReq struct {
	Uid				string
}

type OLGetUidAddrResp struct {
	Err				error
	Addr			string
}
