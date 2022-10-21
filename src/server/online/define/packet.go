package online_define

import (
	"hbservice/src/mservice/define"
)

const (
	// for gateway
	MS_ONLINE_HEARTBEAT				string = "MS_ONLINE_HEARTBEAT"

	// for mservice
	MS_ONLINE_GET_USER_NODE			string = "MS_ONLINE_GET_USER_NODE"
	MS_ONLINE_GET_USER_NODE_BATCH	string = "MS_ONLINE_GET_USER_NODE_BATCH"

	// error code
	ERR_NOT_EXIST_USER				string = "not exist user"

	SERVICE_NAME					string = "online"
)

type OLHeartbeatReq struct {
	User 			UserNode
}

type OLHeartbeatResp struct {
	Err				string
}

type OLGetUserNodeReq struct {
	Uid				string
}

type OLGetUserNodeResp struct {
	Err				string
	User			UserNode
}

type OLGetUserNodeBatchReq struct {
	Uids			[]string
}

type OLGetUserNodeBatchResp struct {
	Err				string
	Users			[]UserNode
}


type UserNode {
	Uid				string
	ServiceId		string
	Host			string
	Port			int
}
