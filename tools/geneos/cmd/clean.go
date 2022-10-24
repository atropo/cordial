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

package cmd

import (
	"strings"

	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
	"github.com/itrs-group/cordial/tools/geneos/internal/instance"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(cleanCmd)

	cleanCmd.Flags().BoolVarP(&cleanCmdFull, "full", "F", false, "Perform a full clean. Removes more files than basic clean and restarts instances")
	cleanCmd.Flags().SortFlags = false
}

var cleanCmdFull bool

var cleanCmd = &cobra.Command{
	Use:   "clean [flags] [TYPE] [NAME...]",
	Short: "Clean-up instance directories",
	Long: strings.ReplaceAll(`
Clean-up instance directories, also restarting instances if doing a
full clean using |-F|. The patterns of files and directories that are
cleaned up are set in the global configuration as |[TYPE]CleanList|
and |[TYPE]PurgeList| and can be seen using the |geneos show|
command, and changed using |geneos set|. The format is a
PathListSeparator (typically a colon) separated list of file globs.
`, "|", "`"),
	Example: strings.ReplaceAll(`
# delete old logs and config file backups without affecting running instance
geneos clean gateway Gateway1
# stop all netprobes and remove all non-essential files from working directories,
# then restart
geneos clean --full netprobe
`, "|", "`"),
	SilenceUsage: true,
	Annotations: map[string]string{
		"wildcard": "true",
	},
	RunE: func(cmd *cobra.Command, _ []string) error {
		ct, args, params := cmdArgsParams(cmd)
		return instance.ForAll(ct, cleanInstance, args, params)
	},
}

func cleanInstance(c geneos.Instance, params []string) (err error) {
	return instance.Clean(c, geneos.Restart(cleanCmdFull))
}
