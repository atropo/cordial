/*
Copyright © 2022 ITRS Group

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.

You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import "sync"

type expandOptions struct {
	defaultValue       any
	expandNonString    bool
	expandNonStringCSV bool
	expressions        bool
	externalFuncMaps   bool
	funcMaps           map[string]func(*Config, string, bool) (string, error)
	initialValue       any
	lookupTables       []map[string]string
	nodecode           bool
	rawstring          bool
	replacements       []string
	trimPrefix         bool
	trimSpace          bool
	usekeyfile         string
}

// ExpandOptions control the way configuration options undergo string
// expansion through the underlying [ExpandString] functions.
// ExpandOptions can be passed to any of the normal lookup functions
// that are provided to override [viper] versions, such as [GetString].
//
// e.g.
//
//	s := config.GetString("config.value", ExternalLookups(false), LookupTable(configMap), Prefix("myconf", myFunc))
type ExpandOptions func(*expandOptions)

// use a normal map, protected with mutex and not sync.Map as this is
// copied into a normal map later on
var defaultFuncMaps = map[string]func(*Config, string, bool) (string, error){
	"http":  fetchURL,
	"https": fetchURL,
	"file":  fetchFile,
}
var defaultFuncMapsMutex sync.Mutex

func evalExpandOptions(c *Config, options ...ExpandOptions) (e *expandOptions) {
	e = &expandOptions{
		externalFuncMaps: true,
		funcMaps:         map[string]func(*Config, string, bool) (string, error){},
		replacements:     []string{},
		trimSpace:        true,
	}

	for _, opt := range c.defaultExpandOptions {
		opt(e)
	}

	for _, opt := range options {
		opt(e)
	}

	if e.externalFuncMaps {
		defaultFuncMapsMutex.Lock()
		for k, v := range e.funcMaps {
			defaultFuncMaps[k] = v
		}
		e.funcMaps = defaultFuncMaps
		defaultFuncMapsMutex.Unlock()
	}

	if e.expressions {
		e.funcMaps["expr"] = expr
	}

	if e.defaultValue == nil {
		e.defaultValue = ""
	}

	return
}

// DefaultExpandOptions sets defaults to all subsequent calls to
// functions that perform configuration expansion. These defaults can be
// reset by calling DefaultExpandOptions with no arguments.
func (c *Config) DefaultExpandOptions(options ...ExpandOptions) {
	c.defaultExpandOptions = options
}

// NoExpand overrides all other options except Default and returns the
// value (or the default) as-is with no expansion applied. This is to
// allow the normal functions and methods to be called but to receive
// the underlying configuration item, such as an encoded password.
func NoExpand() ExpandOptions {
	return func(e *expandOptions) {
		e.rawstring = true
	}
}

// NoDecode disables the expansion of encoded values.
func NoDecode(n bool) ExpandOptions {
	return func(e *expandOptions) {
		e.nodecode = n
	}
}

// LookupTable adds lookup maps, of name/value pairs, to the Expand
// functions. If there are no lookup tables defined then `${item}` is
// treated as an environment variable. When string expansion is done to
// a plain word, ie. without a prefix, then `${item}` is looked up in
// each map, in the order the LookupTable options are given, and first
// match, if any, wins. If there is no match in any of the lookup maps
// then a nil value is returned and environment variables are not
// checked.
func LookupTable(values ...map[string]string) ExpandOptions {
	return func(e *expandOptions) {
		e.lookupTables = append(e.lookupTables, values...)
	}
}

// LookupTables sets the expansion lookup tables to the slice of maps
// passed as values. Any existing lookup tables are discarded.
//
// Deprecated: Use the singular LookupTable with a variadic list of
// tables instead.
func LookupTables(values []map[string]string) ExpandOptions {
	return func(e *expandOptions) {
		e.lookupTables = values
	}
}

// Prefix defines a custom mapping for the given prefix to an
// expand-like function. The prefix should not include the terminating
// ":". If the configuration prefix matches during expansion then the
// function is called with the config data and the contents of the
// expansion including the prefix (for URLs) but stripped of the opening
// `${` and closing `}`. A boolean parameter trims white space from the
// result if true.
func Prefix(prefix string, fn func(*Config, string, bool) (string, error)) ExpandOptions {
	return func(e *expandOptions) {
		e.funcMaps[prefix] = fn
	}
}

// ExternalLookups enables or disables the built-in expansion options
// that fetch data from outside the program, such as URLs and file
// paths. The default is true.
func ExternalLookups(yes bool) ExpandOptions {
	return func(e *expandOptions) {
		e.externalFuncMaps = yes
	}
}

// Expressions enables or disables the built-in expansion for
// expressions via the `github.com/maja42/goval` package. The default is
// false.
func Expressions(yes bool) ExpandOptions {
	return func(e *expandOptions) {
		e.expressions = yes
	}
}

// TrimPrefix enables the removal of the prefix from the string passed
// to expansion functions. If this is not set then URLs can be passed
// as-is since the prefix is part of the URL. If set then URLs would
// need the schema explicitly added after the prefix. Using this option
// allows standard function like [strings.ToUpper] to be used without
// additional wrappers.
func TrimPrefix() ExpandOptions {
	return func(e *expandOptions) {
		e.trimPrefix = true
	}
}

// TrimSpace enables the removal of leading and trailing spaces on all
// values in an expansion. The default is `true`. If a default
// value is given using the Default() then this is never trimmed.
func TrimSpace(yes bool) ExpandOptions {
	return func(e *expandOptions) {
		e.trimSpace = yes
	}
}

// Default sets a default value to be returned if the resulting
// expansion of the whole config value is empty (after any optional
// trimming of leading and trailing spaces). This includes cases where
// external lookups fail or a configuration item is not found. If the
// TrimSpace option is set to false and the expanded value consists
// wholly of whitespace then that is returned and not the default given
// here.
func Default(value any) ExpandOptions {
	return func(e *expandOptions) {
		e.defaultValue = value
	}
}

// Initial sets an initial default value to be used if the configuration
// item is empty (or nil) to start. This differs from Default() which
// supplies a value to use if the value if empty after expansion. The
// initial value, if used, is expanded as would any configuration value.
//
// If config.NoExpand() is also used then this initial value is used as a
// secondary default - i.e. if config.Default() is empty.
func Initial(value any) ExpandOptions {
	return func(e *expandOptions) {
		e.initialValue = value
	}
}

// Replace is used by config.Set* (except config.Set itself) functions
// to replace substrings with the formatted configuration item given as
// name with an equivalent expand string, where the value of the name
// key is only tested as Set time.
//
// e.g. if ${home} is "/home/user" then:
//
//	config.SetString("path", "/home/user/file.txt", config.Replace("home"))
//
// results in path being set to "${home}/file.txt" for future expansion,
// as "home" may change
//
// Replace can be used multiple times, each name being checked in order.
//
// Expand strings in the value are never substituted.
//
// name is not checked for self-referencing
func Replace(name string) ExpandOptions {
	return func(e *expandOptions) {
		e.replacements = append(e.replacements, name)
	}
}

// UseKeyfile overrides the path to the embedded keyfile in the
// `${enc:/path:xxx}` value. This can be useful when the keyfile is
// placed in an alternative location.
func UseKeyfile(file string) ExpandOptions {
	return func(eo *expandOptions) {
		eo.usekeyfile = file
	}
}

// ExpandNonStringToCSV causes any non-string configuration item
// expansions of lists to return a comma-separated list of any strings
// in that list. Non-string values are skipped.
func ExpandNonStringToCSV() ExpandOptions {
	return func(eo *expandOptions) {
		eo.expandNonString = true
		eo.expandNonStringCSV = true
	}
}

// ExpandNonStringToJSON causes any non-string configuration expansions
// to be returned as a JSON encoded string of the item
func ExpandNonStringToJSON() ExpandOptions {
	return func(eo *expandOptions) {
		eo.expandNonString = true
	}
}
