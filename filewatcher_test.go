package wellington

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPartial_map(t *testing.T) {
	path, _ := filepath.Abs("test/sass/import.scss")
	p := Parser{
		BuildDir:   "test/build",
		Includes:   []string{"test/sass"},
		MainFile:   path,
		PartialMap: NewPartialMap(),
		SassDir:    os.Getenv("PWD"),
	}

	p.Start(fileReader("test/sass/import.scss"), "test/")

	if e := 1; len(p.PartialMap.M) != e {
		t.Errorf("got: %d, wanted: %d", len(p.PartialMap.M), e)
	}

	for k := range p.PartialMap.M {
		if e := 1; len(p.PartialMap.M[k]) != e {
			t.Errorf("got: %d wanted: %d", len(p.PartialMap.M[k]), e)
		}
		if p.PartialMap.M[k][0] != path {
			t.Errorf("got: %s wanted: %s", p.PartialMap.M[k][0], path)
		}
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

	timeout := make(chan bool, 1)
	go func() {
		time.Sleep(500 * time.Millisecond)
		timeout <- true
	}()

	// Test file creation event
	go func() {
		select {
		case <-watcherChan:
			break
		case <-timeout:
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
		case <-timeout:
			fmt.Printf("timeout %d\n", len(watcherChan))
			t.Error("Timeout without detecting write")
		}
	}()

	f.WriteString("data")
	f.Sync()

}

func TestRebuild(t *testing.T) {
	w := NewWatcher()
	err := w.rebuild("file/event")

	if e := fmt.Sprintf("build args are nil"); e != err.Error() {
		t.Errorf("wanted: %s\ngot: %s", e, err)
	}
}
