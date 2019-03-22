package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/threefoldtech/zbus/generation"
)

func main() {
	options := generation.NewOptions()
	args := flag.Args()
	if len(args) != 2 {
		log.Println("invalid call to zbus missing fqn")
		log.Println("Usage: zbus [flags] <fqn> <output-file>")
		log.Println("	fdn = path-to-package+InterfaceName")
		log.Println("	example: github.com/me/server+MyApi")
		os.Exit(1)
	}
	fqn := args[0]
	output := args[1]

	generator, err := generation.Generator(fqn)
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := ioutil.TempFile("", "zbus_*.go")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmp.Name())
	generator.Render(tmp)
	tmp.Close()

	cmd := exec.Command(
		"go", "run", tmp.Name(),
		"-module", options.Module,
		"-name", options.Name,
		"-version", options.Version,
		"-package", options.Package,
	)

	outputFile, err := os.Create(output)
	if err != nil {
		log.Fatal(err)
	}

	cmd.Stdout = outputFile
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
