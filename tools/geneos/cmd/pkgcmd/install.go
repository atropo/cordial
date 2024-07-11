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

package pkgcmd

import (
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/itrs-group/cordial/pkg/config"
	"github.com/itrs-group/cordial/tools/geneos/cmd"
	"github.com/itrs-group/cordial/tools/geneos/internal/geneos"
	"github.com/itrs-group/cordial/tools/geneos/internal/instance"
)

var installCmdLocal, installCmdNoSave, installCmdUpdate, installCmdForce, installCmdNexus, installCmdSnapshot bool
var installCmdBase, installCmdOverride, installCmdVersion, installCmdUsername, installCmdPwFile string
var installCmdDownloadOnly bool
var installCmdPassword *config.Plaintext

func init() {
	packageCmd.AddCommand(installCmd)

	installCmdPassword = &config.Plaintext{}

	installCmd.Flags().StringVarP(&installCmdUsername, "username", "u", "", "Username for downloads, defaults to configuration value in download.username")
	installCmd.Flags().StringVarP(&installCmdPwFile, "pwfile", "P", "", "Password file to read for downloads, defaults to configuration value in download::password or otherwise prompts")
	installCmd.Flags().MarkHidden("pwfile")

	installCmd.Flags().BoolVarP(&installCmdLocal, "local", "L", false, "Install from local files only")
	installCmd.Flags().BoolVarP(&installCmdNoSave, "nosave", "n", false, "Do not save a local copy of any downloads")
	installCmd.Flags().BoolVarP(&installCmdDownloadOnly, "download", "D", false, "Download only")

	// note that "local" and "nosave" together are fine. illogical, but fine.
	installCmd.MarkFlagsMutuallyExclusive("nosave", "download")
	installCmd.MarkFlagsMutuallyExclusive("local", "download")

	installCmd.Flags().BoolVarP(&installCmdUpdate, "update", "U", false, "Update the base directory symlink, will restart unprotected instances")
	installCmd.Flags().BoolVarP(&installCmdForce, "force", "F", false, "Force restart of protected instances, implies --update")

	installCmd.Flags().StringVarP(&installCmdBase, "base", "b", "active_prod", "Override the base active_prod link name")

	installCmd.Flags().StringVarP(&installCmdVersion, "version", "V", "latest", "Download this version, defaults to latest. Doesn't work for EL8 archives.")
	installCmd.Flags().StringVarP(&installCmdOverride, "override", "O", "", "Override the TYPE:VERSION for archive files with non-standard names")

	installCmd.Flags().BoolVarP(&installCmdNexus, "nexus", "N", false, "Download from nexus.itrsgroup.com. Requires auth.")
	installCmd.Flags().BoolVarP(&installCmdSnapshot, "snapshots", "S", false, "Download from nexus snapshots (pre-releases), not releases. Requires -N")

	installCmd.Flags().SortFlags = false
}

//go:embed _docs/install.md
var installCmdDescription string

var installCmd = &cobra.Command{
	Use:   "install [flags] [TYPE] [FILE|URL...]",
	Short: "Install Geneos releases",
	Long:  installCmdDescription,
	Example: strings.ReplaceAll(`
geneos install gateway
geneos install fa2 -V 6.5 -U
geneos install netprobe -b active_dev -U
`, "|", "`"),
	SilenceUsage: true,
	Annotations: map[string]string{
		cmd.AnnotationWildcard:  "false",
		cmd.AnnotationNeedsHome: "false",
	},
	RunE: func(command *cobra.Command, _ []string) (err error) {
		if installCmdDownloadOnly {
			if installCmdLocal || installCmdBase != "active_prod" || installCmdUpdate || installCmdNoSave || installCmdOverride != "" {
				return errors.New("flag --download/-D set with other incompatible options")
			}
			// force localhost
			cmd.Hostname = geneos.LOCALHOST
		} else {
			if geneos.LocalRoot() == "" {
				command.SetUsageTemplate(" ")
				return cmd.GeneosUnsetError
			}
		}

		ct, args, params := cmd.ParseTypeNamesParams(command)

		for _, p := range params {
			if strings.HasPrefix(p, "@") {
				return fmt.Errorf("the @HOST format is not valid here, perhaps you meant `-H HOST`?")
			}
		}

		h := geneos.GetHost(cmd.Hostname)

		if installCmdUsername == "" {
			installCmdUsername = config.GetString(config.Join("download", "username"))
		}

		if installCmdPwFile != "" {
			var pp []byte
			if pp, err = os.ReadFile(installCmdPwFile); err != nil {
				return
			}
			installCmdPassword = config.NewPlaintext(pp)
		} else {
			installCmdPassword = config.GetPassword(config.Join("download", "password"))
		}

		if installCmdUsername != "" && (installCmdPassword.IsNil() || installCmdPassword.Size() == 0) {
			installCmdPassword, err = config.ReadPasswordInput(false, 0)
			if err == config.ErrNotInteractive {
				err = fmt.Errorf("%w and password required", err)
				return
			}
		}

		if installCmdForce {
			installCmdUpdate = true
		}

		// base options
		options := []geneos.PackageOptions{
			geneos.Basename(installCmdBase),
			geneos.DoUpdate(installCmdUpdate || installCmdForce),
			geneos.Force(installCmdForce),
			geneos.LocalOnly(installCmdLocal),
			geneos.NoSave(installCmdNoSave || installCmdLocal),
			geneos.Version(installCmdVersion),
			geneos.OverrideVersion(installCmdOverride),
			geneos.Password(installCmdPassword),
			geneos.Username(installCmdUsername),
			geneos.DownloadOnly(installCmdDownloadOnly),
		}

		if installCmdDownloadOnly {
			archive := "."
			if len(args) > 0 {
				archive = args[0]
			} else if len(params) > 0 {
				archive = params[0]
			}
			log.Debug().Msgf("downloading %q version of %s to %s", installCmdVersion, ct, archive)
			options = append(options,
				geneos.LocalArchive(archive),
			)
			if installCmdSnapshot {
				installCmdNexus = true
				options = append(options, geneos.UseNexusSnapshots())
			}
			if installCmdNexus {
				options = append(options, geneos.UseNexus())
			}
			return install(h, ct, options...)
		}

		cs, err := instance.Instances(h, ct, instance.FilterParameters("protected=true", "version="+installCmdBase))
		if err != nil {
			panic(err)
		}
		if len(cs) > 0 && installCmdUpdate && !installCmdForce {
			fmt.Println("There are one or more protected instances using the current version. Use `--force` to override")
			return
		}

		// record which instances to stop early. once we get to components, we don't know about instances
		if installCmdUpdate {
			instances := []geneos.Instance{}
			allInstances, err := instance.Instances(h, nil)
			if err != nil {
				panic(err)
			}

			for _, ct := range ct.OrList() {
				for _, i := range allInstances {
					if i.Config().GetString("version") != installCmdBase {
						continue
					}
					pkg := i.Config().GetString("pkgtype")
					if pkg != "" && pkg == ct.String() {
						instances = append(instances, i)
						continue
					}
					if i.Type() == ct {
						instances = append(instances, i)
					}
				}
			}
			log.Debug().Msgf("instances to restart: %v", instances)
			options = append(options,
				geneos.Restart(instances...),
				geneos.StartFunc(instance.Start),
				geneos.StopFunc(instance.Stop))
		}

		args = append(args, params...)

		// if we have a component on the command line then use an
		// archive from packages/downloads or download from official web
		// site unless -L is given. version numbers checked. default to
		// 'latest'
		//
		// overrides do not work in this case as the version and type
		// have to be part of the archive file name
		if ct != nil || len(args) == 0 {
			log.Debug().Msgf("installing %q version of %s to %s host(s)", installCmdVersion, ct, cmd.Hostname)

			if installCmdSnapshot {
				installCmdNexus = true
				options = append(options, geneos.UseNexusSnapshots())
			}
			if installCmdNexus {
				options = append(options, geneos.UseNexus())
			}
			return install(h, ct, options...)
		}

		// work through command line args and try to install each
		// argument using the naming format of standard downloads
		for _, source := range args {
			log.Debug().Msgf("installing %q version of %s to %s host(s)", installCmdVersion, ct, cmd.Hostname)
			if err = install(h, ct, append(options, geneos.LocalArchive(source))...); err != nil {
				return err
			}
		}
		return nil
	},
}

func install(h *geneos.Host, ct *geneos.Component, options ...geneos.PackageOptions) (err error) {
	for _, h := range h.OrList() {
		if err = ct.MakeDirs(h); err != nil {
			return err
		}
		for _, ct := range ct.OrList() {
			if err = geneos.Install(h, ct, options...); err != nil {
				if errors.Is(err, fs.ErrExist) {
					fmt.Printf("%s installation already exists, skipping", ct)
					err = nil
					continue
				}
				if errors.Is(err, fs.ErrNotExist) && installCmdVersion != "latest" {
					err = nil
					continue
				}
				return err
			}
		}
	}
	return
}
