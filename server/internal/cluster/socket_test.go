package cluster_test

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"google.golang.org/protobuf/proto"

	pb "github.com/Alievs-corp/hypedomino/server/internal/pb/hypedomino/v1"
)

func websocketDial(ctx context.Context, url string, header http.Header) (*websocket.Conn, *http.Response, error) {
	return websocket.Dial(ctx, url, &websocket.DialOptions{HTTPHeader: header})
}

func writeFrame(ctx context.Context, socket *websocket.Conn, frame *pb.Envelope) error {
	data, err := proto.Marshal(frame)
	if err != nil {
		return err
	}
	return socket.Write(ctx, websocket.MessageBinary, data)
}

func readFrame(ctx context.Context, socket *websocket.Conn) (*pb.Envelope, error) {
	_, data, err := socket.Read(ctx)
	if err != nil {
		return nil, err
	}
	frame := &pb.Envelope{}
	return frame, proto.Unmarshal(data, frame)
}
