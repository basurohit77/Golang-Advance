package common

import (
	"github.com/go-redis/redis/v7"
	"log"
	"os"
	"time"
)

func NewRedisClient() (*redis.Client, error){
	uri := os.Getenv("REDIS_URI")

	//handle rediss uri statement
	if uri != "" {
		opt, err := redis.ParseURL(uri)
		if err != nil {
			log.Println("parse url failed: ", err)
			return nil, err
		}
		opt.DB = 15
		opt.DialTimeout = 10 * time.Second
		opt.ReadTimeout = 30 * time.Second
		opt.WriteTimeout = 30 * time.Second
		opt.PoolSize = 10
		opt.PoolTimeout = 30 * time.Second
		opt.IdleTimeout = time.Minute
		opt.IdleCheckFrequency = 100 * time.Millisecond
		return redis.NewClient(opt), nil
	}

	opt := redisOptions()
	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	opt.Addr = redisHost + ":" + redisPort
	opt.Password = redisPassword
	return redis.NewClient(opt), nil
}

func redisOptions() *redis.Options {
	return &redis.Options{
		DB:                 15,
		DialTimeout:        10 * time.Second,
		ReadTimeout:        30 * time.Second,
		WriteTimeout:       30 * time.Second,
		PoolSize:           10,
		PoolTimeout:        30 * time.Second,
		IdleTimeout:        time.Minute,
		IdleCheckFrequency: 100 * time.Millisecond,
	}
}

func Sub(pubsub *redis.PubSub, persist bool, f func(string)) error{

	// Wait for confirmation that subscription is created before publishing anything.
	_, err := pubsub.Receive()
	if err != nil {
		log.Println("subscription failed: ", err)
		return err
	}

	// Go channel which receives messages.
	ch := pubsub.Channel()

	// Consume messages.
	for msg := range ch {
		log.Println(msg.Channel, msg.Payload)
		f(msg.Payload)
		if !persist {
			pubsub.Close()
			break
		}
	}

	return nil
}

func Pub(client *redis.Client, channel string, payload string) error {

	log.Println("pub message: ", channel, payload)

	_, err := client.Publish(channel, payload).Result()
	if err != nil {
		log.Println("publish message failed: ", err)
		return err
	}

	return nil
}
