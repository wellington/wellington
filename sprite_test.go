package sprite_sass_test

import (
	"os"
	"testing"

	. "github.com/drewwells/sprite_sass"
)

func TestCombine(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/139.jpg", "test/140.jpg")
	imgs.Vertical = true
	imgs.Combine()

	if imgs.Height() != 279 {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Height(), 279)
	}

	if imgs.Width() != 96 {
		t.Errorf("Invalid Width found %d, wanted %d", imgs.Width(), 192)
	}

	if x := imgs.X(1); x != 0 {
		t.Errorf("Invalid X found %d, wanted %d", x, 0)
	}

	if y := imgs.Y(1); y != 139 {
		t.Errorf("Invalid Y found %d, wanted %d", y, 139)
	}

	//Quick cache check
	imgs.Combine()
	if imgs.Height() != 279 || imgs.Width() != 96 {
		t.Errorf("Cache invalid")
	}

	testFile := "test/output.jpg"
	err := imgs.Export(testFile)

	defer func() {
		//Cleanup test files
		err := os.Remove(testFile)

		if err != nil {
			panic(err)
		}

	}()

	if err != nil {
		panic(err)
	}
}

func TestLookup(t *testing.T) {

	imgs := ImageList{}
	imgs.Decode("test/139.jpg", "test/140.jpg")
	if f := imgs.Lookup("test/139.jpg"); f != 0 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 0)
	}

	if f := imgs.Lookup("test/140.jpg"); f != 1 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 1)
	}

	if f := imgs.Lookup("140"); f != 1 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 1)
	}

	if f := imgs.Lookup("notafile.jpg"); f != -1 {
		t.Errorf("Found a file that doesn't exist")
	}
}

//Test file globbing
func TestGlob(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/*.png")
	if f := imgs.Lookup("test/139.png"); f != 0 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 0)
	}

	if f := imgs.Lookup("test/140.png"); f != 1 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 1)
	}

	if f := imgs.Lookup("notafile.png"); f != -1 {
		t.Errorf("Found a file that doesn't exist")
	}
}

func TestDecode(t *testing.T) {
	//Should fail with unable to find file
	i := ImageList{}
	err := i.Decode("notafile")

	if err != nil {
		t.Errorf("Error thrown for non-existant file")
	}

	if len(i.Images) > 0 {
		t.Errorf("Found a non-existant file")
	}
}

func TestHorizontal(t *testing.T) {

	imgs := ImageList{}
	imgs.Decode("test/139.jpg", "test/140.jpg")
	imgs.Vertical = false
	imgs.Combine()

	if e := 140; imgs.Height() != e {
		t.Errorf("Invalid Height found %d, wanted %d", imgs.Height(), e)
	}

	if e := 192; imgs.Width() != e {
		t.Errorf("Invalid Width found %d, wanted %d", imgs.Width(), e)
	}

	if e := 96; imgs.X(1) != e {
		t.Errorf("Invalid X found %d, wanted %d", imgs.X(1), e)
	}

	if e := 0; imgs.Y(1) != e {
		t.Errorf("Invalid Y found %d, wanted %d", imgs.Y(1), e)
	}
}
