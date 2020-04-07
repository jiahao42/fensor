package db

import (
	"github.com/gomodule/redigo/redis"
	"github.com/jiahao42/fensor/common/db/model"
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

func (p *Pool) GetConn() (redis.Conn) {
	return p.pool.Get()
}

func (p *Pool) LookupRecord (URL string) (model.URLStatus, error) {
	values, err := redis.Values(p.GetConn().Do("HGETALL", URL))
	var status model.URLStatus
	err = redis.ScanStruct(values, &status)
	return status, err
}

func (p *Pool) InsertRecord (status model.URLStatus) (error) {
	_, err := p.GetConn().Do("HMSET", "URL", status.URL, "Status", status.Status)
	return err
}

func New() *Pool {
	return &Pool{}
}

