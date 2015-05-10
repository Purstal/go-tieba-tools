package main

import (
	//"path/filepath"
	"fmt"
	"io/ioutil"
	"os/exec"
)

func main() {
	dir, err := ioutil.ReadDir("scanned-thread/")

	if err != nil {
		panic(err)
	}

	for _, fi := range dir {
		fileName := fi.Name()
		cmd := exec.Command("post-analyser", "scanned-thread/"+fileName)
		out, _ := cmd.Output()
		fmt.Println(string(out))
	}
}
