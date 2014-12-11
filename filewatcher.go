package sprite_sass

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wellington/spritewell"
	"gopkg.in/fsnotify.v1"
)

const maxTopLevel int = 20

type BuildArgs struct {
	Imgs, Sprites spritewell.SafeImageMap
	Dir           string
	BuildDir      string
	Includes      string
	Font          string
	Gen           string
	Style         int
	Comments      bool
}

type sassWatcher struct {
	FileWatcher             *fsnotify.Watcher
	PartialMap              *SafePartialMap
	TopLevelFileDirectories *[]string
	GlobalBuildArgs         *BuildArgs
}

type SafePartialMap struct {
	sync.RWMutex
	M map[string][]string
}

func NewPartialMap() *SafePartialMap {
	spm := SafePartialMap{
		M: make(map[string][]string, 100)}
	return &spm
}

func (partialMap *SafePartialMap) AddRelation(mainfile string, subfile string) {
	partialMap.Lock()
	//check to see if the map exists, if not initialize the top level map
	if _, exists := partialMap.M[subfile]; !exists {
		partialMap.M[subfile] = make([]string, 0, maxTopLevel)
	}

	partialMap.M[subfile] = appendTopLevelIfMissing(partialMap.M[subfile], mainfile)
	partialMap.Unlock()
}

func FileWatch(partialMap *SafePartialMap, globalBuildArgs *BuildArgs, topLevelFileDirectories *[]string) {
	var err = error(nil)
	var fswatcher *fsnotify.Watcher
	fswatcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer fswatcher.Close()
	sassFileWatcher := sassWatcher{fswatcher, partialMap, topLevelFileDirectories, globalBuildArgs}
	sassFileWatcher.watchFiles()
	sassFileWatcher.startWatching()

}

func (sassFileWatcher *sassWatcher) startWatching() {
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-(*sassFileWatcher.FileWatcher).Events:
				//fmt.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					//fmt.Println("modified file:", event.Name)
					sassFileWatcher.rebuildTopLevelSassFiles(event.Name)
				}
			case err := <-(*sassFileWatcher.FileWatcher).Errors:
				fmt.Println("error:", err)
			}
		}
	}()
	<-done
}

func (sassFileWatcher *sassWatcher) rebuildTopLevelSassFiles(eventFileName string) {
	if strings.HasPrefix(filepath.Base(eventFileName), "_") { //Partial sass file was modified.  Rebuild all top level files that contain it.
		for k := range sassFileWatcher.PartialMap.M[eventFileName] {
			//fmt.Println("Should rebuild:", sassFileWatcher.PartialMap.M[eventFileName][k])
			LoadAndBuild(sassFileWatcher.PartialMap.M[eventFileName][k], sassFileWatcher.GlobalBuildArgs, sassFileWatcher.PartialMap, sassFileWatcher.TopLevelFileDirectories)
		}
	} else { //Top leve file was modified.  Rebuild it.
		//fmt.Println("Should rebuild:", eventFileName)
		LoadAndBuild(eventFileName, sassFileWatcher.GlobalBuildArgs, sassFileWatcher.PartialMap, sassFileWatcher.TopLevelFileDirectories)
	}
}

func (sassFileWatcher *sassWatcher) watchFiles() {
	//Watch the dirs of all sass partials
	for k := range sassFileWatcher.PartialMap.M {
		sassFileWatcher.Watch(filepath.Dir(k))
		//fmt.Println(k)
	}

	//Watch the dirs of all top level files
	for k := range *sassFileWatcher.TopLevelFileDirectories {
		sassFileWatcher.Watch((*sassFileWatcher.TopLevelFileDirectories)[k])
	}
}

func (sassFileWatcher *sassWatcher) Watch(fpath string) {
	if len(fpath) > 0 {
		if err := (*sassFileWatcher.FileWatcher).Add(fpath); nil != err {
			log.Fatalln(err)
		}
	}
}

func appendTopLevelIfMissing(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}
