package sprite_sass

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"image"
	"io"
	"log"
	"math"
	mrand "math/rand"
	"os"
	"path/filepath"
	"strings"

	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
)

type GoImages []image.Image
type ImageList struct {
	GoImages
	BuildDir, ImageDir, GenImgDir string
	Out                           draw.Image
	OutFile                       string
	Combined                      bool
	Files                         []string
	Vertical                      bool
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

func (l ImageList) String() string {
	files := ""
	for _, file := range l.Files {
		files += strings.TrimSuffix(filepath.Base(file),
			filepath.Ext(file)) + " "
	}
	return files
}

// Return relative path to File
// TODO: Return abs path to file
func (l ImageList) File(f string) string {
	pos := l.Lookup(f)
	if pos > -1 {
		return l.Files[pos]
	}
	return ""
}

func (l ImageList) Lookup(f string) int {
	var base string
	pos := -1
	for i, v := range l.Files {
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
	x := 0
	if pos > len(l.GoImages) {
		return -1
	}
	if l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		x += l.ImageWidth(i)
	}
	return x
}

// Return the Y position of an image based
// on the layout (vertical/horizontal) and
// position in Image slice
func (l ImageList) Y(pos int) int {
	y := 0
	if pos > len(l.GoImages) {
		return -1
	}
	if !l.Vertical {
		return 0
	}
	for i := 0; i < pos; i++ {
		y += l.ImageHeight(i)
	}
	return y
}

func (l ImageList) Map(name string) string {
	var res []string
	for i := range l.GoImages {
		base := strings.TrimSuffix(filepath.Base(l.Files[i]),
			filepath.Ext(l.Files[i]))
		res = append(res, fmt.Sprintf(
			"%s: map_merge(%s,(%s: (width: %d, height: %d, "+
				"x: %d, y: %d, url: '%s')))",
			name, name,
			base, l.ImageWidth(i), l.ImageHeight(i),
			l.X(i), l.Y(i), filepath.Join(l.GenImgDir, l.OutFile),
		))
	}
	return "(); " + strings.Join(res, "; ") + ";"
}

func (l ImageList) CSS(s string) string {
	pos := l.Lookup(s)
	if pos == -1 {
		log.Printf("File not found: %s\n Try one of: %s",
			s, l)
		return ""
	}

	return fmt.Sprintf(`url("%s") %s`,
		l.OutFile, l.Position(s))
}

func (l ImageList) Position(s string) string {
	pos := l.Lookup(s)
	if pos == -1 {
		log.Printf("File not found: %s\n Try one of: %s",
			s, l)
		return ""
	}

	return fmt.Sprintf(`%dpx %dpx`, -l.X(pos), -l.Y(pos))
}

func (l ImageList) Dimensions(s string) string {
	if pos := l.Lookup(s); pos > -1 {

		return fmt.Sprintf("width: %dpx;\nheight: %dpx",
			l.ImageWidth(pos), l.ImageHeight(pos))
	}
	return ""
}

func (l ImageList) inline() []byte {

	r, w := io.Pipe()
	go func(w io.WriteCloser) {
		err := png.Encode(w, l.GoImages[0])
		if err != nil {
			panic(err)
		}
		w.Close()
	}(w)
	var scanned []byte
	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanBytes)
	for scanner.Scan() {
		scanned = append(scanned, scanner.Bytes()...)
	}
	return scanned
}

// Inline creates base64 encoded string of the underlying
// image data blog
func (l ImageList) Inline() string {
	encstr := base64.StdEncoding.EncodeToString(l.inline())
	return fmt.Sprintf("url('data:image/png;base64,%s')", encstr)
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

// Return the cumulative Height of the
// image slice.
func (l *ImageList) Height() int {
	h := 0
	ll := *l

	for pos, _ := range ll.GoImages {
		if l.Vertical {
			h += ll.ImageHeight(pos)
		} else {
			h = int(math.Max(float64(h), float64(ll.ImageHeight(pos))))
		}
	}
	return h
}

// Return the cumulative Width of the
// image slice.
func (l *ImageList) Width() int {
	w := 0

	for pos, _ := range l.GoImages {
		if !l.Vertical {
			w += l.ImageWidth(pos)
		} else {
			w = int(math.Max(float64(w), float64(l.ImageWidth(pos))))
		}
	}
	return w
}

// Build an output file location based on
// [genimagedir|location of file matched by glob] + glob pattern
func (l *ImageList) OutputPath(globpath string) error {

	path := filepath.Dir(globpath)
	if path == "." {
		path = "image"
	}
	path = strings.Replace(path, "/", "", -1)
	ext := filepath.Ext(globpath)

	// Encode the image so the bytestring can feed into md5.Sum
	var b bytes.Buffer
	err := png.Encode(&b, l.Out)
	if err != nil {
		// Image is empty or has errors, generate a funny filename
		l.OutFile = funnyNames() + "-" + randString(2) + ".png"
		return err
	}
	// Remove invalid characters from path
	path = strings.Replace(path, "*", "", -1)
	hasher := md5.New()
	hasher.Write(b.Bytes())
	salt := hex.EncodeToString(hasher.Sum(nil))[:6]
	l.OutFile += path + "-" + salt + ext
	return nil
}

// Accept a variable number of image globs appending
// them to the ImageList.
func (l *ImageList) Decode(rest ...string) error {

	// Invalidate the composite cache
	l.Out = nil
	var (
		paths []string
	)

	for _, r := range rest {
		matches, err := filepath.Glob(filepath.Join(l.ImageDir, r))
		if err != nil {
			panic(err)
		}
		paths = append(paths, matches...)
	}

	for _, path := range paths {
		f, err := os.Open(path)
		if err != nil {
			panic(err)
		}
		goimg, str, err := image.Decode(f)
		_ = str // Image format ie. png
		if err != nil {
			// Print errors for images that are not currently supported
			// by Go's image library
			log.Printf("Error processing: %s\n%s", path, err)
			continue
		}
		l.GoImages = append(l.GoImages, goimg)
		l.Files = append(l.Files, path)
	}

	if len(l.Files) == 0 {
		log.Printf("No images were found for glob: %v",
			rest,
		)
	}

	// Combine images so that md5 hash of filename can be created
	l.Combine()
	// Send first glob as definition for output path
	return l.OutputPath(rest[0])
}

// Combine all images in the slice into a final output
// image.
func (l *ImageList) Combine() {

	var (
		maxW, maxH int
	)

	if l.Out != nil {
		return
	}

	maxW, maxH = l.Width(), l.Height()

	curH, curW := 0, 0

	goimg := image.NewRGBA(image.Rect(0, 0, maxW, maxH))
	l.Out = goimg
	for _, img := range l.GoImages {

		draw.Draw(goimg, goimg.Bounds(), img,
			image.Point{
				X: curW,
				Y: curH,
			}, draw.Src)

		if l.Vertical {
			curH -= img.Bounds().Dy()
		} else {
			curW -= img.Bounds().Dx()
		}
	}

	l.Combined = true
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
	abs := filepath.Join(l.GenImgDir, filepath.Base(l.OutFile))
	// Create directory if it doesn't exist
	err := os.MkdirAll(filepath.Dir(abs), 0755)
	if err != nil {
		log.Printf("Failed to create image build dir: %s",
			filepath.Dir(abs))
		return "", err
	}
	fo, err := os.Create(abs)
	if err != nil {
		log.Printf("Failed to create file: %s\n", abs)
		log.Print(err)
		return "", err
	}
	//This call is cached if already run
	l.Combine()
	defer fo.Close()

	err = png.Encode(fo, l.Out)
	if err != nil {
		log.Printf("Failed to create: %s\n%s", abs, err)
		return "", err
	}
	log.Print("Created file: ", abs)
	return abs, nil
}
