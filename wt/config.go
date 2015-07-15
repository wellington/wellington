package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
)

var kvReg = regexp.MustCompile(`^(\S+)\s?=\s?\"(\S+)\"`)

func ConfigParse(path string) map[string]string {
	m := make(map[string]string)
	// Parses and modifies flags to suit what was found
	bs, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println("Error reading file %s: %s", path, err)
		os.Exit(1)
	}

	// Split file by lines
	lines := bytes.Split(bs, []byte("\n"))
	for _, line := range lines {
		// Really simple parser, just a regex
		if kvReg.Match(line) {
			//fmt.Println(string(line))
			ss := kvReg.FindAllStringSubmatch(string(line), -1)
			// Some assumptions follow
			if len(ss[0]) == 3 {
				hit := ss[0]
				m[hit[1]] = hit[2]
			}
		}

	}

	return m
}
