package redisedge

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/gomodule/redigo/redis"
)

// RedisEdge represents a RedisEdge instance
type RedisEdge struct {
	lc   logger.LoggingClient
	dc   map[string]string
	pool *redis.Pool
}

// Initialize returns a connection pool to a RedisEdge server from a URL
func Initialize(dc map[string]string, lc logger.LoggingClient) (re *RedisEdge, err error) {
	re = &RedisEdge{
		lc: lc,
		dc: dc,
		pool: &redis.Pool{
			MaxIdle:     3,
			IdleTimeout: 240 * time.Second,
			Dial:        func() (redis.Conn, error) { return redis.DialURL(dc["RedisURL"]) },
		},
	}
	lc.Debug("redisedge.Initialize() started")

	if err = re.checkModules(); err != nil {
		return re, err
	}
	if err = re.initRedisAIModel(); err != nil {
		return re, err
	}
	if err = re.initRedisAIScript(); err != nil {
		return re, err
	}

	lc.Debug("redisedge.Initialize() exited")
	return re, nil
}

func (re *RedisEdge) checkModules() error {
	// Verify RedisEdge stack modules (module_name:min_version) TODO: move to config
	re.lc.Debug("redisedge.checkModules() started")

	conn := re.pool.Get()

	reqmods := map[string]int{
		"rg": 301,
		"ai": 200,
	}
	mods, err := redis.Values(conn.Do("MODULE", "LIST"))
	if err != nil {
		re.lc.Error(fmt.Sprintf("redisedge.checkModules: Failed getting modules list - %v", err))
		return err
	}
	for _, mod := range mods {
		m, _ := redis.Values(mod, nil)
		n, _ := redis.String(m[1], nil)
		v, _ := redis.Int(m[3], nil)
		_, ok := reqmods[n]
		if ok && v >= reqmods[n] {
			delete(reqmods, n)
		}
		if len(reqmods) == 0 {
			break
		}
	}
	if len(reqmods) > 0 {
		for k, v := range reqmods {
			re.lc.Error(fmt.Sprintf("redisedge.checkModules: Missing requisite RedisEdge module - %s v%d or greater", k, v))
		}
		return fmt.Errorf("Missing requisite RedisEdge module(s)")
	}

	re.lc.Debug("redisedge.checkModules() exited")
	return nil
}

func (re *RedisEdge) initRedisAIModel() error {
	// Initializes the RedisAI model
	re.lc.Debug("redisedge.initRedisAIModel() started")

	conn := re.pool.Get()

	raiModelKey := re.dc["RAIModelKey"]
	raiModelKeyExists, err := redis.Bool(conn.Do("EXISTS", raiModelKey))
	if err != nil {
		re.lc.Error(fmt.Sprintf("redisedge.initRedisAIModel: Failed checking for RedisAI model key - %v", err))
		return err
	}

	if !raiModelKeyExists {
		raiModelPath := re.dc["RAIModelPath"]
		raiModelBackend := re.dc["RAIModelBackend"]
		raiModelDevice := re.dc["RAIModelDevice"]
		raiModelBlob, err := ioutil.ReadFile(raiModelPath)
		if err != nil {
			re.lc.Error(fmt.Sprintf("redisedge.initRedisAIModel: Failed reading RedisAI model file - %v", err))
			return err
		}
		rep, err := redis.String(conn.Do("AI.MODELSET", raiModelKey, raiModelBackend, raiModelDevice, "INPUTS", "input", "OUTPUTS", "output", raiModelBlob))
		if err != nil {
			re.lc.Error(fmt.Sprintf("redisedge.initRedisAIModel: Failed setting AI model key - %v", err))
			return err
		}
		if rep != "OK" {
			err = fmt.Errorf("Unexpected reply when setting AI model key - 'OK' != %s", rep)
			re.lc.Error(fmt.Sprintf("redisedge.initRedisAIModel: %v", err))
			return err
		}
		re.lc.Debug("redisedge.initRedisAIModel(): Created RedisAI model key")
	} else {
		re.lc.Debug("redisedge.initRedisAIModel(): Existing RedisAI model key found")
	}

	re.lc.Debug("redisedge.initRedisAIModel() exited")
	return nil
}

func (re *RedisEdge) initRedisAIScript() error {
	// Initializes the RedisAI script
	re.lc.Debug("redisedge.initRedisAIScript() started")

	conn := re.pool.Get()

	raiScriptKey := re.dc["RAIScriptlKey"]
	raiScriptKeyExists, err := redis.Bool(conn.Do("EXISTS", raiScriptKey))
	if err != nil {
		re.lc.Error(fmt.Sprintf("redisedge.initRedisAIScript: Failed checking for RedisAI script key - %v", err))
		return err
	}

	if !raiScriptKeyExists {
		raiScriptPath := re.dc["RAIScriptPath"]
		raiScriptDevice := re.dc["RAIScriptDevice"]
		raiScriptBlob, err := ioutil.ReadFile(raiScriptPath)
		if err != nil {
			re.lc.Error(fmt.Sprintf("redisedge.initRedisAIScript: Failed reading RedisAI script file - %v", err))
			return err
		}
		rep, err := redis.String(conn.Do("AI.SCRIPTSET", raiScriptKey, raiScriptDevice, raiScriptBlob))
		if err != nil {
			re.lc.Error(fmt.Sprintf("redisedge.initRedisAIScript: Failed setting AI script key - %v", err))
			return err
		}
		if rep != "OK" {
			err = fmt.Errorf("Unexpected reply when setting AI script key - 'OK' != %s", rep)
			re.lc.Error(fmt.Sprintf("redisedge.initRedisAIScript: %v", err))
			return err
		}
		re.lc.Debug("redisedge.initRedisAIScript(): Created RedisAI script key")
	} else {
		re.lc.Debug("redisedge.initRedisAIScript(): Existing RedisAI script key found")
	}

	re.lc.Debug("redisedge.initRedisAIScript() exited")
	return nil
}

// YOLODetect detects people and dogs in a frame
func (re *RedisEdge) YOLODetect(buf []byte) (hoomans uint64, doggos uint64, err error) {
	// TODO: resize
	// TODO: normalize
	// TODO: prepare tensor
	// TODO: run model
	// TODO: run script
	// TODO: extract boxes
	return 0, 0, nil
}
