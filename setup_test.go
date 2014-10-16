package sprite_sass

import (
	"log"
	"os"
	"path/filepath"

	. "github.com/drewwells/spritewell"
)

func cleanUpSprites(sprites map[string]ImageList) {
	if sprites == nil {
		return
	}
	for _, iml := range sprites {
		err := os.Remove(filepath.Join(iml.GenImgDir, iml.OutFile))
		if err != nil {
			log.Fatal(err)
		}
	}
}
