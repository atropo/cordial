Add a new instance of a component `TYPE` with the name `NAME`.

The applicability of the options vary by component `TYPE` and are stored in a configuration file in the instance directory.

The default configuration file format and extension is `json`. There will be support for `yaml` in future releases.
	
The instance will be started after being added if the `--start`/`-S` or `--log`/`-l` option is used. The latter will also follow the log file until interrupted.

Geneos components all use TCP ports either for inbound connections or, in the case of SANs, to identify themselves to the Gateway. The program will choose the next available port from the list in the for each component called `TYPEportrange` (e.g. `gatewayportrange`) in the program configurations. Availability is determined by searching the configurations od all other instances (of any `TYPE`) on the same host. This behaviour can be overridden with the `--port`/`-p` option.

When an instance is started it has an environment made up of the variables in it's configuration file and some necessary defaults, such as `JAVA_HOME`. Additional variables can be set with the `--env`/`-e` option, which can be repeated as many times as required.

File can be imported, just like the `import` command, using one or more `--import`/`-I` options. The syntax is the same as for `import` but because the import source cannot be confused with the `NAME` of the instance using `add` then source can just be a plain file name without the `./` prefix.

The underlying package used by each instance is referenced by a `basename` parameter which defaults to `active_prod`. You can run multiple components of the same type but different releases. You can do this by configuring additional base names in advance with `geneos package update` and then by setting the base name with the `--base`/`-b` option.

For a `TYPE` that supports key files have one created unless one is supplied via the `--keyfile` or `--keycrc` options. The `--keyfile` option uses the file given while the `--keycrc` sets the key file path to a key file with the value given (with or with the `.aes` extension).

Gateways, SANs and Floating probes are given a configuration file based on the templates configured for the different components. The default template can be overridden with the `--template`/`-T` option specifying the source to use. The source can be a local file, a URL or `STDIN`.

Any additional command line arguments are used to set configuration values. Any arguments not in the form `NAME=VALUE` are ignored. Note that `NAME` must be a plain word and must not contain dots (`.`) or double colons (`::`) as these are used as internal delimiters. No component uses hierarchical configuration names except those that can be set by the options above.

You can select the distribution of SAN or Floating Netprobe using the special syntax for the `NAME` in the form `TYPE:NAME`. The only supported `TYPE` at the moment, in addition to the default `netprobe`, is `fa2` allowing you to deploy Fix Analyser 2 based SAN and Floating probes.
