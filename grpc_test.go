package grkv

import (
	"context"
	"net"
	"testing"

	"github.com/msonawane/grkv/kvpb"
	"google.golang.org/grpc"
)

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func TestGRPCSet(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := kvpb.NewKeyValueStoreClient(conn)

	req := &kvpb.SetRequest{}
	for _, d := range testData {
		req.Data = append(req.Data, &kvpb.KeyValue{
			Key:   []byte(d.key),
			Value: []byte(d.value),
		})
	}

	success, err := client.GRPCSet(ctx, req)
	if err != nil {
		t.Error(err)
	}
	if !success.Success {
		t.Errorf("set failed")
	}

}

func TestGRPCGet(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := kvpb.NewKeyValueStoreClient(conn)

	keys := [][]byte{}
	for _, tt := range testData {
		keys = append(keys, []byte(tt.key))
		// fmt.Println(tt.key)
	}
	gr := &kvpb.GetRequest{Keys: keys}

	resp, err := client.GRPCGet(ctx, gr)
	if err != nil {
		t.Error(err)
	}
	if len(resp.Data) != len(keys) {
		t.Errorf("response.Data size should match keys size")
	}
}

func TestGRPCDelete(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	defer conn.Close()
	client := kvpb.NewKeyValueStoreClient(conn)

	keys := [][]byte{}
	for _, tt := range testData {
		keys = append(keys, []byte(tt.key))
	}
	dr := &kvpb.DeleteRequest{Keys: keys}
	success, err := client.GRPCDelete(ctx, dr)
	if err != nil {
		t.Error(err)
	}
	if !success.Success {
		t.Errorf("delete failed")
	}
}
