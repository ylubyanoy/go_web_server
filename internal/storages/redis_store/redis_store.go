package redis_store

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/models"

	"github.com/garyburd/redigo/redis"
)

// ConnManager is user type of redis Pool
type ConnManager struct {
	redisConn *redis.Pool
}

// New is return new connection to redis Pool
func New(redisAddr string) (*ConnManager, error) {
	redisConn := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			redisConn, err := redis.DialURL(redisAddr)
			if err != nil {
				return nil, fmt.Errorf("Can't connect to Redis: %w", err)
			}
			return redisConn, nil
		},
	}
	rc := redisConn.Get()
	_, err := redis.String(rc.Do("PING"))
	if err != nil {
		return nil, fmt.Errorf("Can't connect to Redis: %w", err)
	}
	rc.Close()

	sm := &ConnManager{
		redisConn: redisConn,
	}

	return sm, nil
}

// Check is check key in redis
func (sm *ConnManager) Check(streamerName string) *models.StreamerInfo {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	mkey := streamerName
	data, err := redis.Bytes(cmc.Do("GET", mkey))
	if err != nil {
		log.Printf("cant get data for %s: (%s)", mkey, err)
		return nil
	}
	si := &models.StreamerInfo{}
	err = json.Unmarshal(data, si)
	if err != nil {
		log.Printf("cant unpack data for %s: (%s)", mkey, err)
		return nil
	}
	return si
}

// Create is save key data to redis
func (sm *ConnManager) Create(si *models.StreamerInfo) error {
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
