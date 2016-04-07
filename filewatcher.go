package wellington

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// MaxTopLevel sets the default size of the slice holding the top level
// files for a sass partial in SafePartialMap.M
const MaxTopLevel int = 20

// WatchOptions containers the necessary parameters to run the file watcher
type WatchOptions struct {
	PartialMap *SafePartialMap
	Paths      []string
	BArgs      *BuildArgs
}

// NewWatchOptions returns a new WatchOptions
func NewWatchOptions() *WatchOptions {
	return &WatchOptions{
		PartialMap: NewPartialMap(),
	}
}

// NewWatcher returns a new watcher pointer
func NewWatcher(opts *WatchOptions) (*Watcher, error) {
	if opts == nil {
		opts = &WatchOptions{}
	}
	w := &Watcher{
		opts:    opts,
		errChan: make(chan error),
	}
	w.Init()
	// FIXME: this will leak routines, but watcher should only be
	// called once
	go func() {
		for {
			select {
			case err := <-w.errChan:
				log.Println("watcher: build error:", err)
			}
		}
	}()

	return w, nil
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

// Add places a path in the partial map
func (p *SafePartialMap) Add(key string, paths []string) {
	p.Lock()
	defer p.Unlock()
	p.M[key] = paths
}

// Get is a thread-safe way to access the partial map
func (p *SafePartialMap) Get(key string) ([]string, bool) {
	p.RLock()
	defer p.RUnlock()
	pm, ok := p.M[key]
	return pm, ok
}

// AddRelation links a partial Sass file with the top level file by
// adding a thread safe entry into partialMap.M.
func (p *SafePartialMap) AddRelation(mainfile string, subfile string) {
	existing, _ := p.Get(subfile)
	p.Add(subfile, appendUnique(existing, mainfile))
}

var watcherChan chan (string)

// Watch is the main entry point into filewatcher and sets up the
// SW object that begins monitoring for file changes and triggering
// top level sass rebuilds.
func (w *Watcher) Watch() error {
	if w.opts.PartialMap == nil {
		w.opts.PartialMap = NewPartialMap()
	}

	if len(w.opts.Paths) == 0 {
		return errors.New("No paths to watch")
	}
	err := w.watchFiles()
	if err != nil {
		return err
	}
	w.startWatching()
	return nil
}

func (w *Watcher) watchFiles() error {
	var err error
	//Watch the dirs of all sass partials
	w.opts.PartialMap.RLock()
	for k := range w.opts.PartialMap.M {
		dir := filepath.Dir(k)
		_, err = os.Stat(dir)
		if !os.IsNotExist(err) && filepath.IsAbs(dir) {
			err = w.watch(dir)
			if err != nil {
				return err
			}
		}
	}
	w.opts.PartialMap.RUnlock()
	return nil
}

var rebuildMu sync.RWMutex
var rebuildChan chan ([]string)

// rebuild is notified about sass file updates and looks
// for the file in the partial map.  It also checks
// for whether the file is a non-partial, no _ at beginning,
// and requests the file be rebuilt directly.
func (w *Watcher) rebuild(eventFileName string) error {
	paths, ok := w.opts.PartialMap.Get(eventFileName)
	if !ok {
		// This isn't an error per say, so let's ignore it
		return nil
		// return fmt.Errorf("partial map lookup failed: %s", eventFileName)
	}

	go func(paths []string) {
		rebuildMu.RLock()
		if rebuildChan != nil {
			rebuildChan <- paths
		}
		rebuildMu.RUnlock()
		for i := range paths {
			// TODO: do this in a new goroutine
			err := LoadAndBuild(paths[i], w.opts.BArgs, w.opts.PartialMap)
			if err != nil {
				w.errChan <- err
			} else {
				fmt.Printf("Rebuilt: %s\n", paths[i])
			}
		}
	}(paths)
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
