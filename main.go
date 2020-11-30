package main

import "demo/routers"

func main() {
	r := routers.NewRouter()
	r.Run()
}
