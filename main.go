package main

import (
	"os"
	"bufio"
	"fmt"
)

func main(){
	bufio.NewReader(os.Stdin)
	fmt.Println("Cli shell")
	for{
		fmt.Println("Please enter an ip or an address:")
	}
}