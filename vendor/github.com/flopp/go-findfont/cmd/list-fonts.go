// Copyright 2016 Florian Pigorsch. All rights reserved.
//
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package main

import (
	"fmt"

	"github.com/flopp/go-findfont"
)

func main() {
	for _, path := range findfont.List() {
		fmt.Println(path)
	}
}
