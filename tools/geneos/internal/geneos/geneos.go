package geneos

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/itrs-group/cordial/pkg/config"
	"github.com/itrs-group/cordial/tools/geneos/internal/host"
	"github.com/itrs-group/cordial/tools/geneos/internal/utils"
)

var (
	ErrInvalidArgs  error = errors.New("invalid arguments")
	ErrNotSupported error = errors.New("not supported")
	ErrDisabled     error = errors.New("disabled")
	ErrIsADirectory error = errors.New("is a directory")
)

const RootCAFile = "rootCA"
const SigningCertFile = "geneos"
const DisableExtension = "disabled"

var ConfigFileType = "json"

var GlobalConfigDir = "/etc"
var ConfigSubdirName = "geneos"
var UserConfigFile = "geneos.json"
var GlobalConfigPath = filepath.Join(GlobalConfigDir, ConfigSubdirName, UserConfigFile)

// initialise a Geneos environment.
//
// creates a directory hierarchy and calls the initialisation
// functions for each component, for example to create templates
//
// if the directory is not empty and 'noEmptyOK' is false then
// nothing is changed
func Init(h *host.Host, options ...GeneosOptions) (err error) {
	var uid, gid int

	if h != host.LOCAL && utils.IsSuperuser() {
		err = ErrNotSupported
		return
	}

	opts := EvalOptions(options...)
	if opts.homedir == "" {
		log.Fatal().Msg("homedir not set")
		// default or error
	}

	// dir must first not exist (or be empty) and then be creatable
	//
	// maybe check that the entire list of registered directories are
	// either directories or do not exist
	if _, err := h.Stat(opts.homedir); err != nil {
		if err = h.MkdirAll(opts.homedir, 0775); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	} else if !opts.force {
		// check empty
		dirs, err := h.ReadDir(opts.homedir)
		if err != nil {
			log.Fatal().Err(err).Msg("")
		}
		for _, entry := range dirs {
			if !strings.HasPrefix(entry.Name(), ".") {
				if h != host.LOCAL {
					log.Debug().Msg("remote directories exist, exiting init")
					return nil
				}
				log.Fatal().Msgf("target directory %q exists and is not empty", opts.homedir)
			}
		}
	}

	if h == host.LOCAL {
		config.GetConfig().Set("geneos", opts.homedir)
		config.GetConfig().Set("defaultuser", opts.localusername)

		userConfFile := UserConfigFilePaths()[0]
		if utils.IsSuperuser() {
			userConfDir, err := config.UserConfigDir(opts.localusername)
			if err != nil {
				log.Fatal().Err(err).Msg("")
			}
			userConfFile = filepath.Join(userConfDir, ConfigSubdirName, UserConfigFile)
		}

		if err = host.WriteConfigFile(userConfFile, opts.localusername, 0664, config.GetConfig()); err != nil {
			return err
		}

		// recreate host.LOCAL to load "geneos" and others
		host.LOCAL = nil
		host.LOCAL = host.Get(host.LOCALHOST)
		h = host.LOCAL
	}

	if utils.IsSuperuser() {
		uid, gid, _, err = utils.GetIDs(opts.localusername)
		if err != nil {
			// do something
		}
		if err = host.LOCAL.Chown(opts.homedir, uid, gid); err != nil {
			log.Fatal().Err(err).Msg("")
		}
	}

	// it's not an error to try to re-create existing dirs
	// if err = Root.MakeComponentDirs(h); err != nil {
	// 	return
	// }

	for _, c := range AllComponents() {
		if err := c.MakeComponentDirs(h); err != nil {
			continue
		}
		if c.Initialise != nil {
			c.Initialise(h, c)
		}
	}

	// if we've created directory paths as root, go through and change
	// ownership to the tree
	if utils.IsSuperuser() {
		err = filepath.WalkDir(opts.homedir, func(path string, dir fs.DirEntry, err error) error {
			if err == nil {
				err = host.LOCAL.Chown(path, uid, gid)
			}
			return err
		})
	}

	return
}

// read a local configuration file without the need for a host
// connection, primarily for bootstrapping
func ReadLocalConfigFile(file string, config interface{}) (err error) {
	jsonFile, err := os.ReadFile(file)
	if err != nil {
		return
	}

	// dec := json.NewDecoder(jsonFile)
	return json.Unmarshal(jsonFile, &config)
}

// UserConfigFilePaths returns a slice of all the possible file paths to
// the user configuration file. If arguments are passed then they are
// used, in-turn, as the base filename for each directory. If no
// arguments are passed then the default filename is UserConfigFile. The
// first element is the preferred file and the one that should be used
// to write to.
//
// This function can be used to ensure that as the location changes in
// the future, the code can still look for older copies when the
// preferred path is empty.
func UserConfigFilePaths(bases ...string) (paths []string) {
	userConfDir, err := config.UserConfigDir()
	if err != nil {
		log.Fatal().Err(err).Msg("")
	}

	if len(bases) == 0 {
		bases = []string{UserConfigFile}
	}

	for _, base := range bases {
		paths = append(paths, filepath.Join(userConfDir, ConfigSubdirName, base))
		paths = append(paths, filepath.Join(userConfDir, base))
	}
	return
}
