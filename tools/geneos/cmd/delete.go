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
	"fmt"
	"strings"

	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
	"github.com/itrs-group/cordial/tools/geneos/internal/instance"
	"github.com/spf13/cobra"
)

var deleteCmdStop, deleteCmdForce bool

func init() {
	GeneosCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().BoolVarP(&deleteCmdStop, "stop", "S", false, "Stop instances first")
	deleteCmd.Flags().BoolVarP(&deleteCmdForce, "force", "F", false, "Force delete of protected instances")

	deleteCmd.Flags().SortFlags = false
}

var deleteCmd = &cobra.Command{
	Use:     "delete [flags] [TYPE] [NAME...]",
	GroupID: GROUP_CONFIG,
	Aliases: []string{"rm"},
	Short:   "Delete instances",
	Long: strings.ReplaceAll(`
Delete matching instances.

Instances that are marked |protected| are not deleted without the
|--force|/|-F| option, or they can be unprotected using |geneos
protect -U| first.

Instances that are running are not removed unless the |--stop|/|-S|
option is given.

The instance directory is removed without being backed-up. The user
running the command must have the appropriate permissions and a
partial deletion cannot be protected against.
`, "|", "`"),
	SilenceUsage: true,
	Annotations: map[string]string{
		"wildcard":     "true",
		"needshomedir": "true",
	},
	RunE: func(cmd *cobra.Command, allargs []string) error {
		ct, args, params := CmdArgsParams(cmd)

		return instance.ForAll(ct, Hostname, deleteInstance, args, params)
	},
}

func deleteInstance(c geneos.Instance, params []string) (err error) {
	if instance.IsProtected(c) {
		return geneos.ErrProtected
	}

	if deleteCmdStop {
		if c.Type().RealComponent {
			if err = instance.Stop(c, true, false); err != nil {
				return
			}
		}
	}

	if !instance.IsRunning(c) || deleteCmdForce {
		if err = c.Host().RemoveAll(c.Home()); err != nil {
			return
		}
		fmt.Printf("%s deleted %s:%s\n", c, c.Host().String(), c.Home())
		c.Unload()
		return nil
	}

	return fmt.Errorf("not deleted. Instances must not be running or use the '--force'/'-F' option")
}
