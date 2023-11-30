package main

import "github.com/shadow1ng/fscan/routers"

func main() {
	_ = routers.InitApiRouter().Run(":12100")
}
