// Copyright © 2020 Cosku Bas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package convar

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync/atomic"
)

// LogLevel is the type for the log level.
type LogLevel int32

const (
	// LogNone means no message will be printed to the console buffer.
	LogNone LogLevel = iota
	// LogInfo means only information messages will be printed to the console buffer.
	LogInfo
	// LogWarning means information and warning messages will be printed to the console buffer.
	LogWarning
	// LogError means information, warning and error messages will be printed to the console buffer.
	LogError
)

func (c *Console) log(prefix, format string, a ...interface{}) {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	if len(c.buffer) >= c.bufMaxLines {
		c.buffer = c.buffer[1:]
	}
	out := prefix + fmt.Sprintf(format, a...)
	c.buffer = append(c.buffer, out)
}

// LogInfof prints an information message to the console.
func (c *Console) LogInfof(format string, a ...interface{}) {
	if (int32)(LogInfo) > atomic.LoadInt32((*int32)(&c.logLevel)) {
		return
	}
	c.log(c.logInfoPrefix, format, a...)
}

// LogWarningf prints a warning message to the console.
func (c *Console) LogWarningf(format string, a ...interface{}) {
	if (int32)(LogInfo) > atomic.LoadInt32((*int32)(&c.logLevel)) {
		return
	}
	c.log(c.logWarnPrefix, format, a...)
}

// LogErrorf prints an error message to the console.
func (c *Console) LogErrorf(format string, a ...interface{}) {
	if (int32)(LogInfo) > atomic.LoadInt32((*int32)(&c.logLevel)) {
		return
	}
	c.log(c.logErrPrefix, format, a...)
}

// LogPrintf prints a message to the console without a prefix, regardless of the log level.
func (c *Console) LogPrintf(format string, a ...interface{}) {
	c.log("", format, a...)
}

// SetLogLevel changes the log level that will be written to the console buffer.
func (c *Console) SetLogLevel(level LogLevel) {
	atomic.StoreInt32((*int32)(&c.logLevel), (int32)(level))
}

// Buffer returns the console buffer as a string.
func (c *Console) Buffer() string {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	return strings.Join(c.buffer, "\n")
}

// BufferRaw returns the copy of the underlying raw buffer slice. Each element represents a line.
func (c *Console) BufferRaw() []string {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	ret := make([]string, len(c.buffer))
	copy(ret, c.buffer)
	return ret
}

// BufferWrapped returns the console buffer with each line wrapped to a new line.
// maxWidth is the maximum number of runes allowed before wrapping it to a new line.
//
// Note that this method doesn't take into account the rare cases of where a single character might
// be represented by multiple runes (ex: 'é́́'). Use https://pkg.go.dev/golang.org/x/text/unicode/norm together
// with Buffer() to handle such cases.
func (c *Console) BufferWrapped(maxWidth int) string {
	return strings.Join(c.BufferWrappedRaw(maxWidth), "\n")
}

// BufferWrappedRaw returns the console buffer with each line wrapped to a new line as a slice.
// maxWidth is the maximum number of runes allowed before wrapping it to a new line.
func (c *Console) BufferWrappedRaw(maxWidth int) []string {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	var ret []string
	for _, line := range c.buffer {
		if len([]rune(line)) > maxWidth {
			ret = append(ret, chunks(line, maxWidth)...)
		} else {
			ret = append(ret, line)
		}
	}
	return ret
}

// ClearBuffer clears the console buffer.
func (c *Console) ClearBuffer() {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	c.buffer = c.buffer[:0]
}

// DumpBuffer saves the console buffer to the given file.
func (c *Console) DumpBuffer(filePath string) error {
	c.bufLock.Lock()
	defer c.bufLock.Unlock()
	return ioutil.WriteFile(filePath, []byte(strings.Join(c.buffer, "\n")), os.ModePerm)
}

// Thanks to https://stackoverflow.com/questions/25686109/split-string-by-length-in-golang
func chunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}
	var chunks []string
	chunk := make([]rune, chunkSize)
	len := 0
	for _, r := range s {
		chunk[len] = r
		len++
		if len == chunkSize {
			chunks = append(chunks, string(chunk))
			len = 0
		}
	}
	if len > 0 {
		chunks = append(chunks, string(chunk[:len]))
	}
	return chunks
}
