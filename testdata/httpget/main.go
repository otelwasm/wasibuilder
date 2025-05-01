package main

import (
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

func init() {
	dialer := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	http.DefaultTransport.(*http.Transport).DialContext = dialer.DialContext
}

func main() {
	resp, err := http.Get("http://httpbin.org/get")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	log.Println("Response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		panic("failed to get a successful response")
	}
	io.Copy(os.Stdout, resp.Body)
}
