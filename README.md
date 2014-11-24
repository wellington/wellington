[![wercker status](https://app.wercker.com/status/0e2b532c6e35225334fdeeac0cbb7831/m/master "wercker status")](https://app.wercker.com/project/bykey/0e2b532c6e35225334fdeeac0cbb7831)
[![Coverage Status](https://img.shields.io/coveralls/drewwells/sprite_sass.svg)](https://coveralls.io/r/drewwells/sprite_sass?branch=master)

sprite-sass
===========

Sprite_sass adds the missing pieces to SASS, spriting and image manipulation.  This tool is designed to work directly from sass, so you don't need to learn an entire new DSL just add a few new commands to get toolbox.

```
$images: sprite-map("sprites/*.png");
div {
	@include sprite-dimensions($images, "cat");
	background: sprite($images, "cat");
```
// Generates
```
div {
	width: 140px;
	height: 79px;
	background: url("genimg/sprites-wehqi.png");
}
```
### Why?
Sprite_sass adds file awareness to the sass language.  It has been written in Go for portability, modularity, and speed.  There are no dependencies on Ruby.  The binary includes everything you need to run sprite_sass.  Sprite_sass uses libsass under the covers for processing the output CSS.

#### Installation
Check out the releases for compiled binaries

#### Building from source
Install Go and add $GOPATH/bin to your $PATH. [Detailed instructions](https://golang.org/doc/install)

You must have libsass installed to build this project.  Do so by checking
out the repo or building [libsass](https://github.com/sass/libsass) via the  instruction in the repo.

```
# This will fail if you don't have libsass installed, that's OK.
go get -u github.com/drewwells/sprite_sass/sprite
cd $GOPATH/drewwells/sprite_sass
make deps
# Attempt install again
go get -u github.com/drewwells/sprite_sass/sprite
```

Test out if the installation worked
```
sprite
```

### List of Available Commands
|Command Example|Description|
|-------------------------------------------------------------------|-------------------------------------------------|
|$images: *sprite-map*("glob/pattern");|Creates a reference to your sprites|
|$map: *sprite-file*($images,"file");|Returns a map for use with image-[width|height]|
|height: *image-height*($images,"file");|Inserts the height of the sprite|
|width: *image-width*($images,"file");|Inserts the width of the sprite|
|background: *sprite*($images,"file");|Finds the requested file in the sprite sheet, extension is optional|
|@include *sprite-dimensions*($images,"file");|Creates height/width css for the container size|
|background-image: inline-image($images,"justone");|Base64 encoded data uri of the requested image|
|background: *image-url*("nopixel.png");|Creates a relative path to your image directory|
