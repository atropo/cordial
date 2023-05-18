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
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/itrs-group/cordial/pkg/config"
	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
	"github.com/itrs-group/cordial/tools/geneos/internal/instance"
)

type showCmdInstanceConfig struct {
	Name      string `json:"name,omitempty"`
	Host      string `json:"host,omitempty"`
	Type      string `json:"type,omitempty"`
	Disabled  bool   `json:"disabled"`
	Protected bool   `json:"protected"`
}

type showCmdConfig struct {
	Instance      showCmdInstanceConfig `json:"instance"`
	Configuration interface{}           `json:"configuration,omitempty"`
}

var showCmdRaw bool

func init() {
	GeneosCmd.AddCommand(showCmd)

	showCmd.Flags().BoolVarP(&showCmdRaw, "raw", "r", false, "Show raw (unexpanded) configuration values")

	showCmd.Flags().SortFlags = false
}

var showCmd = &cobra.Command{
	Use:     "show [flags] [TYPE] [NAME...]",
	GroupID: GROUP_VIEW,
	Short:   "Show instance configuration",
	Long: strings.ReplaceAll(`
Show the configuration for all matching instances.

At the moment this is in JSON format and is output as a single,
concatenated JSON array of object literals, one per instance.

Each instance's underlying configuration is in an object key
|configuration|. Only the objects in this |configuration| key are
stored in the instance's actual configuration file and this is the
root for all parameter names used by other commands, i.e. for a value
under |configuration.licdsecure| the parameter you would use for a
|geneos set| command is just |licdsecure|. Confusingly there is a
|configuration.config| object, used for template support. Other
run-time information is shown under the |instance| key and includes
the instance name, the host it is configured on, it's type and so on.
`, "|", "`"),
	Aliases:      []string{"details"},
	SilenceUsage: true,
	Annotations: map[string]string{
		"wildcard":     "true",
		"needshomedir": "true",
	},
	RunE: func(cmd *cobra.Command, args []string) (err error) {
		ct, args, params := CmdArgsParams(cmd)
		results, err := instance.ForAllWithResults(ct, showInstance, args, params)
		if err != nil {
			if err == os.ErrNotExist {
				return fmt.Errorf("no matching instance found")
			}
			return
		}
		b, _ := json.MarshalIndent(results, "", "    ")
		fmt.Println(string(b))
		return
	},
}

func showInstance(c geneos.Instance, params []string) (result interface{}, err error) {
	// remove aliases
	nv := config.New()
	aliases := c.Type().Aliases
	for _, k := range c.Config().AllKeys() {
		// skip any names in the alias table
		log.Debug().Msgf("checking %s", k)
		if _, ok := aliases[k]; !ok {
			log.Debug().Msgf("setting %s", k)
			nv.Set(k, c.Config().Get(k))
		}
	}

	// XXX wrap in location and type
	as := nv.ExpandAllSettings()
	if showCmdRaw {
		as = nv.AllSettings()
	}
	cf := &showCmdConfig{
		Instance: showCmdInstanceConfig{
			Name:      c.Name(),
			Host:      c.Host().String(),
			Type:      c.Type().String(),
			Disabled:  instance.IsDisabled(c),
			Protected: instance.IsProtected(c),
		},
		Configuration: as,
	}

	result = cf
	return
}
