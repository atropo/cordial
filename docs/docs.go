/*
Copyright © 2023 ITRS Group

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

// docs creates documentation from the Cobra commands in the various
// tools linked in the imports
package main

import (
	"os"

	pdCmd "github.com/itrs-group/cordial/integrations/pagerduty/cmd"
	snowCmd "github.com/itrs-group/cordial/integrations/servicenow/cmd"
	dv2email "github.com/itrs-group/cordial/tools/dv2email/cmd"

	"github.com/itrs-group/cordial/tools/geneos/cmd"
	_ "github.com/itrs-group/cordial/tools/geneos/cmd/aescmd"
	_ "github.com/itrs-group/cordial/tools/geneos/cmd/cfgcmd"
	_ "github.com/itrs-group/cordial/tools/geneos/cmd/hostcmd"
	_ "github.com/itrs-group/cordial/tools/geneos/cmd/initcmd"
	_ "github.com/itrs-group/cordial/tools/geneos/cmd/pkgcmd"
	_ "github.com/itrs-group/cordial/tools/geneos/cmd/tlscmd"

	"github.com/spf13/cobra"
)

type docs struct {
	command *cobra.Command
	dir     string
}

var doclist = []docs{
	{cmd.GeneosCmd, "../tools/geneos/docs/commands"},

	{dv2email.DV2EMAILCmd, "tools/dv2email"},

	{snowCmd.RootCmd, "integrations/servicenow"},
	{pdCmd.RootCmd, "integrations/pagerduty"},
}

func main() {
	for _, d := range doclist {
		os.MkdirAll(d.dir, 0775)
		if err := GenMarkdownTree(d.command, d.dir); err != nil {
			panic(err)
		}
	}
}
