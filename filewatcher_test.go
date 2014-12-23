package wellington

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestPartialMap(t *testing.T) {
	path, _ := filepath.Abs("test/sass/import.scss")
	p := Parser{
		BuildDir:   "test/build",
		Includes:   []string{"test/sass"},
		MainFile:   path,
		PartialMap: NewPartialMap(),
	}

	p.Start(fileReader("test/sass/import.scss"), "test/")
	if len(p.PartialMap.M) != 2 {
		t.Errorf("SafePartialMap expected Size: 1 got: %d", len(p.PartialMap.M))
	}

	for k := range p.PartialMap.M {
		if len(p.PartialMap.M[k]) != 1 {
			t.Errorf("SafePartialMap for test/sass/import.scss expected 1 top level file, got %d", len(p.PartialMap.M[k]))
		}
		if p.PartialMap.M[k][0] != path {
			t.Errorf("SafePartialMap expected all top level files to be %s, got %s", path, p.PartialMap.M[k][0])
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
