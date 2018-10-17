package main
import (

	rtypes "github.com/ecoball/go-ecoball/dsn/renter"
	"io/ioutil"
	"fmt"
	"net/http"
	"net/url"
	"encoding/json"
	"log"
	"bytes"
//  "strings"
	"os"
	shell "github.com/ecoball/go-ecoball/dsn/renter/client/web"
	"github.com/ecoball/go-ecoball/dsn/common"
	chunker "gx/ipfs/QmVDjhUMtkRskBFAVNwyXuLSKbeAya7JKPnzAxMKDaK4x4/go-ipfs-chunker"
)
func main() {

	addFile("E:\\临时\\blue.jpg", "localhost:5011")

}

func addFile(path string, shUrl string){
 
	sh := shell.NewShell("localhost:5011")
	file, err := os.Open(path)
	if err!= nil{
		fmt.Println("file not open")
		os.Exit(1)
	}
	defer file.Close()
	cid, err := sh.Add(file)
	if err != nil {
	fmt.Fprintf(os.Stderr, "error: %s", err)
	os.Exit(1)
	}
	fmt.Println("added %s", cid)
	stat, err := os.Lstat(path)
	if err != nil {
		return 
	}
	var PieceSize uint64
	if stat.Size() < common.EraDataPiece * chunker.DefaultBlockSize {
		PieceSize = uint64(stat.Size() / common.EraDataPiece)
	} else {
		PieceSize = uint64(chunker.DefaultBlockSize)
	}
	rscReq := rtypes.RscReq{ Cid: cid, Redundency: 2, IsDir: false, Chunk: PieceSize, FileSize: uint64(stat.Size()) }
	httpPosteraCoding(rscReq)
}

func httpPosteraCoding(rscReq rtypes.RscReq) {

	b ,err := json.Marshal(rscReq)
	if err != nil {
	log.Println("json format error:", err)
		return
	}
	body := bytes.NewBuffer(b)
	resp, err := http.Post("http://localhost:9000/dsn/eracode",
			"application/json;charset=utf-8",
			body)
	
	if err != nil {
	fmt.Println(err)
	}
		if err != nil {
	log.Println("Post failed:", err)
		return
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body)
	if err != nil {
	log.Println("Read failed:", err)
		return
	}
	log.Println("content:", string(content))
}


//test
func httpPostForm() {
resp, err := http.PostForm("http://localhost:8086/dsn/eracode",
url.Values{"key": {"Value"}, "id": {"123"}})
if err != nil {
// handle error
}
defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
if err != nil {
// handle error
}
fmt.Println(string(body))
}
func httpGet(name string) {
resp, err := http.Get("http://localhost:8086/")
if err != nil {
// handle error
}
defer resp.Body.Close()
body, err := ioutil.ReadAll(resp.Body)
if err != nil {
// handle error
}
fmt.Println(string(body))
}