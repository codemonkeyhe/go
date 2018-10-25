package main

import (
	"fmt"

	. "demo/protoDemo/proto"

	"github.com/golang/protobuf/proto"
)

func main() {
	fmt.Println("HI")
	var data string

	var channelStreamInfo *ChannelStreamInfo = &ChannelStreamInfo{
		Version: 1520934201175804449,
		Streams: make([]*StreamInfo, 0),
	}
	var stream *StreamInfo = &StreamInfo{
		StreamName: "xa_235689_235689_0_0_0",
		Json:       "{\"channel\":2,\"encoderType\":1,\"is_trans\":0,\"rate\":40,\"sample_rate\":44100}",
	}
	channelStreamInfo.Streams = append(channelStreamInfo.Streams, stream)

	var avpPayload *AvpPayload = &AvpPayload{
		Addr:        []byte("http://www.yy.com"),
		SslAddr:     []byte("http://www.yy.com"),
		StreamNames: []string{"xv_235689_235689_0_0_0"},
	}

	var liveinfo *LiveInfo = &LiveInfo{
		//LiveKey需要从请求方获得,最好透传给媒体中心.
		//LiveKey:           queryReq.LiveKeys[0],
		LiveKey:           &LiveKey{},
		ChannelStreamInfo: channelStreamInfo,
		AvpPayload:        avpPayload,
	}

	var queryResp ChannelStreamsQueryResp
	queryResp.ResCode = uint32(1024)
	queryResp.ResMsg = "OK"
	//queryResp.Sequence = queryReq.Sequence
	queryResp.LiveInfos = append(queryResp.LiveInfos, liveinfo)

	//pb结构体序列化为[]byte
	buf, err := proto.Marshal(&queryResp)
	if err != nil {
		fmt.Printf("proto.Marshal Err: %v", err)
		return
	}
	data = string(buf)
	//从data反解pb结构体
	tmp := &ChannelStreamsQueryResp{}
	if err = proto.Unmarshal([]byte(data), tmp); err != nil {
		fmt.Printf("pb unmarshal err:%s", err)
	} else {
		fmt.Printf("pb marshal success!\n PTR  Origin Data: %v \n NEW DATA: %v", &queryResp, tmp)
		fmt.Printf("pb marshal success!\n OBJ  Origin Data: %v \n NEW DATA: %v", queryResp, *tmp)
	}

}
