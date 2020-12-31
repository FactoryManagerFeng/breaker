package main

import (
	"code.piupiu.com/book/go_book_common/breaker"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"
)

var b *breaker.Breaker

func init() {
	var settings breaker.Settings
	settings.Name = "test_breaker"
	settings.ReadyToTrip = func(counts breaker.Counts) bool {
		if counts.RequestsNum >= 10 {
			return true
		}
		return false
	}
	settings.OnStateChange = func(name string, from, to breaker.State) {
		fmt.Println(name, from, to)
	}
	b = breaker.NewBreaker(settings)
}

func main() {
	var url = "https://api-vip.qschou.com"
	var wg sync.WaitGroup
	for i := 0; i < 12; i++ {
		wg.Add(1)
		go func(i int) {
			get(url, i)
			wg.Done()
		}(i)
		time.Sleep(200 * time.Millisecond)
	}
	wg.Wait()
}

func get(url string, i int) {
	body, err := b.Execute(func() (interface{}, error) {
		resp, err := http.Get(url)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		err = errors.New("test")
		return body, err
	})
	fmt.Println(body, err, i)
}
