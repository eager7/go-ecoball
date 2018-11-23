package main

import (
	"os"
	"fmt"
	"os/exec"
	"bytes"
)

func shell_exec(str string)(string, error) {
	cmd := exec.Command("/bin/bash","-c", str)
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	return out.String(), err
}

func main(){
	args := os.Args
	if len(args) < 2 || args == nil{
		fmt.Printf("args error")
		return
	}

	//clang
	clang := "./../tools/clang "
	clang += args[1]
	clang += " -O -c --target=wasm32-unknown-unknown -emit-llvm -nostdinc -nostdlib -D WEBASSEMBLY -I ../include/musl -I ../include/abalib -o file.bc"
	result, err := shell_exec(clang)
    if err != nil{
		fmt.Printf("[clang]:error %s\n",err)
		fmt.Printf("\n check for syntax errors \n")
    	return
	}

	//llvm-link
	link := "./../tools/llvm-link file.bc ../lib/lib.bc -only-needed -o sum.bc"
	result, err = shell_exec(link)
	if err != nil{
		return
	}

	//llc
	llc := "./../tools/llc sum.bc -march=wasm32 -filetype=asm -asm-verbose=false -thread-model=single -data-sections -function-sections -o file.o"
	result, err = shell_exec(llc)
	if err != nil{
		return
	}

	//s2wasm
	s2wasm := "./../tools/s2wasm file.o --allocate-stack 20480 --validate wasm -o file"
	result, err = shell_exec(s2wasm)
	if err != nil{
		fmt.Printf("[s2wasm]:%s\n",result)
		return
	}

	//wasm-opt
	opt := "./../tools/wasm-opt file -o out.wasm"
	result, err = shell_exec(opt)
	if err != nil{
		return
	}

	//wasm-dis
	dis := "./../tools/wasm-dis out.wasm -o out.wast"
	result, err = shell_exec(dis)
	if err != nil{
		return
	}

	//mv
	out := "mv out.wasm out.wast ../build"
	shell_exec(out)

	//clean
	clean := "rm file.bc sum.bc file.o file"
	shell_exec(clean)

	return
}
