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

// Package cordial is a collection of tools, packages and integrations for
// Geneos written primarily in Go
package cordial

import (
	_ "embed" // embed the VERSION in the top-level package
	"html"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/glamour"
	"golang.org/x/term"

	"github.com/spf13/cobra"
)

// VERSION is a semi-global string variable
//
//go:embed VERSION
var VERSION string

func renderMD(in string) (out string) {
	var width int = 80
	var err error

	style := glamour.WithAutoStyle()
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		style = glamour.WithStandardStyle("ascii")
	} else {
		width, _, err = term.GetSize(int(os.Stdout.Fd()))
		if err != nil {
			width = 80
		}
		if width > 132 {
			width = 132
		}
	}

	tr, err := glamour.NewTermRenderer(
		style,
		glamour.WithStylesFromJSONBytes([]byte(`{ "document": { "margin": 2 } }`)),
		glamour.WithWordWrap(width-4),
		glamour.WithEmoji(),
	)
	if err != nil {
		return in
	}
	out, err = tr.Render(in)
	if err != nil {
		return in
	}
	out = html.UnescapeString(out)
	return
}

// RenderHelpAsMD updated the given command to use glamour to render the
// command's Long description as markdown formatted text to an ANSI
// terminal.
func RenderHelpAsMD(command *cobra.Command) {
	// render help with glamour
	cobra.AddTemplateFunc("md", renderMD)
	// start on first line, one line gap, usage without leading spaces
	command.SetHelpTemplate(`{{with (or .Long .Short)}}{{. | md | trimTrailingWhitespaces}}{{end}}

{{if or .Runnable .HasSubCommands}}{{.UsageString}}{{end}}
`)
}

// ExecutableName returns a processed executable name. Symlinks are
// resolved and the basename of the resulting file has, at most, one
// extension removed and then if there is a '-' followed by the matching
// cordial version string. Note this final version string must match
// first version argument or the cordial one compiled into the binary.
//
// For example:
//
//	`geneos.exe` -> `geneos`
//	`dv2email-v1.10.0` -> `dv2email`
func ExecutableName(version ...string) (execname string) {
	execname, _ = os.Executable()
	execname, _ = filepath.EvalSymlinks(execname)
	execname = path.Base(filepath.ToSlash(execname))

	// strip any extension from the binary, to allow windows .EXE
	// binary to work. Note we get the extension first, it may be
	// capitalised. This will also remove any other extensions, users
	// should use '-' or '_' instead.
	execname = strings.TrimSuffix(execname, path.Ext(execname))

	// finally strip the VERSION, if found, prefixed by a dash, on the
	// end of the basename
	//
	// this way you can run a versioned binary and still see the right
	// config files
	if len(version) > 0 {
		execname = strings.TrimSuffix(execname, "-"+version[0])
	} else {
		execname = strings.TrimSuffix(execname, "-"+VERSION)
	}
	return
}
