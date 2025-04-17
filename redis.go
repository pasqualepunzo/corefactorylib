package lib

import (
	"encoding/json"
	"os"
	"time"

	"github.com/go-redis/redis"
)

func DoSetRedis(key string, val interface{}) error {

	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("redisHost") + ":" + os.Getenv("redisPort"), // Redis address
		Password: os.Getenv("redisPass"),                                // No password set
		DB:       0,                                                     // Database di default
	})

	// Serializza la struct in JSON
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	// Scrivi su Redis
	err = client.Set(key, b, 24*time.Hour).Err()
	if err != nil {
		return err
	}

	return nil
}
func DoGetRedis(key string) (string, error) {

	client := redis.NewClient(&redis.Options{
		Addr:     "redis-master.redis.svc.cluster.local:6379", // Indirizzo Redis
		Password: "v10l4",                                     // Nessuna password per default
		DB:       0,                                           // Database di default
	})

	// Lettura del dato da Redis
	val, err := client.Get(key).Result()
	if err != nil {
		return "", err
	}

	return val, nil
}
