// Copyright © 2020 Cosku Bas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package convar

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
)

// Save saves all convars to the given config file. Only non-default values are saved.
func (c *Console) Save(filePath string) error {
	var buffer bytes.Buffer
	c.varLock.RLock()
	defer c.varLock.RUnlock()
	for _, cv := range c.variables {
		value := cv.value.Load()
		if value != cv.valDefault && !cv.isFunc {
			buffer.WriteString(fmt.Sprintf("%s %v\n", cv.varName, value))
		}
	}
	return ioutil.WriteFile(filePath, buffer.Bytes(), os.ModePerm)
}

// Load executes each line in the given config file.
// If the any convars in the file are not registered with RegVar before calling this method, they will be ignored.
func (c *Console) Load(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		c.exec(true, scanner.Text())
	}
	return nil
}
