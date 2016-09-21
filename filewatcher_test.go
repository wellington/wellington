package wellington

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestRebuild_error(t *testing.T) {
	var f *os.File
	log.SetOutput(f)

	tdir, err := ioutil.TempDir("", "rebuild_error")
	if err != nil {
		t.Fatal(err)
	}

	pmap := NewPartialMap()
	path := "test/file1a.scss"
	pmap.Add("file/event", []string{path})
	wc, err := NewWatcher(&WatchOptions{
		PartialMap: pmap,
		BArgs: &BuildArgs{
			BuildDir: tdir,
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	doneChanMu.Lock()
	doneChan = make(chan string, 1)
	doneChanMu.Unlock()
	defer func() {
		doneChanMu.Lock()
		doneChan = nil
		doneChanMu.Unlock()
	}()

	done := make(chan error)
	go func(t *testing.T) {
		select {
		case p := <-doneChan:
			if p != path {
				t.Errorf("got: %s wanted: %s", p, path)
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
	err = wc.rebuild("file/event")
	if err != nil {
		t.Error(err)
	}

	err = <-done
	if err != nil {
		t.Fatal(err)
	}
	if err := wc.Close(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(1 * time.Second)
}

func testPartialRelation() (*os.File, string, string, error) {
	tdir, err := ioutil.TempDir(os.TempDir(), "testwatch_")
	if err != nil {
		return nil, "", "", err
	}
	tfile := filepath.Join(tdir, "_new.scss")
	fh, err := os.Create(tfile)
	if err != nil {
		return nil, "", "", err
	}
	return fh, tdir, tfile, nil
}

func TestRebuild_watch(t *testing.T) {
	var f *os.File
	log.SetOutput(f)

	fh, tdir, tfile, err := testPartialRelation()
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

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
		BArgs: &BuildArgs{
			BuildDir: filepath.Join(tdir, "build"),
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	done := make(chan error)
	go func(t *testing.T) {
		select {
		case err := <-w.errChan:
			done <- err
		case <-rebuildChan:
			done <- nil
		case <-time.After(2 * time.Second):
			done <- fmt.Errorf("timeout waiting for rebuild")
		}
	}(t)
	defer w.Close()
	err = w.Watch()
	if err != nil {
		t.Fatal(err)
	}

	fh.WriteString("boom")

	err = <-done
	if err != nil {
		t.Fatal(err)
	}
}

func TestWatch_success(t *testing.T) {
	w, err := NewWatcher(&WatchOptions{})
	if err != nil {
		t.Fatal(err)
	}

	err = w.Watch()
	if err == nil {
		t.Error("No errors thrown for nil directories")
	}
	w.Close()

	fh, tdir, tfile, err := testPartialRelation()
	if err != nil {
		t.Fatal(err)
	}
	defer fh.Close()

	pmap := NewPartialMap()
	pmap.AddRelation("file/event", tfile)

	watcherChanMu.Lock()
	watcherChan = make(chan string, 1)
	watcherChanMu.Unlock()
	defer func() {
		watcherChanMu.Lock()
		watcherChan = nil
		watcherChanMu.Unlock()
	}()

	w, err = NewWatcher(&WatchOptions{
		Paths:      []string{tdir},
		PartialMap: pmap,
	})
	if err != nil {
		t.Fatal(err)
	}
	err = w.Watch()

	wg := sync.WaitGroup{}
	wg.Add(1)
	// Test file creation event
	go func() {
		defer wg.Done()

		select {
		case <-w.closing:
		case name := <-watcherChan:
			fmt.Println(name)
			return
		case <-time.After(5 * time.Second):
			fmt.Printf("timeout %d\n", len(watcherChan))
			t.Error("Timeout without creating file")
		}
	}()

	testFile := "test/watchfile.lock"
	f, err := os.OpenFile(testFile, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		t.Fatalf("creating test file failed: %s", err)
	}
	defer func() {
		// Give time for filesystem to sync before deleting file
		time.Sleep(50 * time.Millisecond)
		os.Remove(testFile)
		f.Close()
	}()
	f.Sync()

	wg.Add(1)
	// Test file modification event
	go func() {
		defer wg.Done()

		select {
		case <-w.closing:
		case name := <-watcherChan:
			fmt.Println(name)
			return
		case <-time.After(2 * time.Second):
			fmt.Printf("timeout %d\n", len(watcherChan))
			t.Error("Timeout without detecting write")
		}
	}()

	f.WriteString("data")
	f.Sync()
	w.Close()
	wg.Wait()
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
