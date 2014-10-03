package sprite_sass

import (
	"fmt"
	"log"
	"path/filepath"
)

func (p Parser) ImageUrl(items []Item) string {
	gdir, err := filepath.Rel(p.BuildDir, p.ImageDir)
	if err != nil {
		log.Println(err)
		return "url(\"\")"
	}
	path := filepath.Join(gdir, items[2].Value)
	return fmt.Sprintf("url(\"%s\")", path)
}
