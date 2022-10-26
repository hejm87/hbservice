package online_define

const (
	// for gateway
	MS_ONLINE_LOGIN					string = "MS_ONLINE_LOGIN"
	MS_ONLINE_LOGOUT				string = "MS_ONLINE_LOGOUT"
	MS_ONLINE_HEARTBEAT				string = "MS_ONLINE_HEARTBEAT"

	// for mservice
	MS_ONLINE_GET_USER_NODE			string = "MS_ONLINE_GET_USER_NODE"
	MS_ONLINE_GET_USER_NODE_BATCH	string = "MS_ONLINE_GET_USER_NODE_BATCH"

	// error code
	ERR_INNER_SERVER				string = "inner server error"
	ERR_USER_NOT_LOGIN				string = "user not login"
	ERR_USER_ALREADY_LOGIN			string = "user already login"
	ERR_USER_NOT_EXISTS				string = "user not exists"

	SERVICE_NAME					string = "online"
)

type OLLoginReq struct {
	User			UserNode
}

type OLLoginResp struct {
	Err				string
}

type OLLogoutReq struct {
	Uid				string
}

type OLLogoutResp struct {
	Err				string
}

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

type UserNode struct {
	Uid				string
	ServiceId		string
	Host			string
	Port			int
}
