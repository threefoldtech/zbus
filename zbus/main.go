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
	fqn := flag.Args()
	if len(fqn) != 1 {
		log.Println("invalid call to zbus missing fqn")
		log.Println("Usage: zbus [flags] <fqn>")
		log.Println("	fdn = path-to-package+InterfaceName")
		log.Println("	example: github.com/me/server+MyApi")
	}

	generator, err := generation.Generator(fqn[0])
	if err != nil {
		log.Fatal(err)
	}

	tmp, err := ioutil.TempFile("", "zbus_*.go")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmp.Name())
	generator.Render(tmp)
	tmp.Close()

	cmd := exec.Command(
		"go", "run", tmp.Name(),
		"-module", options.Module,
		"-name", options.Name,
		"-version", options.Version,
		"-package", options.Package,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
