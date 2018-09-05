package cmd

import (
	"fmt"

	commands "github.com/ipfs/go-ipfs/core/commands"

	cmds "gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	cmdkit "gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit"
)

var storageHelptext = cmdkit.HelpText{
	Tagline:  "Global p2p merkle-dag storage filesystem.",
	Synopsis: "storage [--config=<config> | -c] [--debug=<debug> | -D] [--help=<help>] [-h=<h>] [--local=<local> | -L] [--api=<api>] <command> ...",
	Subcommands: `
BASIC COMMANDS
add <path>    Add a file to distributed storage
cat <ref>     Show distributed storage object data
get <ref>     Download distributed storage objects
ls <ref>      List links from an object
refs <ref>    List hashes of links from an object

DATA STRUCTURE COMMANDS
block         Interact with raw blocks in the datastore
object        Interact with raw dag nodes
files         Interact with objects as if they were a unix filesystem
dag           Interact with IPLD documents (experimental)

ADVANCED COMMANDS
daemon        Start a long-running daemon process
mount         Mount an IPFS read-only mountpoint
resolve       Resolve any type of name
name          Publish and resolve IPNS names
key           Create and list IPNS name keypairs
dns           Resolve DNS links
pin           Pin objects to local storage
repo          Manipulate the IPFS repository
stats         Various operational stats
p2p           Libp2p stream mounting
filestore     Manage the filestore (experimental)

NETWORK COMMANDS
id            Show info about IPFS peers
bootstrap     Add or remove bootstrap peers
swarm         Manage connections to the p2p network
dht           Query the DHT for values or peers
ping          Measure the latency of a connection
diag          Print diagnostics

TOOL COMMANDS
config        Manage configuration
version       Show distributed storage version information
update        Download and apply go-ipfs updates
commands      List all available commands

Use 'storage <command> --help' to learn more about each command.

distributed storage uses a repository in the local file system. By default, the repo is
located at ~/.ipfs. To change the repo location, set the $IPFS_PATH
environment variable:

export IPFS_PATH=/path/to/ipfsrepo

EXIT STATUS

The CLI will exit with one of the following values:

0     Successful execution.
1     Failed executions.
`,
}

// This is the CLI root, used for executing commands accessible to CLI clients.
// Some subcommands (like 'ipfs daemon' or 'ipfs init') are only accessible here,
// and can't be called through the HTTP API.
var Root = &cmds.Command{
	Options:  commands.Root.Options,
	Helptext: storageHelptext,
}

// commandsClientCmd is the "ipfs commands" command for local cli
var commandsClientCmd = commands.CommandsCmd(Root)

// Commands in localCommands should always be run locally (even if daemon is running).
// They can override subcommands in commands.Root by defining a subcommand with the same name.
var localCommands = map[string]*cmds.Command{
	"commands": commandsClientCmd,
}

func init() {
	// setting here instead of in literal to prevent initialization loop
	// (some commands make references to Root)
	Root.Subcommands = localCommands

	for k, v := range commands.Root.Subcommands {
		if _, found := Root.Subcommands[k]; !found {
			Root.Subcommands[k] = v
		}
	}
}

// NB: when necessary, properties are described using negatives in order to
// provide desirable defaults
type cmdDetails struct {
	cannotRunOnClient bool
	cannotRunOnDaemon bool
	doesNotUseRepo    bool

	// doesNotUseConfigAsInput describes commands that do not use the config as
	// input. These commands either initialize the config or perform operations
	// that don't require access to the config.
	//
	// pre-command hooks that require configs must not be run before these
	// commands.
	doesNotUseConfigAsInput bool

	// preemptsAutoUpdate describes commands that must be executed without the
	// auto-update pre-command hook
	preemptsAutoUpdate bool
}

func (d *cmdDetails) String() string {
	return fmt.Sprintf("on client? %t, on daemon? %t, uses repo? %t",
		d.canRunOnClient(), d.canRunOnDaemon(), d.usesRepo())
}

func (d *cmdDetails) Loggable() map[string]interface{} {
	return map[string]interface{}{
		"canRunOnClient":     d.canRunOnClient(),
		"canRunOnDaemon":     d.canRunOnDaemon(),
		"preemptsAutoUpdate": d.preemptsAutoUpdate,
		"usesConfigAsInput":  d.usesConfigAsInput(),
		"usesRepo":           d.usesRepo(),
	}
}

func (d *cmdDetails) usesConfigAsInput() bool { return !d.doesNotUseConfigAsInput }
func (d *cmdDetails) canRunOnClient() bool    { return !d.cannotRunOnClient }
func (d *cmdDetails) canRunOnDaemon() bool    { return !d.cannotRunOnDaemon }
func (d *cmdDetails) usesRepo() bool          { return !d.doesNotUseRepo }

// "What is this madness!?" you ask. Our commands have the unfortunate problem of
// not being able to run on all the same contexts. This map describes these
// properties so that other code can make decisions about whether to invoke a
// command or return an error to the user.
var cmdDetailsMap = map[string]cmdDetails{
	"init":        {doesNotUseConfigAsInput: true, cannotRunOnDaemon: true, doesNotUseRepo: true},
	"daemon":      {doesNotUseConfigAsInput: true, cannotRunOnDaemon: true},
	"commands":    {doesNotUseRepo: true},
	"version":     {doesNotUseConfigAsInput: true, doesNotUseRepo: true}, // must be permitted to run before init
	"log":         {cannotRunOnClient: true},
	"diag/cmds":   {cannotRunOnClient: true},
	"repo/fsck":   {cannotRunOnDaemon: true},
	"config/edit": {cannotRunOnDaemon: true, doesNotUseRepo: true},
}
