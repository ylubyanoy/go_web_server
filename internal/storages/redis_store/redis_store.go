package redis_store

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/ylubyanoy/go_web_server/internal/models"

	"github.com/garyburd/redigo/redis"
)

// ConnManager - user type of redis Pool
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
				return nil, err
			}
			return redisConn, nil
		},
	}
	rc := redisConn.Get()
	_, err := redis.String(rc.Do("PING"))
	if err != nil {
		return nil, fmt.Errorf("can't PING to Redis: %w", err)
	}
	rc.Close()

	sm := &ConnManager{
		redisConn: redisConn,
	}

	return sm, nil
}

// Check - check key in redis
func (sm *ConnManager) Check(streamerName string) *models.StreamerInfo {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	mkey := streamerName
	data, err := redis.Bytes(cmc.Do("GET", mkey))
	if err != nil {
		log.Printf("can't get data for %s: (%s)", mkey, err)
		return nil
	}
	si := &models.StreamerInfo{}
	err = json.Unmarshal(data, si)
	if err != nil {
		log.Printf("can't unpack data for %s: (%s)", mkey, err)
		return nil
	}
	return si
}

// Create - save key data to redis
func (sm *ConnManager) Create(si *models.StreamerInfo, stime int) error {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	dataSerialized, _ := json.Marshal(si)
	mkey := si.ChannelName
	data, err := cmc.Do("SET", mkey, dataSerialized, "EX", stime)
	result, err := redis.String(data, err)
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("result not OK")
	}
	return nil
}

// CreateToken - save token to redis
func (sm *ConnManager) CreateToken(tvalue string, stime int) error {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	tkey := "token"
	data, err := cmc.Do("SET", tkey, tvalue, "EX", stime)
	result, err := redis.String(data, err)
	if err != nil {
		return err
	}
	if result != "OK" {
		return fmt.Errorf("result not OK")
	}
	return nil
}

// CheckToken - check key in redis
func (sm *ConnManager) CheckToken(streamerName string) string {
	cmc := sm.redisConn.Get()
	defer cmc.Close()

	tkey := "token"
	token, err := redis.Bytes(cmc.Do("GET", tkey))
	if err != nil {
		log.Printf("can't get token (%s)", err)
		return ""
	}

	return string(token)
}
