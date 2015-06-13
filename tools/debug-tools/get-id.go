package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/purstal/pbtools/modules/postbar/apis"
)

const usage = `No Usage`

func main() {

	defer func() {
		bufio.NewReader(os.Stdin).ReadLine()
	}()

	if len(os.Args) == 1 {
		fmt.Println(usage)
		return
	}

	if os.Args[1] == "uid#un" {
		if len(os.Args) == 3 {
			uid, err := apis.GetUid(os.Args[2])
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println(uid)
			}
		} else {
			fmt.Println("uid#un un => uid")
			return
		}
	}
}
