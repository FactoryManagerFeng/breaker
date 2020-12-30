package main

import (
	"code.piupiu.com/book/go_book_common/breaker"
	"fmt"
	"io/ioutil"
	"net/http"
)

var b *breaker.Breaker

func init() {
	var settings breaker.Settings
	settings.Name = "test breaker"
	settings.ReadyToTrip = func(counts breaker.Counts) bool {
		ratio := float64(counts.FailNum) / float64(counts.RequestsNum)
		return counts.RequestsNum >= 3 && ratio >= 0.5
	}
	settings.OnStateChange = func(name string, from, to breaker.State) {
		fmt.Println(name, from, to)
	}
	b = breaker.NewBreaker(settings)
}

func main() {
	var url = "http://www.baidu.com"
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

		return body, nil
	})
	if err != nil {
		fmt.Println(err)
	} else {
		result := body.([]byte)
		fmt.Println(string(result))
	}
}
