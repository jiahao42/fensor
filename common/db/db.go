package db

//go:generate errorgen

import (
	"github.com/gomodule/redigo/redis"
	"v2ray.com/core/common/db/model"
	"time"
)


type Pool struct {
	pool *redis.Pool
}

func (p *Pool) Start(protocol, ip, port string) {
	p.pool = &redis.Pool {
		MaxIdle: 10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial(protocol, ip + ":" + port)
		},
	}
}

func (p *Pool) GetConn() (redis.Conn, error) {
	conn := p.pool.Get()
	if conn.Err() != nil {
		return nil, conn.Err()
	}
	return conn, nil
}

func (p *Pool) LookupRecord (URL string) (*model.URLStatus, error) {
	conn, err := p.GetConn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	status := new(model.URLStatus)
	values, err := redis.Values(conn.Do("HGETALL", URL))
	err = redis.ScanStruct(values, status)
  //newDebugMsg("DB: lookup for " + URL + ": " + StructString(status))
	if err != nil {
		return nil, err
	}
	return status, nil
}

func (p *Pool) InsertRecord (status *model.URLStatus) (error) {
	conn, err := p.GetConn()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Do("HMSET", status.URL, "URL", status.URL, "Status", status.Status)
  //newDebugMsg("DB: inserting for " + status.URL + ": " + StructString(status))
	return err
}

func New() *Pool {
	return &Pool{}
}

