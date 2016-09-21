// +build darwin

package wellington

import (
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/fsnotify/fsevents"
)

// Watcher holds all data needed to kick off a build of the css when a
// file changes.
// FileWatcher is the object that triggers builds when a file changes.
// PartialMap contains a mapping of partials to top level files.
// Dirs contains all directories that have top level files.
// GlobalBuildArgs contains build args that apply to all sass files.
type Watcher struct {
	es      *fsevents.EventStream
	opts    *WatchOptions
	errChan chan error
	closing chan struct{}
	closed  chan struct{}
}

// Init initializes the watcher with fsevent watcher
func (w *Watcher) Init() {
	if w.es != nil {
		w.es.Stop()
	}
	w.closing = make(chan struct{})
	w.es = &fsevents.EventStream{
		Latency: 500 * time.Millisecond,
		Flags:   fsevents.FileEvents,
	}
}

func (w *Watcher) startWatching() {
	w.closed = make(chan struct{})
	w.es.Start()
	for {
		select {
		case <-w.closing:
			close(w.closed)
			return
		case msg := <-w.es.Events:
			for _, event := range msg {
				ext := filepath.Ext(event.Path)
				if ext == ".scss" || ext == ".sass" {
					if !checkFlag(event.Flags) {
						log.Println("ignoring fsevent", event.Flags, "on", event.Path)
						continue
					}
					if watcherChan != nil {
						watcherChan <- event.Path
						return
					}
					if strings.HasPrefix(event.Path, "/private") {
						event.Path = strings.TrimPrefix(event.Path, "/private")
					}

					err := w.rebuild(event.Path)
					if err != nil {
						log.Println("rebuild error:", err)
					}
				}
			}
		}
	}
}

var modEvts = []fsevents.EventFlags{
	fsevents.ItemCreated,
	fsevents.ItemRemoved,
	fsevents.ItemRenamed,
	fsevents.ItemModified,
}

// checkFlag lets us know if this event is important
func checkFlag(e fsevents.EventFlags) bool {
	if e&fsevents.ItemIsFile == 0 {
		return false
	}

	for i := range modEvts {
		if e&modEvts[i] == modEvts[i] {
			return true
		}
	}
	return false
}

func (w *Watcher) watch(fpath string) error {
	if len(fpath) > 0 {

		w.es.Paths = appendUnique(w.es.Paths, fpath)
	}
	return nil
}

// Close shuts down the fsevent stream
func (w *Watcher) Close() error {
	close(w.closing)
	if w.es != nil {
		w.es.Stop()
	}
	if w.closed != nil {
		<-w.closed
	}
	return nil
}
