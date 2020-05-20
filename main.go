// test project main.go
package main

import (
	"htpTest/config"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"
)

var (
	requestIndex uint64 = 0
	successCount uint64 = 0
	failCount    uint64 = 0
	cfg          *config.Config
	wg           *sync.WaitGroup
	httpClient   *http.Client
)

func test(w *sync.WaitGroup) {
	w.Add(1)
	defer func() {
		w.Done()
	}()

	resp, err := http.Get(cfg.Url)
	if err != nil {
		atomic.AddUint64(&failCount, 1)
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		atomic.AddUint64(&failCount, 1)
		log.Println(err)
		return
	}

	atomic.AddUint64(&successCount, 1)

	atomic.AddUint64(&requestIndex, 1)
	log.Println(requestIndex, string(buf))
}

func testA(w *sync.WaitGroup) {
	w.Add(1)
	defer func() {
		w.Done()
	}()

	req, err := http.NewRequest("GET", cfg.Url, nil)
	if err != nil {
		atomic.AddUint64(&failCount, 1)
		log.Println(err)
		return
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := httpClient.Do(req)
	if err != nil {
		atomic.AddUint64(&failCount, 1)
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	var buf []byte
	buf, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		atomic.AddUint64(&failCount, 1)
		log.Println(err)
		return
	}

	atomic.AddUint64(&successCount, 1)

	atomic.AddUint64(&requestIndex, 1)
	log.Println(requestIndex, string(buf))
}

func createHttpClient() *http.Client {
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext,
			MaxConnsPerHost: 50,
			MaxIdleConns:    50,
		},
		Timeout: 15 * time.Second,
	}
	return client
}

func main() {
	defer func() {
		time.Sleep(time.Millisecond * time.Duration(100))
	}()

	cfg = config.NewConfig()
	err := cfg.Load("conf/conf.json")
	if err != nil {
		log.Println(err)
		return
	}

	httpClient = createHttpClient()

	wg = &sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < cfg.Count; i++ {
		go testA(wg)
	}
	wg.Wait()

	end := time.Now()

	log.Printf("success:[%d], fail:[%d], time:[%v]", successCount, failCount, end.Sub(start))
}
