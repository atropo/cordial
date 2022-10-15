package geneos

type Options struct {
	nosave        bool
	local         bool
	overwrite     bool
	override      string
	restart       bool
	version       string
	basename      string
	homedir       string
	localusername string
	username      string
	password      string
	platform_id   string
	downloadbase  string
	downloadtype  string
	source        string
}

type GeneosOptions func(*Options)

func EvalOptions(options ...GeneosOptions) (d *Options) {
	// defaults
	d = &Options{
		downloadbase: "releases",
		downloadtype: "resources",
	}
	for _, opt := range options {
		opt(d)
	}
	return
}

// NoSave prevents downloads from being saved in the archive directory
func NoSave(n bool) GeneosOptions {
	return func(d *Options) { d.nosave = n }
}

// LocalOnly stops downloads from external locations
func LocalOnly(l bool) GeneosOptions {
	return func(d *Options) { d.local = l }
}

// Force ignores existing directories or files
func Force(o bool) GeneosOptions {
	return func(d *Options) { d.overwrite = o }
}

// OverrideVersion forces a specific version to be used and failure if not available
func OverrideVersion(s string) GeneosOptions {
	return func(d *Options) { d.override = s }
}

// Restart sets the instances to be restarted around the update
func Restart(r bool) GeneosOptions {
	return func(d *Options) { d.restart = r }
}

// Restart returns the value of the Restart option. This is a helper to
// allow checking outside the cmd package.
// XXX Currently doesn't work.
func (d *Options) Restart() bool {
	return d.restart
}

// Version sets the desired version number, defaults to "latest" in most
// cases. The version number is in the form `[GA]X.Y.Z` (or `RA` for
// snapshots)
func Version(v string) GeneosOptions {
	return func(d *Options) { d.version = v }
}

// Basename sets the package binary basename, defaults to active_prod,
// for symlinks for update.
func Basename(b string) GeneosOptions {
	return func(d *Options) { d.basename = b }
}

// Homedir sets the Geneos installation home directory (aka `geneos` in
// the settings)
func Homedir(h string) GeneosOptions {
	return func(d *Options) { d.homedir = h }
}

// LocalUsername is the user name of the user running the program, or if
// running as root the default username that shiould be used. This is
// different to any remote username for executing commands on remote
// hosts.
func LocalUsername(u string) GeneosOptions {
	return func(d *Options) { d.localusername = u }
}

// Username is the remote access username for downloads
func Username(u string) GeneosOptions {
	return func(d *Options) { d.username = u }
}

// Password is the remote access password for downloads
func Password(p string) GeneosOptions {
	return func(d *Options) { d.password = p }
}

// PlauformID sets the (Linux) platform ID from the OS release info.
// Currently used to distinguish RHEL8 installs from others.
func PlatformID(id string) GeneosOptions {
	return func(d *Options) { d.platform_id = id }
}

// UseNexus sets the flag to use nexus.itrsgroup.com for internal
// downloads instead of the default download URL in the settings. This
// also influences the way the remote path is searched and build, not
// just the base URL.
func UseNexus() GeneosOptions {
	return func(d *Options) { d.downloadtype = "nexus" }
}

// UseSnapshots set the flag to use Nexus Snapshots rather than
// Releases.
func UseSnapshots() GeneosOptions {
	return func(d *Options) { d.downloadbase = "snapshots" }
}

// Source is the source of the installation and overrides all other
// settings include Local and download URLs.
func Source(f string) GeneosOptions {
	return func(d *Options) { d.source = f }
}
