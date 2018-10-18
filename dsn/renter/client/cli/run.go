package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"
	"path/filepath"
	cmds "gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	ipfscli "gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds/cli"
	"github.com/ipfs/go-ipfs/core/coreunix"
	"github.com/ecoball/go-ecoball/dsn/renter"
	"github.com/ecoball/go-ecoball/dsn/common"
	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
)

// ExitError is the error used when a specific exit code needs to be returned.
type ExitError int

func (e ExitError) Error() string {
	return fmt.Sprintf("exit code %d", int(e))
}

// Closer is a helper interface to check if the env supports closing
type Closer interface {
	Close()
}

func addRun(ctx context.Context, root *cmds.Command,
	cmdline []string, stdin, stdout, stderr *os.File,
	buildEnv cmds.MakeEnvironment, makeExecutor cmds.MakeExecutor) (*renter.RscReq, error) {

	printErr := func(err error) {
		fmt.Fprintf(stderr, "Error: %s\n", err)
	}

	req, errParse := ipfscli.Parse(ctx, cmdline[2:], stdin, root)

	// Handle the timeout up front.
	var cancel func()
	if timeoutStr, ok := req.Options[cmds.TimeoutOpt]; ok {
		timeout, err := time.ParseDuration(timeoutStr.(string))
		if err != nil {
			return nil, err
		}
		req.Context, cancel = context.WithTimeout(req.Context, timeout)
	} else {
		req.Context, cancel = context.WithCancel(req.Context)
	}
	defer cancel()

	// this is a message to tell the user how to get the help text
	//printMetaHelp := func(w io.Writer) {
	//	cmdPath := strings.Join(req.Path, " ")
	//	fmt.Fprintf(w, "Use '%s %s --help' for information about this command\n", cmdline[0], cmdPath)
	//}

	printHelp := func(long bool, w io.Writer) {
		helpFunc := ipfscli.ShortHelp
		if long {
			helpFunc = ipfscli.LongHelp
		}

		var path []string
		if req != nil {
			path = req.Path
		}

		if err := helpFunc(cmdline[0], root, path, w); err != nil {
			// This should not happen
			panic(err)
		}
	}

	// BEFORE handling the parse error, if we have enough information
	// AND the user requested help, print it out and exit
	err := ipfscli.HandleHelp(cmdline[0], req, stdout)
	if err == nil {
		return nil, nil
	} else if err != ipfscli.ErrNoHelpRequested {
		return nil, err
	}
	// no help requested, continue.

	// ok now handle parse error (which means cli input was wrong,
	// e.g. incorrect number of args, or nonexistent subcommand)
	if errParse != nil {
		printErr(errParse)

		// this was a user error, print help
		if req != nil && req.Command != nil {
			fmt.Fprintln(stderr) // i need some space
			printHelp(false, stderr)
		}

		return nil, err
	}

	// here we handle the cases where
	// - commands with no Run func are invoked directly.
	// - the main command is invoked.
	if req == nil || req.Command == nil || req.Command.Run == nil {
		printHelp(false, stdout)
		return nil, nil
	}

	cmd := req.Command

	env, err := buildEnv(req.Context, req)
	if err != nil {
		printErr(err)
		return nil, err
	}
	if c, ok := env.(Closer); ok {
		defer c.Close()
	}

	exctr, err := makeExecutor(req, env)
	if err != nil {
		printErr(err)
		return nil, err
	}

	//var (
	//	re     cmds.ResponseEmitter
	//	exitCh <-chan int
	//)

	encTypeStr, _ := req.Options[cmds.EncLong].(string)
	encType := cmds.EncodingType(encTypeStr)

	// use JSON if text was requested but the command doesn't have a text-encoder
	if _, ok := cmd.Encoders[encType]; encType == cmds.Text && !ok {
		req.Options[cmds.EncLong] = cmds.JSON
	}

	fpath := cmdline[3]
	fpath = filepath.ToSlash(filepath.Clean(fpath))
	//fmt.Println(fpath)
	stat, err := os.Lstat(fpath)
	if err != nil {
		return nil, err
	}
	//fmt.Println(fpath, stat.Size())
	var PieceSize uint64
	if stat.Size() < common.EraDataPiece * chunker.DefaultBlockSize {
		PieceSize = uint64(stat.Size() / common.EraDataPiece)
	} else {
		PieceSize = uint64(chunker.DefaultBlockSize)
	}
	if stat.Size() < common.EraDataPiece * chunker.DefaultBlockSize {
		req.Options["chunker"] = fmt.Sprintf("size-%d", PieceSize)
	}

	cre, reponse := cmds.NewChanResponsePair(req)
	errCh := make(chan error, 1)
	go func() {
		err := exctr.Execute(req, cre, env)
		if err != nil {
			errCh <- err
		}
	}()

	var object *coreunix.AddedObject
	i := 0
	for {
		v, err := reponse.RawNext()
		switch err {
		case nil:
			// all good, go on
		case io.EOF:
			cre.Close()
			//return nil, nil
			break
		default:
			return nil, err
		}
		if i == 1 {
			object = v.(*coreunix.AddedObject)
			break
		}
		i++
	}
	fmt.Println(object.Name, object.Hash, object.Size, object.Bytes)
	eraReq := new(renter.RscReq)
	eraReq.Cid = object.Hash
	eraReq.Redundency = 2
	eraReq.FileSize = uint64(stat.Size())
	eraReq.IsDir = false
	eraReq.Chunk = PieceSize
	return eraReq, nil
}


func catRun(ctx context.Context, root *cmds.Command,
	cmdline []string, stdin, stdout, stderr *os.File,
	buildEnv cmds.MakeEnvironment, makeExecutor cmds.MakeExecutor) error {

	printErr := func(err error) {
		fmt.Fprintf(stderr, "Error: %s\n", err)
	}
	cid := cmdline[3]
	cmdline[3] = cmdline[3] + "/file"
	req, errParse := ipfscli.Parse(ctx, cmdline[2:], stdin, root)

	// Handle the timeout up front.
	var cancel func()
	if timeoutStr, ok := req.Options[cmds.TimeoutOpt]; ok {
		timeout, err := time.ParseDuration(timeoutStr.(string))
		if err != nil {
			return err
		}
		req.Context, cancel = context.WithTimeout(req.Context, timeout)
	} else {
		req.Context, cancel = context.WithCancel(req.Context)
	}
	defer cancel()

	// this is a message to tell the user how to get the help text
	/*printMetaHelp := func(w io.Writer) {
		cmdPath := strings.Join(req.Path, " ")
		fmt.Fprintf(w, "Use '%s %s --help' for information about this command\n", cmdline[0], cmdPath)
	}*/

	printHelp := func(long bool, w io.Writer) {
		helpFunc := ipfscli.ShortHelp
		if long {
			helpFunc = ipfscli.LongHelp
		}

		var path []string
		if req != nil {
			path = req.Path
		}

		if err := helpFunc(cmdline[0], root, path, w); err != nil {
			// This should not happen
			panic(err)
		}
	}

	// BEFORE handling the parse error, if we have enough information
	// AND the user requested help, print it out and exit
	err := ipfscli.HandleHelp(cmdline[0], req, stdout)
	if err == nil {
		return nil
	} else if err != ipfscli.ErrNoHelpRequested {
		return err
	}
	// no help requested, continue.

	// ok now handle parse error (which means cli input was wrong,
	// e.g. incorrect number of args, or nonexistent subcommand)
	if errParse != nil {
		printErr(errParse)

		// this was a user error, print help
		if req != nil && req.Command != nil {
			fmt.Fprintln(stderr) // i need some space
			printHelp(false, stderr)
		}

		return err
	}

	// here we handle the cases where
	// - commands with no Run func are invoked directly.
	// - the main command is invoked.
	if req == nil || req.Command == nil || req.Command.Run == nil {
		printHelp(false, stdout)
		return nil
	}

	cmd := req.Command

	env, err := buildEnv(req.Context, req)
	if err != nil {
		printErr(err)
		return err
	}
	if c, ok := env.(Closer); ok {
		defer c.Close()
	}

	exctr, err := makeExecutor(req, env)
	if err != nil {
		printErr(err)
		return err
	}

	var (
		re     cmds.ResponseEmitter
		exitCh <-chan int
	)

	encTypeStr, _ := req.Options[cmds.EncLong].(string)
	encType := cmds.EncodingType(encTypeStr)

	// use JSON if text was requested but the command doesn't have a text-encoder
	if _, ok := cmd.Encoders[encType]; encType == cmds.Text && !ok {
		req.Options[cmds.EncLong] = cmds.JSON
	}

	// first if condition checks the command's encoder map, second checks global encoder map (cmd vs. cmds)
	if enc, ok := cmd.Encoders[encType]; ok {
		re, exitCh = ipfscli.NewResponseEmitter(stdout, stderr, enc, req)
	} else if enc, ok := cmds.Encoders[encType]; ok {
		re, exitCh = ipfscli.NewResponseEmitter(stdout, stderr, enc, req)
	} else {
		return fmt.Errorf("could not find matching encoder for enctype %#v", encType)
	}

	errCh := make(chan error, 1)
	go func() {
		err := exctr.Execute(req, re, env)
		if err != nil {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		printErr(err)

		//if kiterr, ok := err.(*cmdkit.Error); ok {
		//	err = *kiterr
		//}
		//if kiterr, ok := err.(cmdkit.Error); ok && kiterr.Code == cmdkit.ErrClient {
		//	printMetaHelp(stderr)
		//}
		err = eraCat(cid)
		return err

	case code := <-exitCh:
		if code != 0 {
			return ExitError(code)
		}
	}

	return nil
}

