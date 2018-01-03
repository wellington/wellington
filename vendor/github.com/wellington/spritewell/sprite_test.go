package spritewell

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type tmp struct {
	Build string
	Image string
}

// Close removes temporary directories
func (t tmp) Close() error {
	err := os.RemoveAll(t.Build)
	if err != nil {
		return err
	}
	return os.RemoveAll(t.Image)
}

// returns build and image build directories
func setupTemp(name string) tmp {
	tdir, _ := ioutil.TempDir("", "TestSprite_combine")
	gdir := filepath.Join(tdir, "imgs")
	os.MkdirAll(gdir, 0700)
	return tmp{Build: tdir, Image: gdir}
}

func TestSpriteLookup(t *testing.T) {

	imgs := New(nil)
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
		t.Errorf("Found a file that doesn't exist was: %d, expected: %d",
			imgs.Lookup("noatfile.jpg"), -1)
	}
}

func TestSprite_race(t *testing.T) {
	tmp := setupTemp("TestSprite_dimensions")
	imgs := New(&Options{
		GenImgDir: tmp.Image,
		BuildDir:  tmp.Build,
	})
	glob := []string{"test/139.jpg", "test/140.jpg"}
	imgs.Decode(glob...)

	go func() {
		imgs.Decode(glob...)

	}()
	// time.Sleep(1 * time.Millisecond)
	imgs.ImageHeight(0)
	imgs.ImageWidth(0)
	imgs.X(0)
	imgs.Y(0)
	imgs.Dimensions()
	imgs.File(".")
	imgs.OutputPath()
	go imgs.Export()
	imgs.Lookup("poop")
	imgs.GetPack(0)
	_ = imgs.Paths

	a := &imgs
	_ = a
}

func TestSprite_dimensions(t *testing.T) {
	tmp := setupTemp("TestSprite_dimensions")
	imgs := New(&Options{
		BuildDir:  tmp.Build,
		GenImgDir: tmp.Image,
	})
	defer tmp.Close()
	glob := []string{"test/139.jpg", "test/140.jpg"}
	imgs.Decode(glob...)

	bounds := imgs.Dimensions()
	if bounds.Y != 279 {
		t.Errorf("Invalid Height found %d, wanted %d", bounds.Y, 279)
	}

	if bounds.X != 96 {
		t.Errorf("Invalid Width found %d, wanted %d", bounds.X, 192)
	}

	if x := imgs.X(1); x != 0 {
		t.Errorf("Invalid X found %d, wanted %d", x, 0)
	}

	if y := imgs.Y(1); y != 139 {
		t.Errorf("Invalid Y found %d, wanted %d", y, 139)
	}

	if e := -1; e != imgs.SImageWidth("150") {
		t.Errorf("Non-existant image width invalid"+
			"\n    was:%d\nexpected:%d",
			imgs.SImageWidth("150"), e)
	}

	if e := -1; e != imgs.SImageHeight("150") {
		t.Errorf("Non-existant image width invalid"+
			"\n    was:%d\nexpected:%d",
			imgs.SImageHeight("150"), e)
	}

	//Quick cache check

	bounds = imgs.Dimensions()
	if bounds.Y != 279 || bounds.X != 96 {
		t.Errorf("Cache invalid")
	}

	testFile, err := imgs.Export()
	if e := "imgs/1e1dbf.png"; !strings.HasSuffix(testFile, e) {
		t.Fatalf("got: %s wanted: %s", testFile, e)
	}
	if err != nil {
		t.Fatal(err)
	}
	err = imgs.Wait()
	if err != nil {
		t.Fatal(err)
	}
}

//Test file globbing
func TestSpriteGlob(t *testing.T) {
	imgs := New(&Options{
		ImageDir: "test",
	})
	imgs.Decode("*.png")

	// Test [Un]successful lookups
	if f := imgs.Lookup("139.png"); f != 0 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 0)
	}

	if f := imgs.Lookup("140.png"); f != 1 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 1)
	}

	if f := imgs.Lookup("notafile.png"); f != -1 {
		t.Errorf("Found a file that doesn't exist")
	}
}

// ExampleSprite shows how to take all the images matching the glob
// and creating a sprite image in ./test/build/img.
func ExampleSprite() {
	imgs := New(&Options{
		ImageDir:  ".",
		BuildDir:  "test/build",
		GenImgDir: "test/build/img",
	})
	err := imgs.Decode("test/1*.png")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(imgs)

	// Export will start the process of writing the sprite to disk
	of, err := imgs.Export()
	if err != nil {
		log.Fatal(err)
	}
	_ = of
	// Calls are non-blocking, use Wait() to ensure image encoding has
	// completed and results are flushed to disk.
	imgs.Wait()

	// Output:
	// img/68ca3a.png
}

// ExampleSpritePosition shows how to find the position of an image
// in the the spritesheet.
func ExampleSpritePosition() {

	imgs := New(&Options{
		ImageDir:  ".",
		BuildDir:  "test/build",
		GenImgDir: "test/build/img",
	})

	err := imgs.Decode("test/1*.png")
	if err != nil {
		log.Fatal(err)
	}

	pos := imgs.GetPack(imgs.Lookup("140"))

	fmt.Printf(`background: url("%s") -%dpx -%dpx no-repeat;`,
		imgs, pos.X, pos.Y)

	// Output:
	// background: url("img/68ca3a.png") -0px -139px no-repeat;
}

func TestSpriteDecode_fail(t *testing.T) {
	var out bytes.Buffer
	log.SetOutput(&out)
	//Should fail with unable to find file
	i := New(nil)
	err := i.Decode("notafile")
	if err == nil {
		t.Fatal("error expected")
	}

	if e := ErrNoImages; err != e {
		t.Fatalf("got: %s wanted: %s", err, e)
	}

	path, err := i.Export()
	if len(path) > 0 {
		t.Errorf("no path should be returned got: %s", path)
	}
	if err == nil {
		t.Fatal("error expected")
	}

	if e := ErrNoImages; err != e {
		t.Fatalf("got: %s wanted: %s", err, e)
	}

}

func TestSpriteHorizontal(t *testing.T) {

	imgs := New(&Options{
		Pack: "horz",
	})
	imgs.Decode("test/139.jpg", "test/140.jpg")

	bounds := imgs.Dimensions()
	if e := 140; bounds.Y != e {
		t.Errorf("Invalid Height found %d, wanted %d", bounds.Y, e)
	}

	if e := 192; bounds.X != e {
		t.Errorf("Invalid Width found %d, wanted %d", bounds.X, e)
	}

	if e := 96; imgs.X(1) != e {
		t.Errorf("Invalid X found %d, wanted %d", imgs.X(1), e)
	}

	if e := 0; imgs.Y(1) != e {
		t.Errorf("Invalid Y found %d, wanted %d", imgs.Y(1), e)
	}
}

func TestPadding(t *testing.T) {

	imgs := New(&Options{
		Padding: 10,
		Pack:    "horz",
	})
	imgs.Decode("test/139.jpg", "test/140.jpg")

	bounds := imgs.Dimensions()
	if e := 140; bounds.Y != e {
		t.Errorf("Invalid Height found %d, wanted %d", bounds.Y, e)
	}

	if e := 202; bounds.X != e {
		t.Errorf("Invalid Width found %d, wanted %d", bounds.X, e)
	}

	if e := 106; imgs.X(1) != e {
		t.Errorf("Invalid X found %d, wanted %d", imgs.X(1), e)
	}

	if e := 0; imgs.Y(1) != e {
		t.Errorf("Invalid Y found %d, wanted %d", imgs.Y(1), e)
	}
	imgs.optsMu.Lock()
	imgs.opts.Pack = "vert"
	imgs.optsMu.Unlock()
	bounds = imgs.Dimensions()
	if e := 289; bounds.Y != e {
		t.Errorf("Invalid Height found %d, wanted %d", bounds.Y, e)
	}

	if e := 96; bounds.X != e {
		t.Errorf("Invalid Width found %d, wanted %d", bounds.X, e)
	}

	if e := 0; imgs.X(1) != e {
		t.Errorf("Invalid X found %d, wanted %d", imgs.X(1), e)
	}

	if e := 149; imgs.Y(1) != e {
		t.Errorf("Invalid Y found %d, wanted %d", imgs.Y(1), e)
	}

}

func TestSpriteError(t *testing.T) {
	var out bytes.Buffer
	imgs := New(nil)
	log.SetOutput(&out)
	imgs.Decode("test/bad/interlace.png")

	out.ReadString('\n')

	if e := -1; imgs.ImageHeight(-1) != -1 {
		t.Errorf("ImageHeight not found expected: %d, was: %d",
			e, imgs.ImageHeight(-1))
	}

	if e := -1; imgs.ImageWidth(-1) != -1 {
		t.Errorf("ImageWidth not found expected: %d, was: %d",
			e, imgs.ImageWidth(-1))
	}

	if e := ""; imgs.File("notfound") != e {
		t.Errorf("Invalid file call to File expected: %s, was %s",
			e, imgs.File("notfound"))
	}

	log.SetOutput(os.Stdout)
}

func TestCanDecode(t *testing.T) {
	fileMap := []string{"file.png", "file.jpg", "file.gif",
		"dir/dir/file.png", "file.svg"}

	values := []bool{true, true, true, true, false}

	for i := range fileMap {
		b := CanDecode(filepath.Ext(fileMap[i]))
		if values[i] != b {
			t.Errorf("got: %t expected: %t", b, values[i])
		}
	}
}

func TestOutput(t *testing.T) {
	imgs := New(nil)
	imgs.Decode("test/*.png")
	str, err := imgs.OutputPath()
	if err != nil {
		t.Error(err)
	}

	if e := "image/0e64a8.png"; e != str {
		t.Errorf("got: %s wanted: %s", str, e)
	}

	imgs = New(&Options{
		GenImgDir: "../build/img",
		BuildDir:  "../build",
	})
	imgs.Decode("test/*.png")
	str, err = imgs.OutputPath()
	if err != nil {
		t.Error(err)
	}

	if e := "img/dbf3ef.png"; e != str {
		t.Errorf("got: %s wanted: %s", str, e)
	}

}

func TestSprite_many(t *testing.T) {
	tmp := setupTemp("TestSprite_dimensions")
	imgs := New(&Options{
		GenImgDir: tmp.Image,
		BuildDir:  tmp.Build,
		Pack:      "vert",
	})
	imgs.Decode("test/many/*.jpg")
	name, err := imgs.Export()
	if err != nil {
		t.Error(err)
	}
	_ = name

	m := map[string]Pos{
		"bird":   Pos{0, 0},
		"in":     Pos{0, 150},
		"pencil": Pos{0, 300},
		"rss":    Pos{0, 450},
		"twitt":  Pos{0, 600},
	}

	for k, v := range m {
		pos := imgs.GetPack(imgs.Lookup(k))
		if e := v.X; e != pos.X {
			t.Errorf("got: %d wanted: %d", pos.X, e)
		}
		if e := v.Y; e != pos.Y {
			t.Errorf("got: %d wanted: %d", pos.Y, e)
		}
	}

}
