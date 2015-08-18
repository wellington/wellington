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
	wc := NewWatcher()
	go func(t *testing.T) {
		select {
		case err := <-errChan:
			if err == nil {
				t.Fatal(err)
			}
			if e := fmt.Errorf("build args"); e != err {
				t.Fatalf("got: %s wanted: %s", e, err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Fatal("timeout waiting for load error")
		}
	}(t)

	// rebuild doesn't throw errors ever
	wc.rebuild("file/event")
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

	w := NewWatcher()
	w.Dirs = []string{tdir}
	w.PartialMap.AddRelation("tswif", tfile)
	err = w.Watch()
	if err != nil {
		t.Fatal(err)
	}
	rebuildChan = make(chan []string, 1)
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
	w := NewWatcher()
	err := w.Watch()
	if err == nil {
		t.Error("No errors thrown for nil directories")
	}
	w.FileWatcher.Close()

	watcherChan = make(chan string, 1)
	w = NewWatcher()
	w.Dirs = []string{"test"}
	err = w.Watch()

	// Test file creation event
	go func() {
		select {
		case <-watcherChan:
			break
		case <-time.After(2 * time.Second):
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
