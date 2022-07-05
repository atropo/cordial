/*
Copyright © 2022 ITRS Group

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

/*
Package to handle Geneos Gateway specific XPaths

These are a subset of the W3C XPath 1.0 standard and this package is
only interested in absolute paths and no matching is done. Geneos XPaths
are of a particular hierarchy and the ones we are interesting in here
are those used to communicate with the Gateway REST command API.

The two types of path handled are for headline or table cells, which
have the form:

/geneos/gateway/directory/probe/managedEntity/sampler/dataview/ ...
	... headlines/cell or ... rows/row/cell

Each component except "geneos", "directory", "headlines" and "rows" can
have a name and other predicates. The path can terminate at any level
that can carry a name. Apart from names, only attributes in
managedEntities are currently handled in anyway.
*/
package xpath

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

var ErrInvalidPath = errors.New("invalid Geneos XPath")

// A Geneos Gateway XPath
//
// Each field is a pointer, which if nil means the Xpath terminates at that point
// The "rows" boolean indicates in lower level components are headlines or rows
type XPath struct {
	Gateway  *Gateway  `json:"gateway,omitempty"`
	Probe    *Probe    `json:"probe,omitempty"`
	Entity   *Entity   `json:"entity,omitempty"`
	Sampler  *Sampler  `json:"sampler,omitempty"`
	Dataview *Dataview `json:"dataview,omitempty"`
	Headline *Headline `json:"headline,omitempty"`
	Rows     bool      `json:"-"`
	Row      *Row      `json:"row,omitempty"`
	Column   *Column   `json:"column,omitempty"`
}

type Gateway struct {
	Name string `json:"name,omitempty"`
}

type Probe struct {
	Name string `json:"name,omitempty"`
}

type Entity struct {
	Name       string            `json:"name,omitempty"`
	Attributes map[string]string `json:"attributes,omitempty"`
}

type Sampler struct {
	Name string  `json:"name,omitempty"`
	Type *string `json:"type,omitempty"`
}

type Dataview struct {
	Name string `json:"name,omitempty"`
}

type Headline struct {
	Name string `json:"name,omitempty"`
}

type Row struct {
	Name string `json:"name,omitempty"`
}

type Column struct {
	Name string `json:"name,omitempty"`
}

// return an XPath to the level of the element passed,
// which can be populated with fields.
func NewXPathTo(element interface{}) *XPath {
	x := &XPath{}
	return x.PathTo(element)
}

// return an xpath populated to the dataview, with name dv
// if no name is passed, create a wildcard dataview path
func NewDataviewPath(name string) (x *XPath) {
	x = NewXPathTo(&Dataview{Name: name})
	return
}

// return an xpath populated to the table cell identifies by row and column
func NewTableCellPath(row, column string) (x *XPath) {
	x = NewXPathTo(&Column{Name: column})
	x.Rows = true
	x.Row = &Row{Name: row}
	return
}

// return an xpath populated to the headline cell, identified by headline
func NewHeadlinePath(name string) (x *XPath) {
	x = NewXPathTo(&Headline{Name: name})
	return
}

// given an element type, return a new XPath to that
// element, removing lower level elements or adding
// empty elements to the level required. If the XPath
// does not contain an element of the type given then
// use the argument (which can include populated fields),
// but if empty then any existing element will be left
// as-is and not cleaned.
//
// e.g.
//    x := x.PathTo(&Dataview{})
//    y := xpath.PathTo(&Headline{Name: "headlineName"})
//
func (x *XPath) PathTo(element interface{}) *XPath {
	// copy the xpath
	var nx XPath

	// skip through any pointers
	for reflect.ValueOf(element).Kind() == reflect.Ptr {
		element = reflect.Indirect(reflect.ValueOf(element)).Interface()
	}

	// set the element, remove others
	switch e := element.(type) {
	case Gateway:
		nx = XPath{
			Gateway: &e,
		}
	case Probe:
		nx = XPath{
			Gateway: x.Gateway,
			Probe:   &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
	case Entity:
		nx = XPath{
			Gateway: x.Gateway,
			Probe:   x.Probe,
			Entity:  &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
		if nx.Probe == nil {
			nx.Probe = &Probe{}
		}
	case Sampler:
		nx = XPath{
			Gateway: nx.Gateway,
			Probe:   nx.Probe,
			Entity:  nx.Entity,
			Sampler: &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
		if nx.Probe == nil {
			nx.Probe = &Probe{}
		}
		if nx.Entity == nil {
			nx.Entity = &Entity{
				Attributes: make(map[string]string),
			}
		}
	case Dataview:
		nx = XPath{
			Gateway:  nx.Gateway,
			Probe:    nx.Probe,
			Entity:   nx.Entity,
			Sampler:  nx.Sampler,
			Dataview: &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
		if nx.Probe == nil {
			nx.Probe = &Probe{}
		}
		if nx.Entity == nil {
			nx.Entity = &Entity{
				Attributes: make(map[string]string),
			}
		}
		if nx.Sampler == nil {
			nx.Sampler = &Sampler{}
		}
	case Headline:
		nx = XPath{
			Gateway:  nx.Gateway,
			Probe:    nx.Probe,
			Entity:   nx.Entity,
			Sampler:  nx.Sampler,
			Dataview: nx.Dataview,
			Rows:     false,
			Headline: &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
		if nx.Probe == nil {
			nx.Probe = &Probe{}
		}
		if nx.Entity == nil {
			nx.Entity = &Entity{
				Attributes: make(map[string]string),
			}
		}
		if nx.Sampler == nil {
			nx.Sampler = &Sampler{}
		}
		if nx.Dataview == nil {
			nx.Dataview = &Dataview{}
		}
	case Row:
		nx = XPath{
			Gateway:  nx.Gateway,
			Probe:    nx.Probe,
			Entity:   nx.Entity,
			Sampler:  nx.Sampler,
			Dataview: nx.Dataview,
			Rows:     true,
			Row:      &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
		if nx.Probe == nil {
			nx.Probe = &Probe{}
		}
		if nx.Entity == nil {
			nx.Entity = &Entity{
				Attributes: make(map[string]string),
			}
		}
		if nx.Sampler == nil {
			nx.Sampler = &Sampler{}
		}
		if nx.Dataview == nil {
			nx.Dataview = &Dataview{}
		}
	case Column:
		nx = XPath{
			Gateway:  nx.Gateway,
			Probe:    nx.Probe,
			Entity:   nx.Entity,
			Sampler:  nx.Sampler,
			Dataview: nx.Dataview,
			Rows:     true,
			Row:      nx.Row,
			Column:   &e,
		}
		if nx.Gateway == nil {
			nx.Gateway = &Gateway{}
		}
		if nx.Probe == nil {
			nx.Probe = &Probe{}
		}
		if nx.Entity == nil {
			nx.Entity = &Entity{
				Attributes: make(map[string]string),
			}
		}
		if nx.Sampler == nil {
			nx.Sampler = &Sampler{}
		}
		if nx.Dataview == nil {
			nx.Dataview = &Dataview{}
		}
		if nx.Row == nil {
			nx.Row = &Row{}
		}
	}

	return &nx
}

// return true is the XPath appears to be empty
func (x *XPath) IsEmpty() bool {
	return x.Gateway == nil
}

// do we need setters? validation?
func (x *XPath) SetGatewayName(gateway string) {
	x.Gateway = &Gateway{Name: gateway}
}

func (x *XPath) IsTableCell() bool {
	return x.Rows && x.Row != nil && x.Column != nil
}

func (x *XPath) IsHeadline() bool {
	return x.Headline != nil && x.Row == nil
}

func (x *XPath) IsDataview() bool {
	return x.Dataview != nil && x.Headline == nil && x.Row == nil
}

func (x *XPath) IsSampler() bool {
	return x.Sampler != nil && x.Dataview == nil
}

func (x *XPath) IsEntity() bool {
	return x.Entity != nil && x.Sampler == nil
}

func (x *XPath) IsProbe() bool {
	return x.Probe != nil && x.Entity == nil
}

func (x *XPath) IsGateway() bool {
	return x.Gateway != nil && x.Probe == nil
}

// return a string representation of an XPath
func (x *XPath) String() (path string) {
	if x.Gateway == nil {
		return
	}
	path += "/geneos/gateway"
	if x.Gateway.Name != "" {
		path += fmt.Sprintf("[(@name=%q)]", x.Gateway.Name)
	}
	path += "/directory"

	if x.Probe == nil {
		return
	}
	path += "/probe"
	if x.Probe.Name != "" {
		path += fmt.Sprintf("[(@name=%q)]", x.Probe.Name)
	}

	if x.Entity == nil {
		return
	}
	path += "/managedEntity"
	if x.Entity.Name != "" {
		path += fmt.Sprintf("[(@name=%q)]", x.Entity.Name)
	}
	for k, v := range x.Entity.Attributes {
		path += fmt.Sprintf("[(attr(%q)=%q)]", k, v)
	}

	if x.Sampler == nil {
		return
	}
	path += "/sampler"
	if x.Sampler.Name != "" {
		path += fmt.Sprintf("[(@name=%q)]", x.Sampler.Name)
	}
	if x.Sampler.Type != nil {
		path += fmt.Sprintf("[(@type=%q)]", *x.Sampler.Type)
	}

	if x.Dataview == nil {
		return
	}
	path += "/dataview"
	if x.Dataview.Name != "" {
		path += fmt.Sprintf("[(@name=%q)]", x.Dataview.Name)
	}

	if x.Rows {
		path += "/rows"
		if x.Row == nil {
			return
		}
		path += "/row"
		if x.Row.Name != "" {
			path += fmt.Sprintf("[(@name=%q)]", x.Row.Name)
		}
		if x.Column == nil {
			return
		}
		path += "/cell"
		if x.Column.Name != "" {
			path += fmt.Sprintf("[(@column=%q)]", x.Column.Name)
		}
	} else {
		if x.Headline == nil {
			return
		}
		path += "/headlines"
		if x.Headline.Name != "" {
			path += fmt.Sprintf("/cell[(@name=%q)]", x.Headline.Name)
		}
	}
	return
}

// parse an absolute xpath as a string into a Geneos XPath structure or
// return an error very simplistic and hardwired to geneos absolute
// xpaths with no support for embedded separators
func ParseAbs(s string) (x *XPath, err error) {
	x = &XPath{}
	parts := splitWithEscaping(s, '/', '\\')

	// walk through path backwards, dropping through each case
	switch p := len(parts); p {
	case 11:
		// table cell, check @column
		if !strings.HasPrefix(parts[p-1], "cell") {
			err = ErrInvalidPath
			return
		}
		column, _ := getAttr(parts[p-1], "column")
		x.Column = &Column{Name: column}
		p--
		fallthrough
	case 10:
		if strings.HasPrefix(parts[p-1], "row") {
			x.Rows = true
			row, _ := getAttr(parts[p-1], "name")
			x.Row = &Row{Name: row}
		} else if strings.HasPrefix(parts[p-1], "cell") {
			headline, _ := getAttr(parts[p-1], "name")
			x.Headline = &Headline{Name: headline}
		} else {
			err = ErrInvalidPath
			return
		}
		p--
		fallthrough
	case 9:
		if strings.HasPrefix(parts[p-1], "rows") {
			// x.Dataview.Column = column
		} else if strings.HasPrefix(parts[p-1], "headlines") {
			//
		} else {
			err = ErrInvalidPath
			return
		}
		p--
		fallthrough
	case 8:
		// dataview
		if !strings.HasPrefix(parts[p-1], "dataview") {
			err = ErrInvalidPath
			return
		}
		dataview, _ := getAttr(parts[p-1], "name")
		x.Dataview = &Dataview{Name: dataview}
		p--
		fallthrough
	case 7:
		// sampler
		if !strings.HasPrefix(parts[p-1], "sampler") {
			err = ErrInvalidPath
			return
		}
		var tp *string
		if t, ok := getAttr(parts[p-1], "type"); ok {
			tp = &t
		}
		name, _ := getAttr(parts[p-1], "name")
		x.Sampler = &Sampler{
			Name: name,
			Type: tp,
		}
		p--
		fallthrough
	case 6:
		// entity
		if !strings.HasPrefix(parts[p-1], "managedEntity") {
			err = ErrInvalidPath
			return
		}
		entity, _ := getAttr(parts[p-1], "name")
		x.Entity = &Entity{
			Name:       entity,
			Attributes: getAttributes(parts[p-1]),
		}
		p--
		fallthrough
	case 5:
		// probe
		if !strings.HasPrefix(parts[p-1], "probe") {
			err = ErrInvalidPath
			return
		}
		probe, _ := getAttr(parts[p-1], "name")
		x.Probe = &Probe{Name: probe}
		p--
		fallthrough
	case 4:
		// directory
		if parts[p-1] != "directory" {
			err = ErrInvalidPath
			return
		}
		p--
		fallthrough
	case 3:
		// gateway
		if !strings.HasPrefix(parts[p-1], "gateway") {
			err = ErrInvalidPath
			return
		}
		gateway, _ := getAttr(parts[p-1], "name")
		x.Gateway = &Gateway{Name: gateway}
		p--
		fallthrough
	case 2:
		// geneos
		if parts[p-1] != "geneos" {
			err = ErrInvalidPath
			return
		}
		p--
		fallthrough
	case 1:
		// leading '/' stripped, must be empty
		if parts[p-1] != "" {
			err = ErrInvalidPath
			return
		}
		return
	default:
		err = ErrInvalidPath
		return
	}
}

// return Xpath as a string
func (x XPath) MarshalJSON() ([]byte, error) {
	return json.Marshal(x.String())
}

// return an xpath parsed from a string
func (x *XPath) UnmarshalJSON(b []byte) (err error) {
	var s string
	if err = json.Unmarshal(b, &s); err != nil {
		return
	}
	nx, err := ParseAbs(s)
	*x = *nx
	return
}

// split a string on separator (byte) except when separator is escaped
// by the escape byte given. typical usage is for when an xpath element
// has an escaped '/'
func splitWithEscaping(s string, separator, escape byte) []string {
	var token []byte
	var tokens []string
	for i := 0; i < len(s); i++ {
		if s[i] == separator {
			tokens = append(tokens, string(token))
			token = token[:0]
		} else if s[i] == escape && i+1 < len(s) {
			i++
			token = append(token, s[i])
		} else {
			token = append(token, s[i])
		}
	}
	tokens = append(tokens, string(token))
	return tokens
}

// extract the value of an attribute from the xpath component in the form
// [(@attr="value")] - this just uses a regexp and applies to validation
func getAttr(s string, attr string) (v string, ok bool) {
	attrRE := regexp.MustCompile(fmt.Sprintf(`\[\(@%s="(.*?)\\{0}?"\)\]`, attr))
	m := attrRE.FindStringSubmatch(s)
	if m == nil {
		return
	}
	if len(m) > 1 {
		v = m[1]
		ok = true
	}
	return
}

var attributeRE = regexp.MustCompile(`\[\(attr\("(.*?)\\{0}?"\)="(.*?)\\{0}?"\)\]`)

// extract the attributes from the xpath managedEntity component in the form
// [(attr("KEY")="value")] - this just uses a regexp and applies to validation
func getAttributes(s string) (attrs map[string]string) {
	attrs = map[string]string{}
	m := attributeRE.FindAllStringSubmatch(s, -1)
	for _, n := range m {
		if len(n) != 3 {
			continue
		}
		attrs[n[1]] = n[2]
	}
	return
}
