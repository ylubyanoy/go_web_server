package internal

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

type ConnManager struct {
	redisConn *redis.Pool
}

func NewConnManager(conn *redis.Pool) *ConnManager {
	return &ConnManager{
		redisConn: conn,
	}
}

func (sm *ConnManager) Check(streamerName string) *StreamerInfo {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	mkey := streamerName
	data, err := redis.Bytes(cmc.Do("GET", mkey))
	if err != nil {
		log.Printf("cant get data for %s: (%s)", mkey, err)
		return nil
	}
	si := &StreamerInfo{}
	err = json.Unmarshal(data, si)
	if err != nil {
		log.Printf("cant unpack data for %s: (%s)", mkey, err)
		return nil
	}
	return si
}

func (sm *ConnManager) Create(si *StreamerInfo) error {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	dataSerialized, _ := json.Marshal(si)
	mkey := si.ChannelName
	data, err := cmc.Do("SET", mkey, dataSerialized, "EX", 60*10)
	result, err := redis.String(data, err)
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("result not OK")
	}
	return nil
}
