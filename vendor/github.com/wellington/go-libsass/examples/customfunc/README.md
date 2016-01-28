Cryptographically secure pseudorandom number generator in Sass. Well that is easy to do with this custom `crypto()` handler!

Start by registering a Sass function with the name `crypto()`. Now when `crypto()` is found in Sass, the cryptotext Go function will be called.

Input

``` sass
div { text: crypto(); }

```


Output

``` css
div {
  text: 'c91db27d5e580ef4292e'; }
```

Sass function written in Go

``` go
func cryptotext(ctx context.Context, usv libsass.SassValue) (*libsass.SassValue, error) {

	c := 10
	b := make([]byte, c)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	res, err := libsass.Marshal(fmt.Sprintf("'%x'", b))
	return &res, err
}
```
