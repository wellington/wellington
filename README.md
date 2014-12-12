[![wercker status](https://app.wercker.com/status/873d0c7929b8b1e8bc37bcc16829fb5f/m/master "wercker status")](https://app.wercker.com/project/bykey/873d0c7929b8b1e8bc37bcc16829fb5f)
[![Coverage Status](https://coveralls.io/repos/wellington/wellington/badge.png)](https://coveralls.io/r/wellington/wellington)

Wellington
===========

Wellington adds the missing pieces to SASS, spriting and image manipulation.  This tool is designed to work directly from sass, so you don't need to learn an entire new DSL just add a few new commands to get toolbox.

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
Wellington adds file awareness to the sass language.  It has been written in Go for portability, modularity, and speed.  There are no dependencies on Ruby.  The binary includes everything you need to run sprite_sass.  Sprite_sass uses libsass under the covers for processing the output CSS.

#### Installation
Check out the releases for compiled binaries

#### Building from source
Install Go and add $GOPATH/bin to your $PATH. [Detailed instructions](https://golang.org/doc/install)

You must have libsass installed to build this project.  Do so by checking
out the repo or building [libsass](https://github.com/sass/libsass) via the  instruction in the repo.

```
# This will fail if you don't have libsass installed, that's OK.
go get -u github.com/wellington/wellington/wt
cd $GOPATH/wellington/wellington
make deps
# Attempt install again
go get -u github.com/wellington/wellington/wt
```

Test out if the installation worked
```
wt
```

### List of Available Commands
|Command Example|Description|
|-------------------------------------------------------------------|-------------------------------------------------|
|$images: *sprite-map*("glob/pattern", $spacing: 10px);|Creates a reference to your sprites|
|$map: *sprite-file*($spritemap,"file");|Returns encoded data only useful for passing to image-height, image-width|
|height: *image-height*("image.png");|Inserts the height of the sprite|
|width: *image-width*("image.png");|Inserts the width of the sprite|
|@include *sprite-dimensions*($images,"file");|Creates height/width css for the container size|
|background-image: inline-image($images,"justone");|Base64 encoded data uri of the requested image|
|background: *image-url*("nopixel.png");|Returns a relative path to an image in the image directory|
|font-url: *font-url*("arial.eot", $raw);|Returns a relative path to a file in the font directory, optionally do not wrap in url()|
|*sprite*($map,"file")|Returns the path and background position of an image for use with background:|

### Development

Get the code

	go get github.com/wellington/wellington

You may want to cd `$GOPATH/wellington/wellington` and set the origin to your fork.

	git remote rm origin
	git remote add origin git@github.com:username/wellington.git

Testing

    make test

Profiling

	make profile

Docker Container

	make build
	make docker #launch a container

Please use pull requests for contributing code.  Wercker will automatically test and lint your contributions.  Thanks for helping!

### License

Wellington is licensed under MIT.
