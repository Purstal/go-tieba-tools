package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

func main() {
	if len(os.Args) != 3 {
		panic("用法不对.")
	}

	b, err := ioutil.ReadFile("thread-list")
	if err != nil {
		panic(err)
	}

	lines := strings.Split(string(b), "\n")

	rx := regexp.MustCompile(`(\d+)`)

	for _, line := range lines {
		tid := rx.FindString(line)
		if tid != "" {
			cmd := exec.Command("post-scaner", tid, os.Args[1], os.Args[2])
			//cmd.Run()
			out, _ := cmd.Output()
			fmt.Println(string(out))
		}

	}

}
