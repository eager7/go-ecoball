package cli

import (
	"context"
	rclient "github.com/ecoball/go-ecoball/dsn/renter/client"
	"os"
	"fmt"
)

/*
func init() {
	delete(dsncmd.Root.Subcommands, "commands")
}

type writecloser struct {
	io.Writer
	io.Closer
}
type nopCloser struct{}

func (c nopCloser) Close() error { return nil }

//renter add/cat
func main0() {
	action := os.Args[1]
	if action != "add" && action != "cat" {
		panic("unkonwn cmd")
	}

	//era := os.Args[2]
	//redundency, err := strconv.Atoi(era)
	//if err != nil {
	//	panic(err)
	//}
	rsc := RscReq{
		Redundency: 2,
	}
	//if action == "add" {
	//	os.Args[2] = "add"
	//}
	// parse the command path, arguments and options from the command line
	req, err := cli.Parse(context.TODO(), os.Args[1:], os.Stdin, ipfscmd.RootRO)
	if err != nil {
		panic(err)
	}
	for _, v := range req.Path {
		fmt.Println(v)
	}
	for _, v := range req.Arguments {
		fmt.Println(v)
	}
	fmt.Println(req.Files.FileName())
	if req.Options["chunk"] != nil {
		rsc.Chunk = req.Options["chunk"].(string)
	}

	for _, v :=  range os.Args {
		rFlag := strings.Contains(v, "-r")
		rFlag = strings.Contains(v, "--recursive")
		if rFlag {
			rsc.IsDir = true
		}
	}

	// create http rpc client
	client := http.NewClient(":5001")

	// send request to server
	res, err := client.Send(req)
	if err != nil {
		panic(err)
	}

	req.Options["encoding"] = cmds.Text
	req.Command.Type = coreunix.AddedObject{}
	//buf := bytes.NewBuffer(nil)
	//wc := writecloser{Writer: buf, Closer: nopCloser{}}
	// create an emitter
	//re, retCh := cli.NewResponseEmitter(buf, os.Stderr, req.Command.Encoders["Json"], req)
	re, retCh := cli.NewResponseEmitter(os.Stdout, os.Stderr, req.Command.Encoders[cmds.Text], req)
	//rsp := cmds.NewWriterResponseEmitter(wc, req, cmds.Encoders[cmds.JSON])
	if pr, ok := req.Command.PostRun[cmds.CLI]; ok {
		re = pr(req, re)
	}

	//var result coreunix.AddedObject
	//err = json.Unmarshal(buf.Bytes(), &result)
	//if err != nil {
	//	panic(err)
	//}


	wait := make(chan struct{})
	// copy received result into cli emitter
	go func() {
		err = cmds.Copy(re, res)
		if err != nil {
			re.SetError(err, cmdkit.ErrNormal|cmdkit.ErrFatal)
		}
		close(wait)
	}()
	//fmt.Println(len(buf.Bytes()))
	//fmt.Println(buf.String())
	// wait until command has returned and exit
	ret := <-retCh
	<-wait
	os.Exit(ret)
}*/
func add()  {
	conf := rclient.InitDefaultConf()
	ctx := context.Background()
	appClient := rclient.NewRenter(ctx, conf)
	file := os.Args[3]
	ok := appClient.CheckCollateral()
	if !ok {
		fmt.Println("Checking collateral failed")
		return
	}
	cid, err := CliAddFile()
	if err != nil {
		panic(err)
	}
	cid, err = appClient.RscCodingReq(file, cid)
	if err != nil {
		panic(err)
	}
	appClient.InvokeFileContract(file, cid)
	appClient.PayForFile(file, cid)
}

func cat()  {
	conf := rclient.InitDefaultConf()
	ctx := context.Background()
	appClient := rclient.NewRenter(ctx, conf)
	ok := appClient.CheckCollateral()
	if !ok {
		fmt.Println("Checking collateral failed")
		return
	}
	cid := os.Args[3]
	err := CliCatFile()
	if err != nil {
		appClient.RscDecodingReq(cid)
	}
}

func main()  {
	add()
}