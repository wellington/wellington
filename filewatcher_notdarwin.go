// +build !darwin

package wellington

import (
	"fmt"
	"log"

	"gopkg.in/fsnotify.v1"
)

// Watcher holds all data needed to kick off a build of the css when a
// file changes.
// FileWatcher is the object that triggers builds when a file changes.
// PartialMap contains a mapping of partials to top level files.
// Dirs contains all directories that have top level files.
// GlobalBuildArgs contains build args that apply to all sass files.
type Watcher struct {
	fw      *fsnotify.Watcher
	opts    *WatchOptions
	errChan chan error
}

// Init initializes the watcher with fsnotify watcher
func (w *Watcher) Init() {
	var err error
	w.fw, err = fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
}

func (w *Watcher) startWatching() {
	go func() {
		for {
			select {
			case event := <-w.fw.Events:
				if watcherChan != nil {
					watcherChan <- event.Name
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					err := w.rebuild(event.Name)
					if err != nil {
						log.Println("rebuild error:", err)
					}
				}
			case err := <-w.fw.Errors:
				if err != nil {
					log.Println("filewatcher error:", err)
				}
			}
		}
	}()
}

func (w *Watcher) watch(fpath string) error {
	if len(fpath) > 0 {
		fmt.Println("append", fpath)
		if err := w.fw.Add(fpath); nil != err {
			return err
		}
	}
	return nil
}
