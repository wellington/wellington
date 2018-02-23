// +build !darwin

package wellington

import (
	"log"

	fsnotify "gopkg.in/fsnotify.v1"
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
	closing chan struct{}
	closed  chan struct{}
}

// Init initializes the watcher with fsnotify watcher
func (w *Watcher) Init() {
	var err error
	w.fw, err = fsnotify.NewWatcher()
	w.closing = make(chan struct{})
	if err != nil {
		log.Fatal(err)
	}
}

func (w *Watcher) startWatching() {

	for {
		select {
		case <-w.closing:
			close(w.closed)
			return
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
}

func (w *Watcher) watch(fpath string) error {
	if len(fpath) > 0 {
		if err := w.fw.Add(fpath); nil != err {
			return err
		}
	}
	return nil
}

// Close shuts down the fsevent stream
func (w *Watcher) Close() error {
	close(w.closing)
	if w.closed != nil {
		<-w.closed
	}
	if w.fw != nil {
		return w.fw.Close()
	}
	return nil
}
