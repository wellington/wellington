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
	"strconv"
	"strings"
	"sync"

	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

var ErrNoImages = errors.New("no images matched for pattern")

var formats = []string{".png", ".gif", ".jpg"}

type GoImages []image.Image

type Sprite struct {
	buf bytes.Buffer

	optsMu sync.RWMutex
	opts   *Options

	goImagesMu sync.RWMutex
	len        int
	GoImages   GoImages

	outFileMu sync.Mutex
	outFile   string

	combineMu sync.Mutex
	Combined  bool

	globMu       sync.RWMutex
	globs, paths []string

	// Channels to do work
	queue    chan work
	combined chan result

	// Done notifies caller that sprite is written to disk
	done chan error
}

type Options struct {
	BuildDir, ImageDir, GenImgDir string
	Pack                          string
	Padding                       int // Padding in pixels
}

func New(opts *Options) *Sprite {
	if opts == nil {
		opts = &Options{}
	}
	l := &Sprite{
		queue:    make(chan work),
		combined: make(chan result),
		opts:     opts,
		done:     make(chan error),
	}
	go l.loopAndCombine(l.queue, l.combined)
	return l
}

type work struct {
	imgs GoImages
	pos  Pos
	pack string
}

type result struct {
	buf *bytes.Buffer
	err error
}

// SafeImageMap provides a thread-safe data structure for
// creating maps of ImageLists
type SafeImageMap struct {
	sync.RWMutex
	M map[string]*Sprite
}

func NewImageMap() *SafeImageMap {
	img := SafeImageMap{
		M: make(map[string]*Sprite)}
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
func (l *Sprite) String() string {
	paths := ""
	l.globMu.RLock()
	defer l.globMu.RUnlock()
	for _, path := range l.paths {
		path += strings.TrimSuffix(filepath.Base(path),
			filepath.Ext(path)) + " "
	}
	return paths
}

func (l *Sprite) Paths() []string {
	l.globMu.RLock()
	defer l.globMu.RUnlock()
	return l.paths
}

// Return relative path to File
// TODO: Return abs path to file
func (l *Sprite) File(f string) string {
	l.globMu.RLock()
	defer l.globMu.RUnlock()
	pos := l.Lookup(f)
	if pos > -1 {
		return l.paths[pos]
	}
	return ""
}

func (l *Sprite) Len() int {
	l.goImagesMu.RLock()
	defer l.goImagesMu.RUnlock()
	return l.len
}

func (l *Sprite) Lookup(f string) int {
	var base string
	pos := -1
	l.globMu.RLock()
	for i, v := range l.paths {
		base = filepath.Base(v)
		base = strings.TrimSuffix(base, filepath.Ext(v))
		if f == v {
			pos = i
			//Do partial matches, for now
		} else if f == base {
			pos = i
		}
	}
	l.globMu.RUnlock()

	return pos

	// TODO: what's this supposed to be doing?
	// if pos > -1 {
	// 	l.goImagesMu.RLock()
	// 	if l.GoImages[pos] != nil {
	// 		l.goImagesMu.RUnlock()
	// 		return pos
	// 	}
	// }

}

// Return the X position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l *Sprite) X(pos int) int {
	p := l.GetPack(pos)
	return p.X
}

// Return the Y position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l *Sprite) Y(pos int) int {
	p := l.GetPack(pos)
	return p.Y
}

func (l *Sprite) SImageWidth(s string) int {
	if pos := l.Lookup(s); pos > -1 {
		return l.ImageWidth(pos)
	}
	return -1
}

func (l *Sprite) ImageWidth(pos int) int {
	if pos > l.Len() || pos < 0 {
		return -1
	}
	l.goImagesMu.RLock()
	defer l.goImagesMu.RUnlock()
	return l.GoImages[pos].Bounds().Dx()
}

func (l *Sprite) SImageHeight(s string) int {
	if pos := l.Lookup(s); pos > -1 {
		return l.ImageHeight(pos)
	}
	return -1
}

func (l *Sprite) ImageHeight(pos int) int {
	if pos > l.Len() || pos < 0 {
		return -1
	}
	l.goImagesMu.RLock()
	defer l.goImagesMu.RUnlock()
	return l.GoImages[pos].Bounds().Dy()
}

// Dimensions is the total W,H pixels of the generate sprite
func (l *Sprite) Dimensions() Pos {
	// Size of array + 1 gets the dimensions of the entire sprite
	return l.GetPack(l.Len())
}

// OutputPath generates a unique filename based on the relative path
// from image directory to build directory and the files matched in
// the glob lookup.  OutputPath is not cache safe.
func (l *Sprite) OutputPath() (string, error) {
	l.outFileMu.Lock()
	defer l.outFileMu.Unlock()
	if len(l.outFile) > 0 {
		return l.outFile, nil
	}

	l.optsMu.RLock()
	path, err := filepath.Rel(l.opts.BuildDir, l.opts.GenImgDir)
	pack := l.opts.Pack
	padding := l.opts.Padding
	l.optsMu.RUnlock()
	if err != nil {
		return "", err
	}
	// TODO: remove this
	if path == "." {
		path = "image"
	}

	hasher := md5.New()
	l.globMu.RLock()
	if len(l.globs) == 0 {
		return "", ErrNoImages
	}
	seed := pack + strconv.Itoa(padding) + "|" + filepath.ToSlash(path+strings.Join(l.globs, "|"))
	l.globMu.RUnlock()
	hasher.Write([]byte(seed))
	salt := hex.EncodeToString(hasher.Sum(nil))[:6]
	l.outFile = filepath.Join(path, salt+".png")
	return l.outFile, nil
}

// Decode accepts a variable number of glob patterns.  The ImageDir
// is assumed to be prefixed to the globs provided.
func (l *Sprite) Decode(rest ...string) error {

	// Invalidate the composite cache
	var (
		paths []string
		rels  []string
	)
	l.optsMu.RLock()
	absImageDir, _ := filepath.Abs(l.opts.ImageDir)
	relImageDir := l.opts.ImageDir
	l.optsMu.RUnlock()
	for _, r := range rest {
		matches, err := filepath.Glob(filepath.Join(relImageDir, r))
		if err != nil {
			panic(err)
		}
		if len(matches) == 0 {
			// No matches found, try appending * and trying again
			// This supports the case "139" > "139.jpg" "139.png" etc.
			matches, err = filepath.Glob(filepath.Join(relImageDir, r+"*"))
			if err != nil {
				panic(err)
			}
		}
		rel := make([]string, len(matches))
		for i := range rel {
			// Attempt both relative and absolute to path
			if p, err := filepath.Rel(relImageDir, matches[i]); err == nil {
				rel[i] = p
			} else if p, err := filepath.Rel(absImageDir, matches[i]); err == nil {
				rel[i] = p
			}
		}
		rels = append(rels, rel...)
		paths = append(paths, matches...)
	}
	// turn paths into relative paths to the files
	l.globMu.Lock()
	l.paths = rels
	l.globs = paths
	l.globMu.Unlock()

	l.goImagesMu.Lock()
	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			return err
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
	l.len = len(l.GoImages)
	l.goImagesMu.Unlock()

	if len(l.paths) == 0 {
		return ErrNoImages
	}
	l.queue <- work{pos: l.Dimensions(), imgs: l.GoImages}
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

func (l *Sprite) loopAndCombine(queue chan work, resp chan result) {
	for {
		select {
		case work := <-queue:
			imgs := work.imgs
			pos := work.pos
			maxW, maxH := pos.X, pos.Y
			l.combineMu.Lock()
			defer l.combineMu.Unlock()
			goimg := image.NewRGBA(image.Rect(0, 0, maxW, maxH))

			for i := 0; i < len(imgs); i++ {
				pos := l.GetPack(i)
				draw.Draw(goimg, goimg.Bounds(), imgs[i],
					image.Point{
						X: -pos.X,
						Y: -pos.Y,
					}, draw.Src)
			}

			buf := new(bytes.Buffer)
			// Set the buf so bytes.Buffer works
			err := png.Encode(buf, goimg)
			if err != nil {
				log.Fatal(err)
			}
			resp <- result{buf: buf, err: err}
		}
	}
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
func (l *Sprite) GetPack(pos int) Pos {
	l.optsMu.RLock()
	pack := l.opts.Pack
	l.optsMu.RUnlock()
	// Default is vertical
	if pack == "vert" {
		return l.PackVertical(pos)
	} else if pack == "horz" {
		return l.PackHorizontal(pos)
	}
	return l.PackVertical(pos)
}

// PackVertical finds the Pos for a vertically packed sprite
func (l *Sprite) PackVertical(pos int) Pos {
	if pos == -1 || pos == 0 {
		return Pos{0, 0}
	}
	var x, y int
	var rect image.Rectangle
	l.optsMu.RLock()
	padding := l.opts.Padding
	l.optsMu.RUnlock()
	// there are n-1 paddings in an image list
	y = padding * (pos)
	// No padding on the outside of the image
	numimages := l.Len()
	if pos == numimages {
		y -= padding
	}
	for i := 1; i <= pos; i++ {
		l.goImagesMu.RLock()
		rect = l.GoImages[i-1].Bounds()
		l.goImagesMu.RUnlock()
		y += rect.Dy()
		if pos == numimages {
			x = int(math.Max(float64(x), float64(rect.Dx())))
		}
	}

	return Pos{
		x, y,
	}
}

// PackHorzontal finds the Pos for a horizontally packed sprite
func (l *Sprite) PackHorizontal(pos int) Pos {
	if pos == -1 || pos == 0 {
		return Pos{0, 0}
	}
	var x, y int
	var rect image.Rectangle
	l.optsMu.RLock()
	padding := l.opts.Padding
	l.optsMu.RUnlock()

	// there are n-1 paddings in an image list
	x = padding * pos
	// No padding on the outside of the image
	numimages := l.Len()
	if pos == numimages {
		x -= padding
	}
	for i := 1; i <= pos; i++ {
		l.goImagesMu.RLock()
		rect = l.GoImages[i-1].Bounds()
		l.goImagesMu.RUnlock()
		x += rect.Dx()
		if pos == numimages {
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

func (l *Sprite) export() (*os.File, string, error) {
	// Use the auto generated path if none is specified
	// TODO: Differentiate relative file path (in css) to this abs one
	opath, err := l.OutputPath()
	if err != nil {
		return nil, "", err
	}
	os.MkdirAll(filepath.Dir(opath), 0755)
	l.optsMu.RLock()
	abs := filepath.Join(l.opts.GenImgDir, filepath.Base(opath))
	l.optsMu.RUnlock()
	fo, err := os.Create(abs)
	if err != nil {
		if _, err := os.Stat(abs); err == nil {
			return nil, abs, nil
		}
		return nil, "", err
	}
	return fo, abs, err
}

// Export returns the output path of the combined sprite and flushes
// the sprite to disk. This method does not block on disk I/O. See Wait
func (s *Sprite) Export() (path string, err error) {
	of, path, err := s.export()
	if err != nil {
		return
	}
	if of == nil {
		err = errors.New("output file is nil")
		return
	}

	go func(combined chan result, done chan error, of *os.File) {
		// We're good for output file location, listen for combining success
		result := <-combined
		if result.err != nil {
			done <- err
			return
		}
		err := writeToDisk(of, result.buf)
		if err != nil {
			done <- err
			return
		}
		// succeeded in writing sprite
		done <- nil
	}(s.combined, s.done, of)

	return
}

// Wait blocks until sprite is encoded to memory and flushed to disk.
func (s *Sprite) Wait() error {
	return <-s.done
}

var ErrFailedToWrite = errors.New("failed to write sprite to disk")

func writeToDisk(of *os.File, buf *bytes.Buffer) error {
	n, err := io.Copy(of, buf)
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("failed to write file: %s", of.Name())
	}
	log.Print("Created sprite: ", of.Name())
	return nil
}
