/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"bms/client/cmd"
	"fmt"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		fmt.Println(err)
	}
}
