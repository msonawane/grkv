package grkv

import (
	"context"
	"testing"
	"time"

	fuzz "github.com/google/gofuzz"
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

	time.Sleep(10 * time.Second)

	ctx := context.Background()
	t.Run("test for nil data", func(t *testing.T) {
		req := &kvpb.SetRequest{}
		req.Data = make([]*kvpb.KeyValue, 3)
		success, err := store.Set(ctx, req)
		if err != nil {
			t.Error(err)
		}
		if !success.Success {
			t.Errorf("set failed")
		}
	})
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

	t.Run("zcheck replication", func(t *testing.T) {
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
		keys := [][]byte{}
		for _, tt := range testData {
			keys = append(keys, []byte(tt.key))
			// fmt.Println(tt.key)
		}
		keys = append(keys, []byte("not-in-db"))
		gr := &kvpb.GetRequest{Keys: keys}

		resp, err := store2.Get(ctx, gr)
		if err != nil {
			t.Error(err)
		}
		if len(resp.Data) != len(keys)-1 {
			t.Errorf("response.Data size should match keys size, expected %d, got: %d", len(keys)-1, len(resp.Data))
		}

		if len(resp.KeysNotFound) != 1 {
			t.Errorf("response.KeysNotFound should have %d keys, got: %d", 1, len(resp.KeysNotFound))
		}
		if string(resp.KeysNotFound[0]) != "not-in-db" {
			t.Errorf("keys not found should have not-in-db; got %s", resp.KeysNotFound[0])
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
		// if bytes.Compare(resp.KeysNotFound[0], []byte("not-in-db")) == 0 {
		// 	t.Errorf("keys not found should have not-in-db; got %s", resp.KeysNotFound[0])
		// }
		if string(resp.KeysNotFound[0]) != "not-in-db" {
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
		t.Errorf("delete failed")
	}

}

func TestFuzzingSet(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		sr := kvpb.SetRequest{}
		f := fuzz.New().NilChance(0.5)
		f.Fuzz(&sr)
		// t.Logf("sr: %#v\n", &sr)

		success, err := store.set(ctx, &sr)
		if err != nil {
			t.Errorf("set failed, err: %v", err)
		}
		if !success.Success {
			t.Errorf("set failed, success: %v", success)
		}

	}
}

func TestFuzzingGet(t *testing.T) {
	ctx := context.Background()
	for i := 0; i < 1000; i++ {
		gr := kvpb.GetRequest{}
		f := fuzz.New().NilChance(0.5)
		f.Fuzz(&gr)
		// t.Logf("gr: %#v\n", &gr)

		_, err := store.get(ctx, &gr)
		if err != nil {
			t.Errorf("set failed, err: %v", err)
		}

	}
}
