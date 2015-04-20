package context

// Mixins registers the default list of supported mixins
func Mixins(ctx *Context) {
	RegisterHeader(`@mixin sprite-dimensions($map, $name) {
  $file: sprite-file($map, $name);
  height: image-height($file);
  width: image-width($file);
}`)
}
