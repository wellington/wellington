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
Sprite_sass is designed to be portable, moduler, and extremely fast.  Go's flexiblity allows this tool to work with any number of input streams File, HTTP, or standard input.  Go is also easily compiled into many platforms, although there may be some limitations on where libsass will compile.

Sprite_sass uses libsass to process the resulting sass after it has translated all the spriting commands.  Libsass is extremely fast, so you should notice a significant build speed improvement.

#### Installation
Install Go and add $GOPATH/bin to your $PATH. [Detailed instructions](https://golang.org/doc/install)

```
go get -u github.com/drewwells/sprite_sass/cmd/sprite
cd $GOPATH/drewwells/sprite_sass
git submodule update --init --recursive
cd libsass
make
go install github.com/drewwells/sprite_sass/sprite
sprite // Should now be available in your path
```

### List of Available Commands
|Command Example|Description|
|-------------------------------------------------------------------|-------------------------------------------------|
|$images: *sprite-map*("glob/pattern");|Creates a reference to your sprites|
|height: *sprite-height*($images,"file");|Inserts the height of the sprite|
|width: *sprite-width*($images,"file");|Inserts the width of the sprite|
|background: *sprite*($images,"file");|Finds the requested file in the sprite sheet, extension is optional|
|@include *sprite-dimensions*($images,"file");|Creates height/width css for the container size|
|background-image: inline-image($images,"justone");|Base64 encoded data uri of the requested image|
|background: *image-url*("nopixel.png");|Creates a relative path to your image directory|
