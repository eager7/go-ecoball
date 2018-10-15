package main

import (
	"context"
	"fmt"
	"io"
	"os"
	//"strings"
	"time"

	cmds "gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds"
	//"gx/ipfs/QmdE4gMduCKCGAcczM2F5ioYDfdeKuPix138wrES1YSr7f/go-ipfs-cmdkit"
	ipfscli "gx/ipfs/QmNueRyPRQiV7PUEpnP4GgGLuK1rKQLaRW7sfPvUetYig1/go-ipfs-cmds/cli"
	//"bytes"
	//"github.com/ipfs/go-ipfs/core/coreunix"
	//"bytes"
	"github.com/ipfs/go-ipfs/core/coreunix"
	//"github.com/ontio/ontology/p2pserver/actor/req"
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
	buildEnv cmds.MakeEnvironment, makeExecutor cmds.MakeExecutor) error {

	printErr := func(err error) {
		fmt.Fprintf(stderr, "Error: %s\n", err)
	}

	req, errParse := ipfscli.Parse(ctx, cmdline[1:], stdin, root)

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

	//var buf bytes.Buffer
	//pbuf := make([]byte, 1024)
	//buf := bytes.NewBuffer(nil)
	//wc := writecloser{Writer: buf, Closer: nopCloser{}}
	// first if condition checks the command's encoder map, second checks global encoder map (cmd vs. cmds)
	//if enc, ok := cmd.Encoders[encType]; ok {
	//	re, exitCh = ipfscli.NewResponseEmitter(stdout, stderr, enc, req)
	//} else if enc, ok := cmds.Encoders[encType]; ok {
	//	re, exitCh = ipfscli.NewResponseEmitter(stdout, stderr, enc, req)
	//} else {
	//	return fmt.Errorf("could not find matching encoder for enctype %#v", encType)
	//}

	//wre := cmds.NewWriterResponseEmitter(wc, req, cmds.Encoders[cmds.JSON])
	//r, w := io.Pipe()
	//tre := cmds.NewWriterResponseEmitter(w, req, cmds.Encoders[cmds.JSON])
	//tres := cmds.NewReaderResponse(r, cmds.JSON, req)

	//rres := cmds.NewReaderResponse(buf, cmds.JSON, req)

	cre, reponse := cmds.NewChanResponsePair(req)
	errCh := make(chan error, 1)
	go func() {
		err := exctr.Execute(req, cre, env)
		if err != nil {
			errCh <- err
		}
	}()
	/*v, err := reponse.RawNext()
	object := v.(*coreunix.AddedObject)
	fmt.Println(object)*/

	for {
		v, err := reponse.RawNext()
		switch err {
		case nil:
			// all good, go on
		case io.EOF:
			cre.Close()
			return nil
		default:
			return err
		}
		fmt.Println("***********")
		object := v.(*coreunix.AddedObject)
		fmt.Println(object)
	}

	fmt.Println(req.Files.FullPath())

	return nil
}

