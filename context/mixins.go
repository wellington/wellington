package context

func Mixins(ctx *Context) {
	RegisterHeader(`@mixin sprite-dimensions($map, $name) {
  $file: sprite-file($map, $name);
  height: image-height($file);
  width: image-width($file);
}`)
}
