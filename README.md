[![Circle CI](https://circleci.com/gh/wellington/wellington/tree/master.svg?style=svg)](https://circleci.com/gh/wellington/wellington/tree/master)
[![Coverage Status](https://coveralls.io/repos/wellington/wellington/badge.png?branch=master)](https://coveralls.io/r/wellington/wellington?branch=master)
[![Appveyor](https://ci.appveyor.com/api/projects/status/1apfkxe0369ce26d/branch/master?svg=true)](https://ci.appveyor.com/project/drewwells/wellington/branch/master)


Wellington
===========

[![Join the chat at https://gitter.im/wellington/wellington](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/wellington/wellington?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Wellington adds spriting to the lightning fast [libsass](http://libsass.org/). No need to learn a new tool, this all happens right in your Sass!

### Speed Matters

Benchmarks
```
# 40,000 line of code Sass project with 1200 images
wt         3.679s
compass   73.800s
# 20x faster!
```

For more benchmarks, see [realbench](https://github.com/wellington/realbench#results-early-2015-macbook-pro)

#### What it does

wt is a Sass preprocessor tool geared towards projects written in Sass. It focuses on tasks that make working on a Sass site friendlier and much faster. wt extends the Sass language to include spriting and image operations not currently possible in the core language.

```
$images: sprite-map("sprites/*.png");
div {
  width: image-width(sprite-file($images, "cat"));
  height: image-height(sprite-file($images, "cat"));
  background: sprite($images, "cat");
}
```

The output CSS
```
div {
  width: 140px;
  height: 79px;
  background: url("genimg/sprites-wehqi.png") 0px 0px;
}
```

#### Available commands

```bash
$ wt -h

wt is a Sass project tool made to handle large projects. It uses the libSass compiler for efficiency and speed.

Usage:
  wt [flags]
  wt [command]

Available Commands:
  serve       Starts a http server that will convert Sass to CSS
  compile     Compile Sass stylesheets to CSS
  watch       Watch Sass files for changes and rebuild CSS

Flags:
  -b, --build="": Path to target directory to place generated CSS, relative paths inside project directory are preserved
      --comment[=true]: Turn on source comments
  -c, --config="": Temporarily disabled: Location of the config file
      --cpuprofile="": Go runtime cpu profilling for debugging
      --css-dir="": Compass backwards compat, does nothing. Reference locations relative to Sass project directory
      --debug[=false]: Show detailed debug information
  -d, --dir="": Path to locate images for spriting and image functions
      --font=".": Path to directory containing fonts
      --gen=".": Path to place generated images
      --images-dir="": Compass backwards compat, use -d instead
      --javascripts-dir="": Compass backwards compat, ignored
      --no-line-comments[=false]: UNSUPPORTED: Disable line comments
  -p, --proj="": Path to directory containing Sass stylesheets
      --relative-assets[=false]: UNSUPPORTED: Make compass asset helpers generate relative urls to assets.
      --sass-dir="": Compass backwards compat, use -p instead
  -s, --style="nested": nested style of output CSS
                        available options: nested, expanded, compact, compressed
      --time[=false]: Retrieve timing information
  -v, --version[=false]: Show the app version

Use "wt [command] --help" for more information about a command.
```


#### Try before you buy

You can try out Wellington on Codepen, fork the [Wellington Playground](http://codepen.io/pen/def?fork=KwggLx)! This live example has images you can use, or you can bring your Sass.

There are many examples on Codepen just see the Wellington [collection](http://codepen.io/collection/DbNZQJ/)

#### Installation

Wellington can be installed via brew

	brew install wellington
	wt -h

Or, you can run wellington in a docker container

	docker run -v $(pwd):/data -it drewwells/wellington wt compile proj.scss

## Documentation

### Why?

Sass is a fantastic language. It adds a lot of power to standard CSS. If only our clients were happy with the functionality that Sass provided. For the life of Sass, there has been only one tool that attempted to extend Sass for everything that's needed to build a site. While Ruby is great for development, it does have some drawbacks. As our Sass powered website grew, Compass and Ruby Sass started to become a real drag on build times and development happiness. A typical build including transpiling Sass to CSS, RequireJS JavaScript, and minfication of CSS, JS, and images would spend half the time processing the Sass.

There had to be a better way. Libsass was starting to gain some traction, but it didn't do everything we needed. So I wrote Wellington to be a drop in replacement for the spriting functions familar to those used to Compass. This makes it super simple to swap out Compass with Wellington in your Sass projects.

### See how the sausage is made

#### Building from source
Install Go and add $GOPATH/bin to your $PATH. [Detailed instructions](https://golang.org/doc/install). Wellington requires Go 1.3.1+.

```
go get -u github.com/wellington/wellington
cd $GOPATH/src/github.com/wellington/wellington

make

# You should not have wt in your path
wt -h
```

It's a good idea to export `PKG_CONFIG_PATH` so that pkg-config can find `libsass.pc`. Otherwise, `go ...` commands will fail.

```
export PKG_CONFIG_PATH=$GOPATH/src/github.com/wellington/libsass/lib/pkgconfig
```

Set your fork as the origin.

    cd $GOPATH/src/github.com/wellington/wellington
	git remote rm origin
	git remote add origin git@github.com:username/wellington.git

Testing

    make test

Profiling

	make profile

Build a Docker Container. The wt container is 33.6 MB in size, but builds in a much larger container 844.7 MB.

	make build
	make docker #launch a container

Please use pull requests for contributing code.  [CircleCI](https://circleci.com/gh/wellington/wellington) will automatically test and lint your contributions.  Thanks for helping!

### Getting Help

Ask questions in the QA forum on [Google Group](https://groups.google.com/forum/#!forum/wellington-development)

### License

Wellington is licensed under MIT.


[![Bitdeli Badge](https://d2weczhvl823v0.cloudfront.net/wellington/wellington/trend.png)](https://bitdeli.com/free "Bitdeli Badge")
