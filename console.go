// Copyright © 2020 Cosku Bas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package convar

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
)

// Console is a Quake-like console implementation for games.
type Console struct {
	variables     map[string]*ConVar
	varLock       sync.RWMutex
	buffer        []string
	bufLock       sync.Mutex
	bufMaxLines   int
	logLevel      LogLevel
	logInfoPrefix string
	logWarnPrefix string
	logErrPrefix  string
}

// NewConsole creates a new console instance with the given settings.
// If buffer size reaches bufMaxLines, old lines will be discarded.
// Only logs that are of smaller level than logLevel will be written to the buffer.
func NewConsole(bufMaxLines int, logLevel LogLevel, logInfoPrefix string, logWarnPrefix string, logErrPrefix string) *Console {
	c := &Console{
		variables:     make(map[string]*ConVar),
		bufMaxLines:   bufMaxLines,
		logLevel:      logLevel,
		logInfoPrefix: logInfoPrefix,
		logWarnPrefix: logWarnPrefix,
		logErrPrefix:  logErrPrefix,
	}
	return c
}

// RegDefaultConVars registers an assortment of useful convars.
//		con_dump:		Saves the console buffer to a file.
//		con_clear:		Clears the console buffer.
//		var_reset:		Resets given convar to its default value without triggering its callback.
//		var_reset_all:	Resets all convars to their default values without triggering their callbacks.
//		var_load:		Loads convars from a file, overwriting the ones that are already in the memory.
//		var_save:		Saves convars to a file.
//		var_list:		Lists all convars with their description.
func (c *Console) RegDefaultConVars() {
	c.RegConVar(
		NewConVar("con_dump", reflect.String, true, "Saves the console buffer to a file.", "console.log", func(con *Console, oldVal, newVal interface{}) {
			file := newVal.(string)
			if file == "" {
				file = oldVal.(string)
			}
			if err := con.DumpBuffer(file); err != nil {
				con.LogErrorf("%v", err)
				return
			}
			con.LogInfof("%s is saved", file)
		}),
	)
	c.RegConVar(
		NewConVar("con_clear", reflect.Int, true, "Clears the console buffer.", 0, func(con *Console, oldVal, newVal interface{}) {
			con.ClearBuffer()
		}),
	)
	c.RegConVar(
		NewConVar("var_reset_all", reflect.Int, true, "Resets all convars to their default values.", 0, func(con *Console, oldVal, newVal interface{}) {
			con.ResetAllVar()
		}),
	)
	c.RegConVar(
		NewConVar("var_reset", reflect.String, true, "Resets given convar to its default value.", "", func(con *Console, oldVal, newVal interface{}) {
			if newVal == nil {
				con.LogErrorf(errNilValue)
				return
			}
			cv := con.ConVar(newVal.(string))
			if cv == nil {
				con.LogErrorf(errVarNotFound, newVal.(string))
				return
			}
			cv.Reset()
			con.LogInfof("%s is reset", newVal.(string))
		}),
	)
	c.RegConVar(
		NewConVar("var_load", reflect.String, true, "Loads convars from a file, overwriting the ones that are already in the memory.", "convars.ini", func(con *Console, oldVal, newVal interface{}) {
			file := newVal.(string)
			if file == "" {
				file = oldVal.(string)
			}
			if err := con.Load(file); err != nil {
				con.LogErrorf("%v", err)
				return
			}
			con.LogInfof("%s is loaded", file)
		}),
	)
	c.RegConVar(
		NewConVar("var_save", reflect.String, true, "Saves convars to a file.", "convars.ini", func(con *Console, oldVal, newVal interface{}) {
			file := newVal.(string)
			if file == "" {
				file = oldVal.(string)
			}
			if err := con.Save(file); err != nil {
				con.LogErrorf("%v", err)
				return
			}
			con.LogInfof("%s is saved", file)
		}),
	)
	c.RegConVar(
		NewConVar("var_list", reflect.Int, true, "Lists all convars with their description.", 0, func(con *Console, oldVal, newVal interface{}) {
			cvs := con.ConVars()
			for _, cv := range cvs {
				con.LogInfof("%s: %s", cv.varName, cv.varDesc)
			}
		}),
	)
}

// RegConVar registers a new convar to be used in the console.
func (c *Console) RegConVar(cv *ConVar) {
	c.varLock.Lock()
	defer c.varLock.Unlock()
	cv.console = c
	c.variables[cv.varName] = cv
}

// ExecCmd parses and executes a console command string.
func (c *Console) ExecCmd(cmd string) (*ConVar, error) {
	return c.exec(false, cmd)
}

// ResetAllVar resets all convars to their default values.
// It doesn't trigger the set/update callback.
func (c *Console) ResetAllVar() {
	c.varLock.RLock()
	defer c.varLock.RUnlock()
	for _, cv := range c.variables {
		cv.value.Store(cv.valDefault)
	}
}

// ConVar returns the convar with the given name. Returns nil if it doesn't exist.
func (c *Console) ConVar(varName string) *ConVar {
	c.varLock.RLock()
	defer c.varLock.RUnlock()
	cv, ok := c.variables[strings.ToLower(varName)]
	if !ok {
		return nil
	}
	return cv
}

// ConVars returns a slice of all registered convars.
func (c *Console) ConVars() []*ConVar {
	c.varLock.RLock()
	defer c.varLock.RUnlock()
	var cvs []*ConVar
	for _, cv := range c.variables {
		cvs = append(cvs, cv)
	}
	return cvs
}

// Suggest suggests a list of size n, populated with the convars that have the substring str in their names.
func (c *Console) Suggest(str string, n int) []*ConVar {
	var (
		allCvs = c.ConVars()
		cvs    []*ConVar
	)
	if len([]rune(str)) < 3 {
		// 3 feels like a good minimum number of runes to trigger a suggestion feature
		// Open up an issue if you feel like it's not the best
		return cvs
	}
	for _, cv := range allCvs {
		if strings.Contains(cv.varName, strings.ToLower(str)) {
			cvs = append(cvs, cv)
			if len(cvs) >= n {
				return cvs
			}
		}
	}
	return cvs
}

func (c *Console) exec(fromFile bool, cmd string) (*ConVar, error) {
	cmd = strings.TrimSpace(strings.ToLower(cmd))
	tokens := strings.Fields(cmd)
	lent := len(tokens)
	if lent == 0 || tokens[0] == "#" {
		// Empty command or comment line
		return nil, nil
	}

	c.varLock.RLock()
	cv, ok := c.variables[tokens[0]]
	c.varLock.RUnlock()
	if !ok {
		return nil, fmt.Errorf(errVarNotFound, tokens[0])
	}

	// If the command is executed from a file and it's a func then ignore it
	if fromFile && cv.isFunc {
		return nil, nil
	}

	var (
		err    error
		value  interface{}
		valStr string
	)

	if lent == 1 {
		valStr = "0"
	} else {
		valStr = strings.TrimSpace(cmd[len(tokens[0]):])
	}

	switch cv.varType {
	case reflect.Bool, reflect.Int:
		value, err = strconv.Atoi(valStr)
	case reflect.Float64:
		value, err = strconv.ParseFloat(valStr, 64)
	case reflect.String:
		// Everything after the convar is considered part of the string
		// This also evaluates to an empty string in case of no value is put
		if lent == 1 {
			valStr = ""
		}
		value = strings.TrimSpace(cmd[len(tokens[0]):])
	}
	if err != nil {
		return nil, fmt.Errorf(errBadStringConversion, valStr, cv.varType)
	}

	// ex: write will apply below rules
	// cl_reload		(func)	run function with new value 'default', don't set any value
	// cl_reload	10	(func)	run function with new value 10, don't set any value
	// cl_width			(var)	don't run function, don't set any value
	// cl_width		10	(var)	run function with new value 10, set value to 10
	err = cv.write(cv.varType, value, lent)
	if err != nil {
		return nil, err
	}
	return cv, nil
}
