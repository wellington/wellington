package spritewell

import (
	"bytes"
	"io"
	"os"
	"testing"
)

func TestIsSVG(t *testing.T) {
	f, err := os.Open("test/gopher-front.svg")
	if err != nil {
		t.Error(err)
	}
	b := IsSVG(f)
	if e := true; e != b {
		t.Errorf("got: %t wanted: %t", b, e)
	}

	var buf bytes.Buffer
	f, err = os.Open("test/pixel.png")
	if err != nil {
		t.Error(err)
	}
	tr := io.TeeReader(f, &buf)
	b = IsSVG(tr)
	if e := false; e != b {
		t.Errorf("got: %t wanted: %t", b, e)
	}
	if e := 172; e != len(buf.Bytes()) {
		t.Errorf("got: %d wanted: %d", len(buf.Bytes()), e)
	}

	buf.Reset()
	f, err = os.Open("test/139.png")
	if err != nil {
		t.Error(err)
	}
	tr = io.TeeReader(f, &buf)
	b = IsSVG(tr)
	if e := false; e != b {
		t.Errorf("got: %t wanted: %t", b, e)
	}
	if bytes.MinRead != len(buf.Bytes()) {
		t.Errorf("got: %d wanted: %d", len(buf.Bytes()), bytes.MinRead)
	}

}

func TestInline(t *testing.T) {

	esc := `url("data:image/svg+xml;utf8,%3C%3Fxml%20version=%221.0%22%20encoding=%22utf-8%22%3F%3E%3C%21--%20Generator:%20Adobe%20Illustrator%2018.1.0,%20SVG%20Export%20Plug-In%20.%20SVG%20Version:%206.00%20Build%200%29%20%20--%3E%3Csvg%20version=%221.1%22%20id=%22Gopher%22%20xmlns=%22http://www.w3.org/2000/svg%22%20xmlns:xlink=%22http://www.w3.org/1999/xlink%22%20x=%220px%22%20y=%220px%22%09%20viewBox=%220%200%20215.6%20281.6%22%20enable-background=%22new%200%200%20215.6%20281.6%22%20xml:space=%22preserve%22%3E%3Cg%3E%09%3Cpath%20fill=%22%238CC5E7%22%20d=%22M207.3,44.6c-6.7-13.7-22.9-1.6-27-5.9c-21-21.6-46.4-27-66.3-28c0,0-9,0-11,0c-20,0.5-45.4,6.3-66.3,28%09%09c-4.1,4.3-20.4-7.8-27,5.9c-7.7,16,15.7,17.6,14.5,24.7c-2.3,12.8-0.8,31.8,1,50.5C28,151.5,4.3,227.4,53.6,257.9%09%09c9.3,5.8,34.4,9,56.2,9.5l0,0c0,0,0.1,0,0.1,0c0,0,0.1,0,0.1,0l0,0c21.8-0.5,43.9-3.7,53.2-9.5c49.4-30.5,25.7-106.4,28.6-138.1%09%09c1.7-18.7,3.2-37.7,1-50.5C191.6,62.2,215,60.5,207.3,44.6z%22/%3E%09%3Cg%3E%09%09%3Cpath%20fill=%22%23E0DEDC%22%20d=%22M143.2,54.3c-33.4,3.9-28.9,38.7-16,50c24,21,49,0,46.2-21.2C170.9,62.7,153.6,53.1,143.2,54.3z%22/%3E%09%09%3Ccircle%20fill=%22%23111212%22%20cx=%22145.5%22%20cy=%2284.3%22%20r=%2211.4%22/%3E%09%09%3Ccircle%20fill=%22%23FFFFFF%22%20cx=%22142.5%22%20cy=%2279.4%22%20r=%223.6%22/%3E%09%3C/g%3E%09%3Cg%3E%09%09%3Cpath%20fill=%22%23B8937F%22%20d=%22M108.5,107c-16,2.4-21.7,7-20.5,14.2c2,11.8,39.7,10.5,40.9,0.6C129.9,113.3,114.8,106.1,108.5,107z%22/%3E%09%09%3Cpath%20d=%22M98.2,111.8c-2.7,9.8,21.7,8.3,21.1,2c-0.3-3.7-3.6-8.4-12.3-8.2C103.6,105.7,99.4,107.2,98.2,111.8z%22/%3E%09%09%3Cpath%20fill=%22%23E0DEDC%22%20d=%22M99,127.7c-0.9,0.4-2.4,10.2,2.2,10.7c3.1,0.3,11.6,1.3,13.6,0c3.9-2.5,3.5-8.5,1.3-10%09%09%09C112.4,126,100,127.2,99,127.7z%22/%3E%09%3C/g%3E%09%3Cg%3E%09%09%3Cpath%20fill=%22%23E0DEDC%22%20d=%22M73.6,54.3c33.4,3.9,28.9,38.7,16,50c-24,21-49,0-46.2-21.2C46,62.7,63.3,53.1,73.6,54.3z%22/%3E%09%09%3Ccircle%20fill=%22%23111212%22%20cx=%2271.4%22%20cy=%2284.3%22%20r=%2211.4%22/%3E%09%09%3Ccircle%20fill=%22%23FFFFFF%22%20cx=%2274.4%22%20cy=%2279.4%22%20r=%223.6%22/%3E%09%3C/g%3E%09%3Cpath%20fill=%22%23B8937F%22%20d=%22M193.6,186.7c11,0.1,5.6-23.5-1.2-18.8c-3.3,2.3-3.9,7.6-3.9,12.1C188.5,182.5,190.5,186.6,193.6,186.7z%22/%3E%09%3Cpath%20fill=%22%23B8937F%22%20d=%22M23.3,186.7c-11,0.1-5.6-23.5,1.2-18.8c3.3,2.3,3.9,7.6,3.9,12.1C28.4,182.5,26.4,186.6,23.3,186.7z%22/%3E%09%3Cpath%20fill=%22%23B8937F%22%20d=%22M172.7,259.2c-6-8.9-11.4-2-20.1,2.4c-4.1,2.1,6.8,9.6,19,4C174.8,264.1,174.7,262.1,172.7,259.2z%22/%3E%09%3Cpath%20fill=%22%23B8937F%22%20d=%22M44.2,260.2c6-8.9,11.4-2,20.1,2.4c4.1,2.1-6.8,9.6-19,4C42.1,265.1,42.2,263.1,44.2,260.2z%22/%3E%09%3Cpath%20fill=%22%233C89BF%22%20d=%22M188.6,47c-0.6,2.1,2.1,1.8,3.1,8.3c0.4,2.4,9-3.5,5.5-7.8C194.3,43.9,189.1,44.9,188.6,47z%22/%3E%09%3Cpath%20fill=%22%233C89BF%22%20d=%22M28.3,47c0.6,2.1-2.1,1.8-3.1,8.3c-0.4,2.4-9-3.5-5.5-7.8C22.5,43.9,27.7,44.9,28.3,47z%22/%3E%3C/g%3E%3C/svg%3E")`

	b64 := `url("data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBlbmNvZGluZz0idXRmLTgiPz4NCjwhLS0gR2VuZXJhdG9yOiBBZG9iZSBJbGx1c3RyYXRvciAxOC4xLjAsIFNWRyBFeHBvcnQgUGx1Zy1JbiAuIFNWRyBWZXJzaW9uOiA2LjAwIEJ1aWxkIDApICAtLT4NCjxzdmcgdmVyc2lvbj0iMS4xIiBpZD0iR29waGVyIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB4PSIwcHgiIHk9IjBweCINCgkgdmlld0JveD0iMCAwIDIxNS42IDI4MS42IiBlbmFibGUtYmFja2dyb3VuZD0ibmV3IDAgMCAyMTUuNiAyODEuNiIgeG1sOnNwYWNlPSJwcmVzZXJ2ZSI+DQo8Zz4NCgk8cGF0aCBmaWxsPSIjOENDNUU3IiBkPSJNMjA3LjMsNDQuNmMtNi43LTEzLjctMjIuOS0xLjYtMjctNS45Yy0yMS0yMS42LTQ2LjQtMjctNjYuMy0yOGMwLDAtOSwwLTExLDBjLTIwLDAuNS00NS40LDYuMy02Ni4zLDI4DQoJCWMtNC4xLDQuMy0yMC40LTcuOC0yNyw1LjljLTcuNywxNiwxNS43LDE3LjYsMTQuNSwyNC43Yy0yLjMsMTIuOC0wLjgsMzEuOCwxLDUwLjVDMjgsMTUxLjUsNC4zLDIyNy40LDUzLjYsMjU3LjkNCgkJYzkuMyw1LjgsMzQuNCw5LDU2LjIsOS41bDAsMGMwLDAsMC4xLDAsMC4xLDBjMCwwLDAuMSwwLDAuMSwwbDAsMGMyMS44LTAuNSw0My45LTMuNyw1My4yLTkuNWM0OS40LTMwLjUsMjUuNy0xMDYuNCwyOC42LTEzOC4xDQoJCWMxLjctMTguNywzLjItMzcuNywxLTUwLjVDMTkxLjYsNjIuMiwyMTUsNjAuNSwyMDcuMyw0NC42eiIvPg0KCTxnPg0KCQk8cGF0aCBmaWxsPSIjRTBERURDIiBkPSJNMTQzLjIsNTQuM2MtMzMuNCwzLjktMjguOSwzOC43LTE2LDUwYzI0LDIxLDQ5LDAsNDYuMi0yMS4yQzE3MC45LDYyLjcsMTUzLjYsNTMuMSwxNDMuMiw1NC4zeiIvPg0KCQk8Y2lyY2xlIGZpbGw9IiMxMTEyMTIiIGN4PSIxNDUuNSIgY3k9Ijg0LjMiIHI9IjExLjQiLz4NCgkJPGNpcmNsZSBmaWxsPSIjRkZGRkZGIiBjeD0iMTQyLjUiIGN5PSI3OS40IiByPSIzLjYiLz4NCgk8L2c+DQoJPGc+DQoJCTxwYXRoIGZpbGw9IiNCODkzN0YiIGQ9Ik0xMDguNSwxMDdjLTE2LDIuNC0yMS43LDctMjAuNSwxNC4yYzIsMTEuOCwzOS43LDEwLjUsNDAuOSwwLjZDMTI5LjksMTEzLjMsMTE0LjgsMTA2LjEsMTA4LjUsMTA3eiIvPg0KCQk8cGF0aCBkPSJNOTguMiwxMTEuOGMtMi43LDkuOCwyMS43LDguMywyMS4xLDJjLTAuMy0zLjctMy42LTguNC0xMi4zLTguMkMxMDMuNiwxMDUuNyw5OS40LDEwNy4yLDk4LjIsMTExLjh6Ii8+DQoJCTxwYXRoIGZpbGw9IiNFMERFREMiIGQ9Ik05OSwxMjcuN2MtMC45LDAuNC0yLjQsMTAuMiwyLjIsMTAuN2MzLjEsMC4zLDExLjYsMS4zLDEzLjYsMGMzLjktMi41LDMuNS04LjUsMS4zLTEwDQoJCQlDMTEyLjQsMTI2LDEwMCwxMjcuMiw5OSwxMjcuN3oiLz4NCgk8L2c+DQoJPGc+DQoJCTxwYXRoIGZpbGw9IiNFMERFREMiIGQ9Ik03My42LDU0LjNjMzMuNCwzLjksMjguOSwzOC43LDE2LDUwYy0yNCwyMS00OSwwLTQ2LjItMjEuMkM0Niw2Mi43LDYzLjMsNTMuMSw3My42LDU0LjN6Ii8+DQoJCTxjaXJjbGUgZmlsbD0iIzExMTIxMiIgY3g9IjcxLjQiIGN5PSI4NC4zIiByPSIxMS40Ii8+DQoJCTxjaXJjbGUgZmlsbD0iI0ZGRkZGRiIgY3g9Ijc0LjQiIGN5PSI3OS40IiByPSIzLjYiLz4NCgk8L2c+DQoJPHBhdGggZmlsbD0iI0I4OTM3RiIgZD0iTTE5My42LDE4Ni43YzExLDAuMSw1LjYtMjMuNS0xLjItMTguOGMtMy4zLDIuMy0zLjksNy42LTMuOSwxMi4xQzE4OC41LDE4Mi41LDE5MC41LDE4Ni42LDE5My42LDE4Ni43eiIvPg0KCTxwYXRoIGZpbGw9IiNCODkzN0YiIGQ9Ik0yMy4zLDE4Ni43Yy0xMSwwLjEtNS42LTIzLjUsMS4yLTE4LjhjMy4zLDIuMywzLjksNy42LDMuOSwxMi4xQzI4LjQsMTgyLjUsMjYuNCwxODYuNiwyMy4zLDE4Ni43eiIvPg0KCTxwYXRoIGZpbGw9IiNCODkzN0YiIGQ9Ik0xNzIuNywyNTkuMmMtNi04LjktMTEuNC0yLTIwLjEsMi40Yy00LjEsMi4xLDYuOCw5LjYsMTksNEMxNzQuOCwyNjQuMSwxNzQuNywyNjIuMSwxNzIuNywyNTkuMnoiLz4NCgk8cGF0aCBmaWxsPSIjQjg5MzdGIiBkPSJNNDQuMiwyNjAuMmM2LTguOSwxMS40LTIsMjAuMSwyLjRjNC4xLDIuMS02LjgsOS42LTE5LDRDNDIuMSwyNjUuMSw0Mi4yLDI2My4xLDQ0LjIsMjYwLjJ6Ii8+DQoJPHBhdGggZmlsbD0iIzNDODlCRiIgZD0iTTE4OC42LDQ3Yy0wLjYsMi4xLDIuMSwxLjgsMy4xLDguM2MwLjQsMi40LDktMy41LDUuNS03LjhDMTk0LjMsNDMuOSwxODkuMSw0NC45LDE4OC42LDQ3eiIvPg0KCTxwYXRoIGZpbGw9IiMzQzg5QkYiIGQ9Ik0yOC4zLDQ3YzAuNiwyLjEtMi4xLDEuOC0zLjEsOC4zYy0wLjQsMi40LTktMy41LTUuNS03LjhDMjIuNSw0My45LDI3LjcsNDQuOSwyOC4zLDQ3eiIvPg0KPC9nPg0KPC9zdmc+")`

	f, err := os.Open("test/gopher-front.svg")
	if err != nil {
		t.Error(err)
	}
	var fb, fbb, buf bytes.Buffer

	tr := io.TeeReader(f, &fb)
	Inline(tr, &buf)
	if esc != buf.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", buf.String(), esc)
	}

	tr = io.TeeReader(&fb, &fbb)
	buf.Reset()
	Inline(tr, &buf, false)
	if esc != buf.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", buf.String(), esc)
	}

	buf.Reset()
	Inline(&fbb, &buf, true)
	if b64 != buf.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", buf.String(), b64)
	}

}

func TestBinaryInline(t *testing.T) {

	f, err := os.Open("test/pixel.png")
	if err != nil {
		t.Error(err)
	}
	var buf bytes.Buffer
	Inline(f, &buf)

	e := `url("data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAEElEQVR4nGJiAAJAAAAA//8ADwADkGsgEwAAAABJRU5ErkJg")`

	if e != buf.String() {
		t.Errorf("got:\n%s\nwanted:\n%s", buf.String(), string(e))
	}
}
