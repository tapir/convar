// Copyright © 2020 Cosku Bas. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE.md file.

package convar

import (
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
)

// ConVar represents a console variable.
type ConVar struct {
	console    *Console
	varName    string
	varType    reflect.Kind
	varDesc    string
	value      atomic.Value
	valDefault interface{}
	valSet     ValSetFunc
	isFunc     bool
}

// NewConVar returns a convar of the given name and type. Convar names are case insensitive.
// varDefault is the default value.
// varDesc is the description of the convar.
// valSet is a callback function that is triggered everytime the convar's value is changed.
//
// NewConVar will panic if there are any errors.
// NewConVar should ideally be called for each convar at the begging of the application and before loading a config file.
// A convar cannot be safely used if it's not registered to a console instance via RegVar.
//
// When isFunc is true, a convar is treated in a special way:
// 		Convar is not saved to or loaded from the config file. This can be used to protect users from doing things like cyclic loading.
// 		SetInt, SetBool, SetFloat64, SetString functions do not change the value but instead trigger the callback with the given value.
// 		Value is always equal to default value.
func NewConVar(varName string, varType reflect.Kind, isFunc bool, varDesc string, valDefault interface{}, valSet ValSetFunc) *ConVar {
	varName = strings.ToLower(varName)
	if varType != reflect.TypeOf(valDefault).Kind() {
		// Type of valDefault and the given varType don't match
		// We panic here because ideally RegVar should be called once at the beggining
		panic(fmt.Errorf(errTypeMismatch, valDefault, varName, varType))
	}
	if !(varType == reflect.Int || varType == reflect.Float64 || varType == reflect.String) {
		panic(fmt.Errorf(errUnsupportedType, varType))
	}
	cv := &ConVar{
		varName:    varName,
		varType:    varType,
		varDesc:    varDesc,
		valDefault: valDefault,
		valSet:     valSet,
		isFunc:     isFunc,
	}
	cv.value.Store(valDefault)
	return cv
}

// ValSetFunc is the function signature of the value set/update callback.
type ValSetFunc func(con *Console, oldVal, newVal interface{})

func (cv *ConVar) write(varType reflect.Kind, value interface{}, argc int) error {
	if value == nil {
		return fmt.Errorf(errNilValue)
	}

	if varType != reflect.TypeOf(value).Kind() {
		// Type of value and given varType don't match
		return fmt.Errorf(errTypeMismatch, value, cv.varName, varType)
	}

	if cv.varType != varType {
		// Type of the found convar doesn't match with the given varType
		return fmt.Errorf(errVarBadType, cv.varName, varType)
	}

	if cv.isFunc {
		cv.valSet(cv.console, cv.valDefault, value)
		return nil
	}

	// If no argument was given and convar is not a function, we don't set the value
	if argc < 2 {
		return nil
	}

	oldVal := cv.value.Load()
	if oldVal == value {
		// Silently stop if the old and new values are the same
		return nil
	}
	cv.value.Store(value)
	cv.valSet(cv.console, oldVal, value)
	return nil
}

// Bool returns the value of the convar as a boolean. The underlying type for a boolean is integer.
func (cv *ConVar) Bool() (bool, error) {
	value := cv.value.Load()
	if reflect.TypeOf(value).Kind() != reflect.Int {
		return false, fmt.Errorf(errVarBadType, cv.varName, reflect.Int)
	}
	return value.(int) == 1, nil
}

// Int returns the value of the convar as an integer.
func (cv *ConVar) Int() (int, error) {
	value := cv.value.Load()
	if reflect.TypeOf(value).Kind() != reflect.Int {
		return 0, fmt.Errorf(errVarBadType, cv.varName, reflect.Int)
	}
	return value.(int), nil
}

// Float64 returns the value of the convar as a float64.
func (cv *ConVar) Float64() (float64, error) {
	value := cv.value.Load()
	if reflect.TypeOf(value).Kind() != reflect.Float64 {
		return 0, fmt.Errorf(errVarBadType, cv.varName, reflect.Float64)
	}
	return value.(float64), nil
}

// String returns the value of the convar as a string.
func (cv *ConVar) String() (string, error) {
	value := cv.value.Load()
	if reflect.TypeOf(value).Kind() != reflect.String {
		return "", fmt.Errorf(errVarBadType, cv.varName, reflect.String)
	}
	return value.(string), nil
}

// Interface returns the value of the convar as an interface which is the underlying data type for all convars.
// Interface will never return an error.
func (cv *ConVar) Interface() (interface{}, error) {
	return cv.value.Load(), nil
}

// SetBool sets the value of an integer convar from a boolean. true means 1 and false means 0.
func (cv *ConVar) SetBool(value bool) error {
	if value {
		return cv.write(reflect.Int, 1, 2)
	}
	return cv.write(reflect.Int, 0, 2)
}

// SetInt sets the convar to the given int value.
func (cv *ConVar) SetInt(value int) error {
	return cv.write(reflect.Int, value, 2)
}

// SetFloat64 sets the convar to the given float64 value.
func (cv *ConVar) SetFloat64(value float64) error {
	return cv.write(reflect.Float64, value, 2)
}

// SetString sets the convar to the given string value.
func (cv *ConVar) SetString(value string) error {
	return cv.write(reflect.String, value, 2)
}

// Name returns the name of the convar.
func (cv *ConVar) Name() string {
	return cv.varName
}

// Desc returns the description of the convar.
func (cv *ConVar) Desc() string {
	return cv.varDesc
}

// Type returns the type of the convar.
func (cv *ConVar) Type() reflect.Kind {
	return cv.varType
}

// Reset resets a convar to its default value.
// Value set/update callback function is not triggered.
func (cv *ConVar) Reset() {
	cv.value.Store(cv.valDefault)
}

// IsFunc returns true if the convar is set as a function.
func (cv *ConVar) IsFunc() bool {
	return cv.isFunc
}
