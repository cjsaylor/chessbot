// Copyright 2016 Florian Pigorsch. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"

	"github.com/flopp/go-findfont"
)

func main() {
	flag.Parse()
	for _, font := range flag.Args() {
		if filePath, err := findfont.Find(font); err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("Found %s at %s\n", font, filePath)
		}
	}
}
