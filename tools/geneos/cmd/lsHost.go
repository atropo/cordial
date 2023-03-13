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
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/itrs-group/cordial/pkg/config"
	"github.com/itrs-group/cordial/tools/geneos/internal/host"
	"github.com/spf13/cobra"
)

type lsHostCmdType struct {
	Name      string
	Username  string
	Hostname  string
	Port      int64
	Directory string
}

var lsHostCmdJSON, lsHostCmdIndent, lsHostCmdCSV bool

var lsHostCmdEntries []lsHostCmdType

var lsHostCSVWriter *csv.Writer

func init() {
	lsCmd.AddCommand(lsHostCmd)

	lsHostCmd.Flags().BoolVarP(&lsHostCmdJSON, "json", "j", false, "Output JSON")
	lsHostCmd.Flags().BoolVarP(&lsHostCmdIndent, "pretty", "i", false, "Output indented JSON")
	lsHostCmd.Flags().BoolVarP(&lsHostCmdCSV, "csv", "c", false, "Output CSV")

	lsHostCmd.Flags().SortFlags = false
}

var lsHostCmd = &cobra.Command{
	Use:     "host [flags] [TYPE] [NAME...]",
	Aliases: []string{"hosts", "remote", "remotes"},
	Short:   "List hosts, optionally in CSV or JSON format",
	Long: strings.ReplaceAll(`
List the matching remote hosts.
`, "|", "`"),
	SilenceUsage: true,
	Annotations: map[string]string{
		"wildcard": "false",
	},
	RunE: func(cmd *cobra.Command, _ []string) (err error) {
		// ct, args, params := cmdArgsParams(cmd)
		switch {
		case lsHostCmdJSON, lsHostCmdIndent:
			lsHostCmdEntries = []lsHostCmdType{}
			err = loopHosts(lsInstanceJSONHosts)
			var b []byte
			if lsHostCmdIndent {
				b, _ = json.MarshalIndent(lsHostCmdEntries, "", "    ")
			} else {
				b, _ = json.Marshal(lsHostCmdEntries)
			}
			fmt.Println(string(b))
		case lsHostCmdCSV:
			lsHostCSVWriter = csv.NewWriter(os.Stdout)
			lsHostCSVWriter.Write([]string{"Type", "Name", "Disabled", "Username", "Hostname", "Port", "Directory"})
			err = loopHosts(lsInstanceCSVHosts)
			lsHostCSVWriter.Flush()
		default:
			lsTabWriter = tabwriter.NewWriter(os.Stdout, 3, 8, 2, ' ', 0)
			fmt.Fprintf(lsTabWriter, "Name\tUsername\tHostname\tPort\tDirectory\n")
			err = loopHosts(lsInstancePlainHosts)
			lsTabWriter.Flush()
		}
		if err == os.ErrNotExist {
			err = nil
		}
		return
	},
}

func loopHosts(fn func(*host.Host) error) error {
	for _, h := range host.RemoteHosts() {
		fn(h)
	}
	return nil
}

func lsInstancePlainHosts(h *host.Host) (err error) {
	fmt.Fprintf(lsTabWriter, "%s\t%s\t%s\t%d\t%s\n", h.GetString("name"), h.GetString("username"), h.GetString("hostname"), h.GetInt("port", config.Default(22)), h.GetString("geneos"))
	return
}

func lsInstanceCSVHosts(h *host.Host) (err error) {
	lsHostCSVWriter.Write([]string{h.String(), h.GetString("username"), h.GetString("hostname"), fmt.Sprint(h.GetInt("port"), config.Default(22)), h.GetString("geneos")})
	return
}

func lsInstanceJSONHosts(h *host.Host) (err error) {
	lsHostCmdEntries = append(lsHostCmdEntries, lsHostCmdType{h.String(), h.GetString("username"), h.GetString("hostname"), h.GetInt64("port", config.Default(22)), h.GetString("geneos")})
	return
}
