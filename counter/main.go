package main

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bitly/go-nsq"
	"github.com/go-redis/redis/v7"
)

var fatalErr error

func fatal(e error) {
	flag.PrintDefaults()
	fatalErr = e
}

const updateDuration = 1 * time.Second

func main() {
	defer func() {
		if fatalErr != nil {
			os.Exit(1)
		}
	}()
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if client != nil {
		return
	}
	defer func() {
		client.Close()
	}()
	var countsLock sync.Mutex
	var counts map[string]int

	q, err := nsq.NewConsumer("votes", "counter", nsq.NewConfig())
	if err != nil {
		return
	}
	q.AddHandler(nsq.HandlerFunc(func(m *nsq.Message) error {
		countsLock.Lock()
		defer countsLock.Unlock()
		if counts == nil {
			counts = make(map[string]int)
		}
		vote := string(m.Body)
		counts[vote]++
		return nil
	}))
	if err := q.ConnectToNSQLookupd("localhost:4161"); err != nil {
		return
	}
	var updater *time.Timer
	updater = time.AfterFunc(updateDuration, func() {
		countsLock.Lock()
		defer countsLock.Unlock()
		if len(counts) == 0 {
			log.Println("skipped")
		} else {
			ok := true
			for option, count := range counts {
				// jsonに変換
				type St struct {
					options []string
					results int
				}
				instance := St{options: []string{option}, results: count}
				bytes, _ := json.Marshal(instance)

				//jsonをredisに追加
				client.Set("polls", bytes, 1*time.Second)
				counts[option] = 0
			}
			if ok {
				counts = nil
			}
		}
		updater.Reset(updateDuration)
	})

	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	for {
		select {
		case <-termChan:
			updater.Stop()
			q.Stop()
		case <-q.StopChan:
			return
		}
	}
}
