package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// setup subcommands
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	buildInputPath := buildCmd.String("i", "", "specify input file (.yaml)")
	buildOutputDir := buildCmd.String("o", ".", "specify output directory")

	serverCmd := flag.NewFlagSet("serv", flag.ExitOnError)
	serverInputPath := serverCmd.String("i", "", "specify input file (.yaml)")
	serverPort := serverCmd.Int("p", 8000, "specify server port")

	flag.Usage = func() {
		fmt.Println("swagroller: avaiable commands")
		fmt.Println("  ", buildCmd.Name())
		fmt.Println("  ", serverCmd.Name())
		fmt.Print("\n")

		buildCmd.Usage()
		serverCmd.Usage()
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case buildCmd.Name():
		buildCmd.Parse(os.Args[2:])

	case serverCmd.Name():
		serverCmd.Parse(os.Args[2:])

	default:
		flag.Usage()
		os.Exit(1)
	}

	switch {
	case buildCmd.Parsed():
		doBuild(buildCmd.Usage, *buildInputPath, *buildOutputDir)

	case serverCmd.Parsed():
		doServer(serverCmd.Usage, *serverInputPath, *serverPort)
	}
}
