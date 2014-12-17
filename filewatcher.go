package wellington

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/wellington/spritewell"
	"gopkg.in/fsnotify.v1"
)

//Sets the default size of the slice holding the top level files for a sass partial in SafePartialMap.M
const MaxTopLevel int = 20

//BuildArgs holds universal arguments for a build that the parser uses during the initial build and the filewatcher
//passes back to the parser on any file changes.
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

//SW holds all data needed to kick off a build of the css when a file changes.
//FileWatcher is the object that triggers builds when a file changes.
//PartialMap contains a mapping of partials to top level files.
//TopLevelFileDirectories contains all directories that have top level files.
//GlobalBuildArgs contains build args that apply to all sass files.
type SW struct {
	FileWatcher             *fsnotify.Watcher
	PartialMap              *SafePartialMap
	TopLevelFileDirectories []string
	GlobalBuildArgs         *BuildArgs
}

//SafePartialMap is a thread safe map of partial sass files to top level files.
//The file watcher will detect changes in a partial and kick off builds for all
//top level files that contain that partial.
type SafePartialMap struct {
	sync.RWMutex
	M map[string][]string
}

//NewPartialMap creates a initialized SafeParitalMap with with capacity 100
func NewPartialMap() *SafePartialMap {
	spm := SafePartialMap{
		M: make(map[string][]string, 100)}
	return &spm
}

//AddRelation links a partial Sass file with the top level file by adding a thread safe
//entry into partialMap.M.
func (partialMap *SafePartialMap) AddRelation(mainfile string, subfile string) {
	partialMap.Lock()
	//check to see if the map exists, if not initialize the top level map
	if _, exists := partialMap.M[subfile]; !exists {
		partialMap.M[subfile] = make([]string, 0, MaxTopLevel)
	}

	partialMap.M[subfile] = appendTopLevelIfMissing(partialMap.M[subfile], mainfile)
	partialMap.Unlock()
}

//FileWatch is the main entry point into filewatcher and sets up the SW object that begins
//monitoring for file changes and triggering top level sass rebuilds.
func FileWatch(partialMap *SafePartialMap, globalBuildArgs *BuildArgs, topLevelFileDirectories []string) {
	var err = error(nil)
	var fswatcher *fsnotify.Watcher
	fswatcher, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	defer fswatcher.Close()
	sassFileWatcher := SW{fswatcher, partialMap, topLevelFileDirectories, globalBuildArgs}
	sassFileWatcher.watchFiles()
	sassFileWatcher.startWatching()

}

func (sassFileWatcher *SW) startWatching() {
	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-(*sassFileWatcher.FileWatcher).Events:
				if event.Op&fsnotify.Write == fsnotify.Write {
					sassFileWatcher.rebuildTopLevelSassFiles(event.Name)
				}
			case err := <-(*sassFileWatcher.FileWatcher).Errors:
				fmt.Println("error:", err)
			}
		}
	}()
	<-done
}

func (sassFileWatcher *SW) rebuildTopLevelSassFiles(eventFileName string) {
	if strings.HasPrefix(filepath.Base(eventFileName), "_") { //Partial sass file was modified.  Rebuild all top level files that contain it.
		for k := range sassFileWatcher.PartialMap.M[eventFileName] {
			LoadAndBuild(sassFileWatcher.PartialMap.M[eventFileName][k], sassFileWatcher.GlobalBuildArgs, sassFileWatcher.PartialMap, sassFileWatcher.TopLevelFileDirectories)
		}
	} else { //Top leve file was modified.  Rebuild it.
		LoadAndBuild(eventFileName, sassFileWatcher.GlobalBuildArgs, sassFileWatcher.PartialMap, sassFileWatcher.TopLevelFileDirectories)
	}
}

func (sassFileWatcher *SW) watchFiles() {
	//Watch the dirs of all sass partials
	for k := range sassFileWatcher.PartialMap.M {
		sassFileWatcher.watch(filepath.Dir(k))
	}

	//Watch the dirs of all top level files
	for k := range sassFileWatcher.TopLevelFileDirectories {
		sassFileWatcher.watch((sassFileWatcher.TopLevelFileDirectories)[k])
	}
}

func (sassFileWatcher *SW) watch(fpath string) {
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
