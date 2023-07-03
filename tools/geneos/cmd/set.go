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
	_ "embed"
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/itrs-group/cordial/pkg/config"
	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
	"github.com/itrs-group/cordial/tools/geneos/internal/instance"
)

var setCmdKeyfile config.KeyFile
var setCmdValues = instance.SetConfigValues{}

//go:embed _docs/set.md
var setCmdDescription string

func init() {
	GeneosCmd.AddCommand(setCmd)

	setCmd.Flags().VarP(&setCmdKeyfile, "keyfile", "k", "keyfile to use for encoding secrets\ndefault is instance configured keyfile")

	setCmd.Flags().VarP(&setCmdValues.SecureParams, "secure", "s", "encode a secret for NAME, prompt if VALUE not supplied, using a keyfile")

	setCmd.Flags().VarP(&setCmdValues.Envs, "env", "e", instance.EnvsOptionsText)
	setCmd.Flags().VarP(&setCmdValues.SecureEnvs, "secureenv", "E", "encode a secret for env var NAME, prompt if VALUE not supplied, using a keyfile")
	setCmd.Flags().VarP(&setCmdValues.Includes, "include", "i", instance.IncludeValuesOptionsText)
	setCmd.Flags().VarP(&setCmdValues.Gateways, "gateway", "g", instance.GatewaysOptionstext)
	setCmd.Flags().VarP(&setCmdValues.Attributes, "attribute", "a", instance.AttributesOptionsText)
	setCmd.Flags().VarP(&setCmdValues.Types, "type", "t", instance.TypesOptionsText)
	setCmd.Flags().VarP(&setCmdValues.Variables, "variable", "v", instance.VarsOptionsText)

	setCmd.Flags().SortFlags = false
}

var setCmd = &cobra.Command{
	Use:     "set [flags] [TYPE] [NAME...] [KEY=VALUE...]",
	GroupID: CommandGroupConfig,
	Short:   "Set Instance Parameters",
	Long:    setCmdDescription,
	Example: `
geneos set gateway MyGateway licdsecure=false
geneos set infraprobe -e JAVA_HOME=/usr/lib/java8/jre -e TNS_ADMIN=/etc/ora/network/admin
geneos set -p secret netprobe local1
geneos set ...
`,
	SilenceUsage: true,
	Annotations: map[string]string{
		"wildcard":     "true",
		"needshomedir": "true",
	},
	RunE: func(cmd *cobra.Command, origargs []string) (err error) {
		if len(origargs) == 0 && cmd.Flags().NFlag() == 0 {
			return cmd.Usage()
		}
		ct, args, params := CmdArgsParams(cmd)

		// check if secure args are set, prompt once for each without a supplied value

		if err = promptForSecrets("Parameter", setCmdValues.SecureParams); err != nil {
			return nil
		}
		if err = promptForSecrets("Environment Variable", setCmdValues.SecureEnvs); err != nil {
			return nil
		}

		return Set(ct, args, params)
	},
}

func Set(ct *geneos.Component, args, params []string) error {
	return instance.ForAll(ct, Hostname, setInstance, args, params)
}

func setInstance(c geneos.Instance, params []string) (err error) {
	log.Debug().Msgf("c %s params %v", c, params)

	cf := c.Config()

	setCmdValues.Params = params

	if err = instance.SetInstanceValues(c, setCmdValues, setCmdKeyfile); err != nil {
		return
	}

	if cf.Type == "rc" {
		err = instance.Migrate(c)
	} else {
		err = cf.Save(c.Type().String(),
			config.Host(c.Host()),
			config.SaveDir(instance.ParentDirectory(c)),
			config.SetAppName(c.Name()),
		)
	}

	return
}

func promptForSecrets(prompt string, v instance.SecureValues) (err error) {
	for _, s := range v {
		if s.Plaintext.IsNil() {
			// prompt
			s.Plaintext, err = config.ReadPasswordInput(true, 3,
				fmt.Sprintf("Enter Secret for %s %q", prompt, s.Value),
				fmt.Sprintf("Re-enter Secret for %s %q", prompt, s.Value),
			)
			if err != nil {
				return
			}
		}
		// v[i] = s
	}
	return
}
