package client

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/shadow1ng/fscan/common"
	"github.com/shadow1ng/fscan/config"
	"github.com/tomatome/grdp/glog"
	"time"
)

type RedisContext struct {
	RedisClient *redis.ClusterClient
}

func InitRedis(redisConfig config.Redis) {
	//glog.SetLevel(glog.INFO)
	//logger := log.New(os.Stdout, "", 0)
	//glog.SetLogger(logger)
	common.Context.Redis.RedisClient = redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    redisConfig.Nodes,
		Password: redisConfig.Password,
	})
}

func (redisContext *RedisContext) SetKey(k, v string, ex interface{}) (bool, error) {
	ctx := context.Background()
	_, err := redisContext.RedisClient.Set(ctx, k, v, getExpire(ex)).Result()
	if err != nil {
		glog.Errorf("redis set %s nx error : %s", k, err.Error())
		return false, err
	}
	return true, nil
}

func (redisContext *RedisContext) GetKey(k string) (string, error) {
	resp, err := redisContext.RedisClient.Get(context.Background(), k).Result()
	if err != nil {
		glog.Errorf("redis GetKey %s nx error : %s", k, err.Error())
		return "", err
	}
	return resp, nil
}

func (redisContext *RedisContext) HGetAll(k string) (map[string]string, error) {
	resp, err := redisContext.RedisClient.HGetAll(context.Background(), k).Result()
	if err != nil {
		glog.Errorf("redis HGetAll %s error : %s", k, err.Error())
		return nil, err
	}
	return resp, nil
}

func (redisContext *RedisContext) HExists(k, v string) (bool, error) {
	_, err := redisContext.RedisClient.HExists(context.Background(), k, v).Result()
	if err != nil {
		glog.Errorf("redis HExists %s error : %s", k, err.Error())
		return false, err
	}
	return true, nil
}

func (redisContext *RedisContext) SAdd(k, v string) (bool, error) {
	_, err := redisContext.RedisClient.SAdd(context.Background(), k, v).Result()
	if err != nil {
		glog.Errorf("redis SAdd %s error : %s", k, err.Error())
		return false, err
	}
	return true, nil
}

func (redisContext *RedisContext) Del(keys ...string) (bool, error) {
	_, err := redisContext.RedisClient.Del(context.Background(), keys...).Result()
	if err != nil {
		glog.Errorf("redis Del %s error : %s", keys, err.Error())
		return false, err
	}
	return true, nil
}

func getExpire(ex interface{}) time.Duration {
	if ex == nil {
		return time.Duration(0) * time.Second
	}
	return time.Duration(ex.(int)) * time.Second
}

//
//func HSetKey(h, k, v string, ex interface{}) error {
//	rs := RedisClient.Get()
//	if rs == nil {
//		glog.Error("redis session get error, by pass .")
//		return nil
//	}
//	defer rs.Close()
//	var err error
//	_, err = redis.Int64(rs.Do("HSET", h, k, v))
//	if err != nil {
//		glog.Errorf("redis set %s nx error : %s", h, err.Error())
//		return err
//	}
//	if ex != nil {
//		_, err = redis.Int64(rs.Do("EXPIRE", h, ex.(int)))
//	}
//	if err != nil {
//		glog.Errorf("redis hset %s nx error : %s", h, err.Error())
//		return err
//	}
//	return nil
//}
//
//func HGetKey(h, k string) (string, error) {
//	rs := RedisClient.Get()
//	if rs == nil {
//		glog.Error("redis session get error, by pass .")
//		return "unknown", errors.New("redis session get error")
//	}
//	defer rs.Close()
//	var resp string
//	var err error
//	resp, err = redis.String(rs.Do("HGET", h, k))
//	if err != nil {
//		if err == redis.ErrNil {
//			return "Nil", errors.New("nil error")
//		} else {
//			glog.Errorf("redis hget %s nx error : %s", h, err.Error())
//			return "unknown", err
//		}
//	}
//	return resp, nil
//}
//
//func GetKey(k string) (string, error) {
//	rs := RedisClient.Get()
//	defer rs.Close()
//	resp, err := redis.String(rs.Do("GET", k))
//	if err != nil {
//		glog.Errorf("redis get %s nx error : %s", k, err.Error())
//		return "", err
//	}
//	return resp, nil
//}
//
//func ExistKey(k string) (bool, error) {
//	rs := RedisClient.Get()
//	defer rs.Close()
//	rrsp, _ := redis.Bool(rs.Do("EXISTS", k))
//	return rrsp, nil
//}
//
//func ScanKeys(maxCount int) ([]interface{}, error) {
//	rs := RedisClient.Get()
//	defer rs.Close()
//	resp, err := redis.Values(rs.Do("SCAN", 0, "MATCH", "*", "COUNT", maxCount))
//	if err != nil {
//		glog.Errorf("redis scan keys error : %s", err.Error())
//		return nil, err
//	}
//	return resp, nil
//}
//
//func DelKey(k string) (string, error) {
//	rs := RedisClient.Get()
//	defer rs.Close()
//	rrsp, _ := redis.String(rs.Do("DEL", k))
//	return rrsp, nil
//}
//
//func ExpireKey(k string, ex int) (bool, error) {
//	rs := RedisClient.Get()
//	if rs == nil {
//		glog.Error("redis session get error, by pass .")
//		return false, nil
//	}
//	defer rs.Close()
//	var resp string
//	var err error
//	resp, err = redis.String(rs.Do("EXPIRE", k, strconv.Itoa(ex)))
//	if err != nil {
//		glog.Errorf("redis expire %s nx error : %s", k, err.Error())
//		return false, err
//	}
//	if resp != "OK" {
//		glog.Warn("redis of key %s set resp %s.", k, resp)
//		return false, err
//	}
//	return true, nil
//}
//
//func SAddKey(s, k string, ex int) (bool, error) {
//	rs := RedisClient.Get()
//	if rs == nil {
//		glog.Error("redis session get error, by pass .")
//		return false, nil
//	}
//	defer rs.Close()
//	var rrsp int64
//	var err error
//	rrsp, err = redis.Int64(rs.Do("SADD", s, k))
//	if err != nil {
//		glog.Errorf("redis sadd %s nx error : %s", k, err.Error())
//		return false, err
//	}
//	if rrsp == 0 {
//		glog.Warn("redis of key %s sadd resp %d.", k, rrsp)
//		return false, err
//	}
//	resp, err := redis.Int64(rs.Do("EXPIRE", s, strconv.Itoa(ex)))
//	if err != nil {
//		glog.Errorf("redis expire %s nx error : %s", s, err.Error())
//		return false, err
//	}
//	if resp == 0 {
//		glog.Warn("redis of key %s set resp %d.", s, resp)
//		return false, err
//	}
//	return true, nil
//}
