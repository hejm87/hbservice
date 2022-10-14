//package db_test
package hbservice_test

import (
	"fmt"
	"testing"
	"gopkg.in/redis.v5"
)

func TestRedis(t *testing.T) {
	client := redis.NewClient(&redis.Options {
		Addr: 		"localhost:6379",
		Password:	"",
		DB:			0,
	})
	pong, err := client.Ping().Result()
	fmt.Println(pong, err)

	err = client.Set("key", "value", 0).Err()
	if err != nil {
		t.Fatalf("redis.Set error:%#v", err)
	}

	var value string
	value, err = client.Get("key").Result()
	if err == redis.Nil {
		t.Fatalf("redis.Get not exists value")
	} else if err != nil {
		t.Fatalf("redis.Get error:%#v", err)
	} else {
		if value != "value" {
			t.Fatalf("redis.Get value not match")
		}
	}
}