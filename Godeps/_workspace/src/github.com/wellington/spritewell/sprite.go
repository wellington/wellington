package spritewell

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	mrand "math/rand"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

var formats = []string{".png", ".gif", ".jpg"}

type GoImages []image.Image
type ImageList struct {
	bytes.Buffer
	GoImages
	BuildDir, ImageDir, GenImgDir string
	Out                           draw.Image
	OutFile                       string
	Combined                      bool
	Globs, Paths                  []string
	Padding                       int    // Padding in pixels
	Pack                          string // default is vertical
}

// SafeImageMap provides a thread-safe data structure for
// creating maps of ImageLists
type SafeImageMap struct {
	sync.RWMutex
	M map[string]ImageList
}

func NewImageMap() *SafeImageMap {
	img := SafeImageMap{
		M: make(map[string]ImageList)}
	return &img
}

func funnyNames() string {

	names := []string{"White_and_Nerdy",
		"Fat",
		"Eat_It",
		"Foil",
		"Like_a_Surgeon",
		"Amish_Paradise",
		"The_Saga_Begins",
		"Polka_Face"}
	return names[mrand.Intn(len(names))]
}

// String generates CSS path to output file
func (l ImageList) String() string {
	paths := ""
	for _, path := range l.Paths {
		path += strings.TrimSuffix(filepath.Base(path),
			filepath.Ext(path)) + " "
	}
	return paths
}

// Return relative path to File
// TODO: Return abs path to file
func (l ImageList) File(f string) string {
	pos := l.Lookup(f)
	if pos > -1 {
		return l.Paths[pos]
	}
	return ""
}

func (l ImageList) Lookup(f string) int {
	var base string
	pos := -1
	for i, v := range l.Paths {
		base = filepath.Base(v)
		base = strings.TrimSuffix(base, filepath.Ext(v))
		if f == v {
			pos = i
			//Do partial matches, for now
		} else if f == base {
			pos = i
		}
	}

	if pos > -1 {
		if l.GoImages[pos] != nil {
			return pos
		}
	}
	// TODO: Find a better way to send these to cli so tests
	// aren't impacted.
	// Debug.Printf("File not found: %s\n Try one of %s", f, l)

	return -1
}

// Return the X position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) X(pos int) int {
	p := l.GetPack(pos)
	return p.X
}

// Return the Y position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) Y(pos int) int {
	p := l.GetPack(pos)
	return p.Y
}

func (l ImageList) SImageWidth(s string) int {
	if pos := l.Lookup(s); pos > -1 {
		return l.ImageWidth(pos)
	}
	return -1
}

func (l ImageList) ImageWidth(pos int) int {
	if pos > len(l.GoImages) || pos < 0 {
		return -1
	}

	return l.GoImages[pos].Bounds().Dx()
}

func (l ImageList) SImageHeight(s string) int {
	if pos := l.Lookup(s); pos > -1 {
		return l.ImageHeight(pos)
	}
	return -1
}

func (l ImageList) ImageHeight(pos int) int {
	if pos > len(l.GoImages) || pos < 0 {
		return -1
	}
	return l.GoImages[pos].Bounds().Dy()
}

// Dimensions is the total W,H pixels of the generate sprite
func (l *ImageList) Dimensions() Pos {
	// Size of array + 1 gets the dimensions of the entire sprite
	return l.GetPack(len(l.GoImages))
}

// OutputPath generates a unique filename based on the relative path
// from image directory to build directory and the files matched in
// the glob lookup.  OutputPath is not cache safe.
func (l *ImageList) OutputPath() (string, error) {
	path, err := filepath.Rel(l.BuildDir, l.GenImgDir)
	if err != nil {
		return "", err
	}
	if path == "." {
		path = "image"
	}

	hasher := md5.New()
	hasher.Write([]byte(path + strings.Join(l.Globs, "|")))
	salt := hex.EncodeToString(hasher.Sum(nil))[:6]
	l.OutFile = filepath.Join(path, salt+".png")
	return l.OutFile, nil
}

// Decode accepts a variable number of glob patterns.  The ImageDir
// is assumed to be prefixed to the globs provided.
func (l *ImageList) Decode(rest ...string) error {

	// Invalidate the composite cache
	l.Out = nil
	var (
		paths []string
	)
	absImageDir, _ := filepath.Abs(l.ImageDir)
	for _, r := range rest {
		matches, err := filepath.Glob(filepath.Join(l.ImageDir, r))
		if err != nil {
			panic(err)
		}
		if len(matches) == 0 {
			// No matches found, try appending * and trying again
			// This supports the case "139" > "139.jpg" "139.png" etc.
			matches, err = filepath.Glob(filepath.Join(l.ImageDir, r+"*"))
			if err != nil {
				panic(err)
			}
		}
		rel := make([]string, len(matches))
		for i := range rel {
			// Attempt both relative and absolute to path
			if p, err := filepath.Rel(l.ImageDir, matches[i]); err == nil {
				rel[i] = p
			} else if p, err := filepath.Rel(absImageDir, matches[i]); err == nil {
				rel[i] = p
			}
		}
		l.Paths = append(l.Paths, rel...)
		paths = append(paths, matches...)
	}
	// turn paths into relative paths to the files

	l.Globs = paths
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		goimg, _, err := image.Decode(f)
		if err != nil {
			ext := filepath.Ext(path)
			if !CanDecode(ext) {
				return fmt.Errorf("format: %s not supported", ext)
			} else {
				return fmt.Errorf("Error processing: %s\n%s", path, err)
			}
		}
		l.GoImages = append(l.GoImages, goimg)
	}

	if len(l.Paths) == 0 {
		return fmt.Errorf("No images were found for pattern: %v",
			rest,
		)
	}

	return nil
}

// CanDecode checks if the file extension is supported by
// spritewell.
func CanDecode(ext string) bool {
	for i := range formats {
		if ext == formats[i] {
			return true
		}
	}
	return false
}

// Combine all images in the slice into a final output
// image.
func (l *ImageList) Combine() (string, error) {

	var (
		maxW, maxH int
	)

	if l.Out != nil {
		return "", errors.New("Sprite is empty")
	}

	pos := l.Dimensions()
	maxW, maxH = pos.X, pos.Y

	goimg := image.NewRGBA(image.Rect(0, 0, maxW, maxH))
	l.Out = goimg
	for i := range l.GoImages {

		pos := l.GetPack(i)

		draw.Draw(goimg, goimg.Bounds(), l.GoImages[i],
			image.Point{
				X: -pos.X,
				Y: -pos.Y,
			}, draw.Src)
	}
	l.Combined = true

	// Set the buf so bytes.Buffer works
	err := png.Encode(&l.Buffer, goimg)
	if err != nil {
		return "", err
	}
	return l.OutputPath()
}

// Pos represents the x, y coordinates of an image
// in the sprite sheet.
type Pos struct {
	X, Y int
}

// GetPack retrieves the Pos of an image in the
// list of images.
// TODO: Changing l.Pack will update the positions, but
// the sprite file will need to be regenerated via Decode.
func (l *ImageList) GetPack(pos int) Pos {
	// Default is vertical
	if l.Pack == "vert" {
		return l.PackVertical(pos)
	} else if l.Pack == "horz" {
		return l.PackHorizontal(pos)
	}
	return l.PackVertical(pos)
}

// PackVertical finds the Pos for a vertically packed sprite
func (l *ImageList) PackVertical(pos int) Pos {
	if pos == -1 || pos == 0 {
		return Pos{0, 0}
	}
	var x, y int
	var rect image.Rectangle
	// there are n-1 paddings in an image list
	y = l.Padding * (pos)
	// No padding on the outside of the image
	if pos == len(l.GoImages) {
		y -= l.Padding
	}
	for i := 1; i <= pos; i++ {
		rect = l.GoImages[i-1].Bounds()
		y += rect.Dy()
		if pos == len(l.GoImages) {
			x = int(math.Max(float64(x), float64(rect.Dx())))
		}
	}

	return Pos{
		x, y,
	}
}

// PackHorzontal finds the Pos for a horizontally packed sprite
func (l *ImageList) PackHorizontal(pos int) Pos {
	if pos == -1 || pos == 0 {
		return Pos{0, 0}
	}
	var x, y int
	var rect image.Rectangle

	// there are n-1 paddings in an image list
	x = l.Padding * pos
	// No padding on the outside of the image
	if pos == len(l.GoImages) {
		x -= l.Padding
	}
	for i := 1; i <= pos; i++ {
		rect = l.GoImages[i-1].Bounds()
		x += rect.Dx()
		if pos == len(l.GoImages) {
			y = int(math.Max(float64(y), float64(rect.Dy())))
		}
	}

	return Pos{
		x, y,
	}
}

func randString(n int) string {
	const alphanum = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, n)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}

// Export saves out the ImageList to the specified file
func (l *ImageList) Export() (string, error) {
	// Use the auto generated path if none is specified
	// TODO: Differentiate relative file path (in css) to this abs one
	opath, err := l.OutputPath()
	if err != nil {
		return "", err
	}
	os.MkdirAll(filepath.Dir(opath), 0755)
	abs := filepath.Join(l.GenImgDir, filepath.Base(opath))
	fo, err := os.Create(abs)
	if err != nil {
		if _, err := os.Stat(abs); err == nil {
			return abs, nil
		}
		return "", err
	}
	if err != nil {
		log.Printf("Failed to create file: %s\n", abs)
		log.Print(err)
		return "", err
	}
	//This call is cached if already run
	l.Combine()
	defer fo.Close()

	n, err := io.Copy(fo, &l.Buffer)
	if n == 0 {
		log.Fatalf("no bytes written of: %d", l.Buffer.Len())
	}
	if err != nil {
		log.Printf("Failed to create: %s\n%s", abs, err)
		return "", err
	}
	log.Print("Created sprite: ", abs)
	return abs, nil
}
