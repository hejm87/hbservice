package hbservice_test

import (
	"testing"
	"encoding/json"
	"github.com/mitchellh/mapstructure"
)

type TestResp struct {
    A       int
    B       string
}

type PushPacketItem struct {
    Cmd     string
    Body    interface {}
}

type PushPacket struct {
    Count       int
    Items       []PushPacketItem
}

func TestJsonToMapstruct (t *testing.T) {
    var packet PushPacket
    packet.Count = 2
    packet.Items = append(packet.Items, make_item("test", 1, "hello1"))
    packet.Items = append(packet.Items, make_item("test", 2, "hello2"))

    var err error
    var bytes []byte
    var cpacket PushPacket

    t.Logf("############### convert_json ...\n")
    if bytes, err = json.Marshal(&packet); err != nil {
        t.Logf("convert_json, error:%#v\n", err)
    }
    t.Logf("json:%s\n", string(bytes))

    t.Logf("############### convert_struct ...\n")
    if err = json.Unmarshal(bytes, &cpacket); err == nil {
        t.Logf("convert_struct, error:%#v\n", err)
    }
    t.Logf("struct.Count:%d\n", cpacket.Count)
    for i := 0; i < cpacket.Count; i += 1 {
        var resp TestResp
        if err := mapstructure.Decode(cpacket.Item[i].Body, &resp); err != nil {
            t.Logf("convert_map_struct, error:%#v\n", err)
        }
        t.Logf("struct.Item[%d].Cmd:%s\n", i, cpacket.Items[i].Cmd)
        t.Logf("struct.Item[%d].Body.A:%d\n", i, resp.A)
        t.Logf("struct.Item[%d].Body.B:%s\n", i, resp.B)
    }
}

func make_item(cmd string, a int, b string) PushPacketItem {
    return PushPacketItem {
        Cmd:    cmd,
        Body:   TestResp {A: 1, B: "hello"},
    }
}