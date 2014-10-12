package sprite_sass

import (
	"bytes"
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

	if e := ""; e != imgs.Dimensions("150") {
		t.Errorf("Non-existant image width invalid"+
			"\n    was:%s\nexpected:%s",
			imgs.Dimensions("150"), e)
	}

	//Quick cache check
	imgs.Combine()
	if imgs.Height() != 279 || imgs.Width() != 96 {
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

// Test image dimension calls
func TestSpriteImageDimensions(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/*.png")

	if e := "width: 96px;\nheight: 139px"; e != imgs.Dimensions("139") {
		t.Errorf("Dimensions invalid was: %s\nexpected: %s\n",
			imgs.Dimensions("139"), e)
	}

	if e := 139; e != imgs.SImageHeight("139") {
		t.Errorf("Height invalid    was:%d\nexpected:%d",
			imgs.SImageHeight("139"), e)
	}

	if e := 96; e != imgs.SImageWidth("139") {
		t.Errorf("Height invalid was:%d\nexpected:%d",
			imgs.SImageWidth("139"), e)
	}

	if e := "-96px 0px"; imgs.Position("140") != e {
		t.Errorf("Invalid position found was: %s\nexpected:%s",
			imgs.Position("140"), e)
	}

	output := imgs.CSS("140")
	if e := `url("test-585dca.png") -96px 0px`; output != e {
		t.Errorf("Invalid CSS generated on test     was: %s\nexpected: %s",
			output, e)
	}

}

//Test file globbing
func TestSpriteGlob(t *testing.T) {
	imgs := ImageList{
		ImageDir: "test/img",
	}
	imgs.Decode("*.png")

	// Test [Un]successful lookups
	if f := imgs.Lookup("test/img/139.png"); f != 0 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 0)
	}

	if f := imgs.Lookup("test/img/140.png"); f != 1 {
		t.Errorf("Invalid file location given found %d, expected %d", f, 1)
	}

	if f := imgs.Lookup("notafile.png"); f != -1 {
		t.Errorf("Found a file that doesn't exist")
	}
}

func TestSpriteExport(t *testing.T) {
	imgs := ImageList{
		ImageDir:  ".",
		GenImgDir: "build/test",
	}
	imgs.Decode("test/img/*.png")
	of := imgs.OutFile

	if e := "testimg-d65510.png"; e != of {
		t.Errorf("Outfile misnamed \n     was: %s\nexpected: %s", of, e)
	}
}

func TestSpriteDecode(t *testing.T) {
	//Should fail with unable to find file
	i := ImageList{}
	err := i.Decode("notafile")

	if e := "png: invalid format: invalid image size: 0x0"; err.Error() != e {
		t.Errorf("Unexpected error thrown was: %s expected: %s",
			e, err)
	}

	if len(i.GoImages) > 0 {
		t.Errorf("Found a non-existant file")
	}
}

func TestSpriteHorizontal(t *testing.T) {

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

func TestSpriteInline(t *testing.T) {
	imgs := ImageList{}
	imgs.Decode("test/pixel.png")
	imgs.Combine()
	bytes := imgs.inline()

	// Bytes are non-deterministic, so check length and move on
	if len(bytes) != 73 {
		t.Errorf("Pixel blog data had an invalid length"+
			"\n     was: %d\nexpected: 300-350", len(bytes))
	}

	str := imgs.Inline()
	if len(str) != 129 {
		t.Errorf("CSS length has an invalid length:%d expected: 400-500",
			len(str))
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
		t.Errorf("Interlaced error not received expected:\n%s was:\n%s",
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

	out.Reset()
	if e := ""; imgs.CSS("") != e {
		t.Errorf("Invalid css for non-file expected: %s, was: %s",
			e, imgs.CSS(""))
	}
	out.Reset()
	if e := ""; imgs.Position("") != e {
		t.Errorf("Invalid css for non-file expected: %s, was: %s",
			e, imgs.Position(""))
	}
	log.SetOutput(os.Stdout)
}
