package wellington

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRebuild(t *testing.T) {
	var f *os.File
	log.SetOutput(f)
	pmap := NewPartialMap()
	path := "test/file1a.scss"
	pmap.Add("file/event", []string{path})
	wc, err := NewWatcher(&WatchOptions{
		PartialMap: pmap,
	})
	if err != nil {
		t.Fatal(err)
	}

	rebuildMu.Lock()
	rebuildChan = make(chan []string, 1)
	rebuildMu.Unlock()
	defer func() {
		rebuildMu.Lock()
		rebuildChan = nil
		rebuildMu.Unlock()
	}()

	done := make(chan error)
	go func(t *testing.T) {
		select {
		case p := <-rebuildChan:
			if p[0] != path {
				t.Errorf("got: %s wanted: %s", p[0], path)
			}
			done <- nil
		case err := <-wc.errChan:
			if err == nil {
				done <- err
			}
			if e := fmt.Errorf("build args"); e != err {
				done <- fmt.Errorf("got: %s wanted: %s", e, err)
			}
		case <-time.After(5000 * time.Millisecond):
			done <- fmt.Errorf("timeout waiting for load error")
		}
	}(t)

	// rebuild doesn't throw errors ever
	wc.rebuild("file/event")

	err = <-done
	if err != nil {
		t.Fatal(err)
	}
}

func TestRebuild_watch(t *testing.T) {
	tdir, err := ioutil.TempDir(os.TempDir(), "testwatch_")
	if err != nil {
		t.Fatal(err)
	}
	tfile := filepath.Join(tdir, "_new.scss")
	fh, err := os.Create(tfile)
	if err != nil {
		t.Fatal(err)
	}

	rebuildMu.Lock()
	rebuildChan = make(chan []string, 1)
	rebuildMu.Unlock()
	defer func() {
		rebuildMu.Lock()
		rebuildChan = nil
		rebuildMu.Unlock()
	}()

	pMap := NewPartialMap()
	pMap.AddRelation("tswif", tfile)
	w, err := NewWatcher(&WatchOptions{
		Paths:      []string{tdir},
		PartialMap: pMap,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = w.Watch()
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan bool, 1)
	go func(t *testing.T) {
		select {
		case <-rebuildChan:
			done <- true
		case <-time.After(2 * time.Second):
			done <- false
		}
		done <- true
	}(t)
	fh.WriteString("boom")
	success := <-done
	if !success {
		t.Fatal("Timeout waiting for rebuild")
	}

}

func TestWatch(t *testing.T) {
	w, err := NewWatcher(NewWatchOptions())
	if err != nil {
		t.Fatal(err)
	}
	err = w.Watch()
	if err == nil {
		t.Error("No errors thrown for nil directories")
	}
	// w.fw.Close()

	watcherChan = make(chan string, 1)
	w, err = NewWatcher(&WatchOptions{
		Paths:      []string{"test"},
		PartialMap: NewPartialMap(),
	})
	if err != nil {
		t.Fatal(err)
	}
	err = w.Watch()

	// Test file creation event
	go func() {
		select {
		case <-watcherChan:
			break
		case <-time.After(5 * time.Second):
			fmt.Printf("timeout %d\n", len(watcherChan))
			t.Error("Timeout without creating file")
		}
	}()

	testFile := "test/watchfile.lock"
	f, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE, 0666)
	defer func() {
		// Give time for filesystem to sync before deleting file
		time.Sleep(50 * time.Millisecond)
		os.Remove(testFile)
		f.Close()
	}()
	if err != nil {
		t.Fatalf("creating test file failed: %s", err)
	}
	f.Sync()

	// Test file modification event
	go func() {
		select {
		case <-watcherChan:
			break
		case <-time.After(2 * time.Second):
			fmt.Printf("timeout %d\n", len(watcherChan))
			t.Error("Timeout without detecting write")
		}
	}()

	f.WriteString("data")
	f.Sync()
}

func TestWatch_errors(t *testing.T) {
	w, err := NewWatcher(nil)
	if err != nil {
		t.Fatal(err)
	}

	if w.opts == nil {
		t.Fatal("unexpected nil")
	}
}

func TestAppendUnique(t *testing.T) {
	lst := []string{"a", "b", "c"}
	new := appendUnique(lst, "a")
	if len(new) != len(lst) {
		t.Errorf("got: %d wanted: %d", len(new), len(lst))
	}

	new = appendUnique(lst, "d")
	if len(new) != len(lst)+1 {
		t.Errorf("got: %d wanted: %d", len(new), len(lst)+1)
	}
}
