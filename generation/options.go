package generation

import (
	"flag"
	"log"
	"os"
)

// Options defines a stub optiosn (ObjectID)
type Options struct {
	Module  string
	Name    string
	Version string
	Package string
}

// NewOptions creates a new options from command line arguments.
// It terminates if required options are not provided
func NewOptions() Options {
	opt := Options{}
	flag.StringVar(&opt.Module, "module", "", "module name as registered by the zbus server")
	flag.StringVar(&opt.Name, "name", "", "object name as registered by the zbus server")
	flag.StringVar(&opt.Version, "version", "", "object version as registered bt the zbus server")
	flag.StringVar(&opt.Package, "package", "", "package of generated stub")
	
	var help bool
	flag.BoolVar(&help, "help", false, "print this usage")
	flag.Parse()
	if help {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if opt.Module == "" {
		log.Fatalf("module is required")
	}

	if opt.Name == "" {
		log.Fatalf("name is required")
	}
	if opt.Version == "" {
		log.Fatalf("version is required")
	}
	if opt.Package == "" {
		log.Fatalf("package is required")
	}

	return opt
}
