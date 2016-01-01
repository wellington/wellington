spritewell
==========

[![Circle CI](https://circleci.com/gh/wellington/spritewell.svg?style=svg)](https://circleci.com/gh/wellington/spritewell)

Spritewell performs image composition on a glob of source images. This is useful for creating spritesheets of images. This is a thread safe library and is optimized for multicore systems.


Documentation available at: http://godoc.org/github.com/wellington/spritewell

Currently two different types of positioning are available, Horizontal and Vertical.  Padding between images is also supported.

This project does the heavily lifting of image processing for [Wellington](http://getwt.io).

To use spritewell: http://godoc.org/github.com/wellington/spritewell#example-Sprite

```
imgs := New(&Options{
    ImageDir:  ".",
    BuildDir:  "test/build",
    GenImgDir: "test/build/img",
})

imgs.Decode("test/*.png")
of, _ := imgs.OutputPath()
fmt.Println(of)

// Calls are non-blocking, use Wait() to ensure image encoding has
// completed and results are flushed to disk.
imgs.Wait()
```
