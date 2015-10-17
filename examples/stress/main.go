package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/wellington/wellington"
)

func main() {
	sassFile := "../../test/sass/file.scss"
	tdir, _ := ioutil.TempDir("basic", "")
	args := &wellington.BuildArgs{}
	args.BuildDir = tdir
	pmap := wellington.NewPartialMap()

	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			b := wellington.NewBuild([]string{sassFile}, args, pmap)
			err := b.Run()
			if err != nil {
				log.Println(i, err)
			}
		}()
		fmt.Println(i)
	}

	wg.Wait()
}
