package main

import (
	"io"
	"log"
	"net/http"
	"os"
)

func main() {
	resp, err := http.Get("http://example.com")
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
