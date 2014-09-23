package sprite_sass

import (
	"fmt"
	"os"
)

func (p Parser) ImageUrl(items []Item) string {
	path := p.ImageDir + "/" + items[2].Value
	if _, err := os.Stat(path); err == nil {
		return fmt.Sprintf("url(\"%s\")", path)
	}
	// TODO: Error scenario, find a way to surface these
	return "transparent"
}
