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

// Sets the default size of the slice holding the top level files for a
// sass partial in SafePartialMap.M
const MaxTopLevel int = 20

// BuildArgs holds universal arguments for a build that the parser
// uses during the initial build and the filewatcher passes back to
// the parser on any file changes.
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

// Watcher holds all data needed to kick off a build of the css when a
// file changes.
// FileWatcher is the object that triggers builds when a file changes.
// PartialMap contains a mapping of partials to top level files.
// Dirs contains all directories that have top level files.
// GlobalBuildArgs contains build args that apply to all sass files.
type Watcher struct {
	FileWatcher *fsnotify.Watcher
	PartialMap  *SafePartialMap
	Dirs        []string
	BArgs       *BuildArgs
}

func NewWatcher() *Watcher {
	var fswatcher *fsnotify.Watcher
	fswatcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}

	return &Watcher{
		FileWatcher: fswatcher,
	}
}

// SafePartialMap is a thread safe map of partial sass files to top
// level files. The file watcher will detect changes in a partial and
// kick off builds for all top level files that contain that partial.
type SafePartialMap struct {
	sync.RWMutex
	M map[string][]string
}

// NewPartialMap creates a initialized SafeParitalMap with with capacity 100
func NewPartialMap() *SafePartialMap {
	spm := &SafePartialMap{
		M: make(map[string][]string, 100)}
	return spm
}

// AddRelation links a partial Sass file with the top level file by
// adding a thread safe entry into partialMap.M.
func (p *SafePartialMap) AddRelation(mainfile string, subfile string) {
	p.Lock()
	//check to see if the map exists, if not initialize the top level map
	if _, ok := p.M[subfile]; !ok {
		p.M[subfile] = make([]string, 0, MaxTopLevel)
	}

	p.M[subfile] = appendUnique(p.M[subfile], mainfile)
	p.Unlock()
}

// Watch is the main entry point into filewatcher and sets up the
// SW object that begins monitoring for file changes and triggering
// top level sass rebuilds.
func (w *Watcher) Watch() error {
	if w.PartialMap == nil {
		w.PartialMap = NewPartialMap()
	}
	if len(w.Dirs) == 0 {
		return fmt.Errorf("No directories to watch")
	}
	err := w.watchFiles()
	if err != nil {
		return err
	}
	w.startWatching()
	return nil
}

func (w *Watcher) watchFiles() error {
	//Watch the dirs of all sass partials
	w.PartialMap.RLock()
	for k := range w.PartialMap.M {
		err := w.watch(filepath.Dir(k))
		if err != nil {
			return err
		}
	}
	w.PartialMap.RUnlock()

	//Watch the dirs of all top level files
	for k := range w.Dirs {
		err := w.watch(w.Dirs[k])
		if err != nil {
			return err
		}
	}
	return nil
}

var watcherChan chan (string)

func (w *Watcher) startWatching() {
	go func() {
		for {
			select {
			default:
				// only called when w.FileWatcher.Events is set to nil.
			case event := <-w.FileWatcher.Events:
				if watcherChan != nil {
					watcherChan <- event.Name
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					w.rebuild(event.Name)
				}
			case err := <-w.FileWatcher.Errors:
				if err != nil {
					fmt.Println("error:", err)
				}
			}
		}
	}()
}

// rebuild is notified about sass file updates and looks
// for the file in the partial map.  It also checks
// for whether the file is a non-partial, no _ at beginning,
// and requests the file be rebuilt directly.
func (w *Watcher) rebuild(eventFileName string) error {
	// Top level file modified, rebuild it directly
	if !strings.HasPrefix(filepath.Base(eventFileName), "_") {
		err := LoadAndBuild(eventFileName, w.BArgs, w.PartialMap)
		if err != nil {
			return err
		}
	}
	w.PartialMap.RLock()
	go func(paths []string) {
		for i := range paths {
			// TODO: do this in a new goroutine
			LoadAndBuild(paths[i], w.BArgs, w.PartialMap)
		}
	}(w.PartialMap.M[eventFileName])
	w.PartialMap.RUnlock()
	return nil
}

func (w *Watcher) watch(fpath string) error {
	if len(fpath) > 0 {
		if err := w.FileWatcher.Add(fpath); nil != err {
			return err
		}
	}
	return nil
}

func appendUnique(slice []string, s string) []string {
	for _, ele := range slice {
		if ele == s {
			return slice
		}
	}
	return append(slice, s)
}
