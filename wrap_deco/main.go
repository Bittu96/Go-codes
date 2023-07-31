package main

import (
	"fmt"
)

type Client struct{}

func (c Client) api(request string) string {
	response := fmt.Sprintf("special %v", request)
	return response
}

func New() API {
	return Client{}
}

// API interface - collection of apis
type API interface {
	api(request string) string
}

// Wrapper struct with same api method
type Wrapper struct {
	target API
}

func (w Wrapper) api(request string) string {
	defer fmt.Println("added chicken")
	return w.target.api("chicken " + request)
}

func wrap(client API) API {
	return Wrapper{target: client}
}

// Wrapper2 struct with same api method
type Wrapper2 struct {
	target API
}

func (w Wrapper2) api(request string) string {
	return w.target.api("dum " + request)
}

func wrap2(client API) API {
	return Wrapper2{target: client}
}

// test wraps
func main() {
	client := New()
	fmt.Println(client.api("biryani"))
	//wrapper 1
	client = wrap(client)
	fmt.Println(client.api("biryani"))
	//// wrapper 2
	//client = wrap2(client)
	//fmt.Println(client.api("biryani"))
	return
}
