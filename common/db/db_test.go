package db_test

import (
	"v2ray.com/core/common/db"
	"v2ray.com/core/common/db/model"
	"testing"
)

func TestDBConnection(t *testing.T) {
	pool := db.New()
	pool.Start("tcp", "localhost", "6379")
	status1 := model.URLStatus{"example.com", 0}
	pool.InsertRecord(status1)
	status2, _ := pool.LookupRecord("example.com")
	if (status1 != status2) {
		t.Error("DB record doesn't match", status1, status2)
	}
}
