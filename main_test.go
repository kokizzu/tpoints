package main_test

import (
	"fmt"
	"github.com/kokizzu/gotro/I"
	"github.com/kokizzu/gotro/L"
	"github.com/kokizzu/gotro/T"
	"math/rand"
	"net/http"
	"testing"
	"tpoints/handler"
	"tpoints/model"
)

func TestSync(t *testing.T) {
	// prepare mongodb 
	server := handler.InitServer()
	defer server.Close()
	userId := `1`
	err := server.Points.DeleteAll_ByUser(userId)
	if err != nil {
		t.Error(err)
		return
	}
	tables := []struct {
		past  int64
		delta int64
		sum   int64
	}{
		{0, 1, 1},
		{1, 1, 2},
		{2, 2, 4},
		{4, 2, 6},
		{6, 2, 8},
		{8, 1, 9},
	}
	for _, table := range tables {
		p := model.Point{
			QueueAt: T.UnixNano(),
			Delta:   table.delta,
			UserId:  userId,
		}
		err := server.Points.Process(&p)
		if err != nil {
			t.Error(err)
			return
		}
		if p.Sum != table.sum {
			t.Errorf("Sum of (%d+%d) was incorrect, got: %d, want: %d.", table.past, table.delta, p.Sum, table.sum)
		}
	}
}

const URL_PREFIX = `http://127.0.0.1:1323`
const URL_SYNC = `/points/add`
const URL_ASYNC = `/points/queue`
const MAX_USER = 100
const MAX_POINT = 30

func BenchmarkTeardown(b *testing.B) {
	server := handler.InitServer()
	defer server.Close()
	for i := 0; i < MAX_USER; i++ {
		server.Points.DeleteAll_ByUser(I.ToStr(i))
	}
}

func BenchmarkSync(b *testing.B) {
	for n := 0; n < b.N; n++ {
		userId := rand.Intn(MAX_USER)
		point := 1 + rand.Intn(MAX_POINT)
		doPost(fmt.Sprintf(`%s%s/%d/%d`, URL_PREFIX, URL_SYNC, userId, point))
	}
}

func BenchmarkAsync(b *testing.B) {
	for n := 0; n < b.N; n++ {
		userId := rand.Intn(MAX_USER)
		point := 1 + rand.Intn(MAX_POINT)
		doPost(fmt.Sprintf(`%s%s/%d/%d`, URL_PREFIX, URL_ASYNC, userId, point))
	}
}

func doPost(url string) {
	resp, err := http.Post(url,``,nil)
	if L.IsError(err,`failed http.post %s`,url) {
		return
	}
	defer resp.Body.Close()
}
