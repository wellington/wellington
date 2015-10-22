package libsass

// Mixins registers the default list of supported mixins
func init() {
	RegisterHeader(`@mixin sprite-dimensions($map, $name) {
  $file: sprite-file($map, $name);
  height: image-height($file);
  width: image-width($file);
}`)
}
