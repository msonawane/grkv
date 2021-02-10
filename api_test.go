package grkv_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/msonawane/grkv/kvpb"
)

var testData = []struct {
	key   string
	value string
}{
	{"a", "a"},
	{"b", "b"},
}

func TestSet(t *testing.T) {
	ctx := context.Background()
	for _, tt := range testData {
		t.Run(tt.key, func(t *testing.T) {
			req := kvpb.SetRequest{
				Data: []*kvpb.KeyValue{
					{
						Key:   []byte(tt.key),
						Value: []byte(tt.value),
					}},
			}
			success, err := store.Set(ctx, &req)
			if err != nil {
				t.Error(err)
			}
			if !success.Success {
				t.Errorf("set did not success")
			}
		})
	}

	t.Run("MultiSet", func(t *testing.T) {
		req := &kvpb.SetRequest{}
		for _, d := range testData {
			req.Data = append(req.Data, &kvpb.KeyValue{
				Key:   []byte(d.key),
				Value: []byte(d.value),
			})
		}

		success, err := store.Set(ctx, req)
		if err != nil {
			t.Error(err)
		}
		if !success.Success {
			t.Errorf("set did not succied")
		}
	})

}

func TestGet(t *testing.T) {
	ctx := context.Background()
	for _, tt := range testData {
		t.Run(tt.key, func(t *testing.T) {
			keys := [][]byte{[]byte(tt.key)}
			gr := &kvpb.GetRequest{Keys: keys}
			resp, err := store.Get(ctx, gr)
			if err != nil {
				t.Error(err)
			}
			if len(resp.KeysNotFound) > 0 {
				t.Errorf("key %s not found", tt.key)
			}
		})
	}

	t.Run("MultiGet", func(t *testing.T) {
		keys := [][]byte{}
		for _, tt := range testData {
			keys = append(keys, []byte(tt.key))
			// fmt.Println(tt.key)
		}
		keys = append(keys, []byte("not-in-db"))
		gr := &kvpb.GetRequest{Keys: keys}

		resp, err := store.Get(ctx, gr)
		if err != nil {
			t.Error(err)
		}
		if len(resp.Data) != len(keys)-1 {
			t.Errorf("response.Data size should match keys size")
		}

		if len(resp.KeysNotFound) != 1 {
			t.Errorf("response.KeysNotFound should have 1 key")
		}
		if bytes.Compare(resp.KeysNotFound[0], []byte("not-in-db")) != 0 {
			t.Errorf("keys not found should have not-in-db; got %s", resp.KeysNotFound[0])
		}
	})

}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	keys := [][]byte{}
	for _, tt := range testData {
		keys = append(keys, []byte(tt.key))
	}
	dr := &kvpb.DeleteRequest{Keys: keys}
	success, err := store.Delete(ctx, dr)
	if err != nil {
		t.Error(err)
	}
	if !success.Success {
		t.Errorf("delete did not succied")
	}

}
