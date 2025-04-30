package main

import (
	"io"
	"net"
	"net/http"
	"os"
)

func main() {
	dialer := net.Dialer{}

	tr := http.Transport{
		DialContext: dialer.DialContext,
	}

	client := http.Client{
		Transport: &tr,
	}

	resp, err := client.Get("http://example.com")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		panic("failed to get a successful response")
	}
	io.Copy(os.Stdout, resp.Body)
}
