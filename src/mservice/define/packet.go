package mservice_define

///////////////////////////////////////////////
//			MicroServicePacket
///////////////////////////////////////////////
type MServiceHeader struct {
	Service			string
	Method			string
	TraceID			string
}

type MServicePacket struct {
	Header			MServiceHeader
	Body			interface {}
}

func (p *MServicePacket) ToStr() string {
	return ""
}

///////////////////////////////////////////////
//			MicroGatewayPacket
///////////////////////////////////////////////
const (
	GW_LOGIN		string = "LOGIN"
	GW_PING			string = "PING"
	GW_PONG			string = "PONG"
	GW_CMD			string = "CMD"
)

type MGatewayHeader struct {
	Cmd				string
	Uid				string
}

type MGatewayPacket struct {
	Header			MGatewayHeader
	Body			interface {}
}
