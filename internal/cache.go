package internal

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/garyburd/redigo/redis"
)

// ConnManager is user type of redis Pool
type ConnManager struct {
	redisConn *redis.Pool
}

// NewConnManager is return new connection to redis Pool
func NewConnManager(conn *redis.Pool) *ConnManager {
	return &ConnManager{
		redisConn: conn,
	}
}

// Check is check key in redis
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

// Create is save key data to redis
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
