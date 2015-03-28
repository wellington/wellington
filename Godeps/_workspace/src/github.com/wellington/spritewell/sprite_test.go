package spritewell

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func cleanUpSprites(sprites map[string]ImageList) {
	if sprites == nil {
		return
	}
	for _, iml := range sprites {
		err := os.Remove(filepath.Join(iml.GenImgDir, iml.OutFile))
		if err != nil {
			log.Fatal(err)
		}
	}
}

func TestSpriteLookup(t *testing.T) {

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
		t.Errorf("Found a file that doesn't exist was: %d, expected: %d",
			imgs.Lookup("noatfile.jpg"), -1)
	}
}

func TestSpriteCombine(t *testing.T) {
	imgs := ImageList{}
	glob := []string{"test/139.jpg", "test/140.jpg"}
	imgs.Decode(glob...)
	_, err := imgs.Combine()

	if err != nil {
		t.Error(err)
	}
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
	imgs.Combine()
	bounds = imgs.Dimensions()
	if bounds.Y != 279 || bounds.X != 96 {
		t.Errorf("Cache invalid")
	}

	testFile, err := imgs.Export()

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

//Test file globbing
func TestSpriteGlob(t *testing.T) {
	imgs := ImageList{
		ImageDir: "test",
	}
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

func ExampleSpriteExport() {
	// This shouldn't be part of spritewell
	imgs := ImageList{
		ImageDir:  ".",
		BuildDir:  "test/build",
		GenImgDir: "test/build/img",
	}
	imgs.Decode("test/*.png")
	of, err := imgs.Combine()

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(of)

	// Output:
	// img/203b63.png
}

func TestSpriteDecode(t *testing.T) {
	var out bytes.Buffer
	log.SetOutput(&out)
	//Should fail with unable to find file
	i := ImageList{}
	i.Decode("notafile")
	_, err := i.Combine()
	if e := "png: invalid format: invalid image size: 0x0"; err.Error() != e {
		t.Errorf("Unexpected error thrown was: %s expected: %s",
			e, err)
	}

	if len(i.GoImages) > 0 {
		t.Errorf("Found a non-existant file")
	}
	log.SetOutput(os.Stdout)
}

func TestSpriteHorizontal(t *testing.T) {

	imgs := ImageList{}
	imgs.Pack = "horz"
	imgs.Decode("test/139.jpg", "test/140.jpg")
	imgs.Combine()

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

	imgs := ImageList{}
	imgs.Padding = 10
	imgs.Pack = "horz"
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

	imgs.Pack = "vert"
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
	imgs := ImageList{}
	log.SetOutput(&out)
	imgs.Decode("test/bad/interlace.png")
	imgs.Combine()
	_ = imgs
	out.ReadString('\n')
	str := out.String()
	strFirst := strings.Split(str, "\n")[0]
	if e := "png: unsupported feature: compression, " +
		"filter or interlace method"; e != strFirst {
		// No longer an error in 1.4+
		t.Skipf("Interlaced error not received expected:\n%s was:\n%s",
			e, strFirst)
	}

	if e := -1; imgs.Y(1) != e {
		t.Errorf("Invalid position expected: %d, was: %d", e, imgs.Y(1))
	}

	if e := -1; imgs.X(1) != e {
		t.Errorf("Invalid position expected: %d, was: %d", e, imgs.X(1))
	}

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
	imgs := ImageList{}
	imgs.Decode("test/*.png")
	str, err := imgs.OutputPath()
	if err != nil {
		t.Error(err)
	}

	if e := "image/803bf3.png"; e != str {
		t.Errorf("got: %s wanted: %s", str, e)
	}

	imgs.GenImgDir = "../build/img"
	imgs.BuildDir = "../build"
	str, err = imgs.OutputPath()
	if err != nil {
		t.Error(err)
	}

	if e := "img/203b63.png"; e != str {
		t.Errorf("got: %s wanted: %s", str, e)
	}

}

func TestMany(t *testing.T) {
	imgs := ImageList{}
	imgs.GenImgDir = "test/build"
	imgs.Pack = "vert"
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
