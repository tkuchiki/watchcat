package main

import (
	"log"

	"github.com/tkuchiki/watchcat"
)

func main() {
	watcher := watchcat.NewWatcher()
	defer watcher.CloseFP()
	log.Fatal(watcher.Cat())
}
