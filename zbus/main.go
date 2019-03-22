package main

import (
	"flag"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	"github.com/threefoldtech/zbus/generation"
)

func run(options generation.Options, fqn, output string) error {
	generator, err := generation.Generator(fqn)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempFile("", "zbus_*.go")
	if err != nil {
		return err
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
	tmpOutput, err := ioutil.TempFile("", "zbus_gen_*.go")
	if err != nil {
		return err
	}

	defer os.Remove(tmpOutput.Name())

	cmd.Stdout = tmpOutput
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		os.Remove(tmpOutput.Name())
		return err
	}

	tmpOutput.Close()

	return move(tmpOutput.Name(), output)
}

//move make sure to move even if temp is on a separate mount, this is why we can't use os.Rename directly
func move(src, dst string) error {
	dstF, err := os.Create(dst)
	if err != nil {
		return err
	}

	srcF, err := os.Open(src)
	if err != nil {
		return err
	}

	_, err = io.Copy(dstF, srcF)
	return err
}

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

	if err := run(options, fqn, output); err != nil {
		log.Fatal(err)
	}
}
