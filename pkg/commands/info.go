package commands

import (
	"fmt"
	"os"
	"runtime"

	"github.com/bugsnag/osext"
	"github.com/moul/advanced-ssh-config/vendor/github.com/codegangsta/cli"

	"github.com/moul/advanced-ssh-config/pkg/config"
	// . "github.com/moul/advanced-ssh-config/pkg/logger"
)

func cmdInfo(c *cli.Context) {
	conf, err := config.Open()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Debug mode (client): %v\n", os.Getenv("ASSH_DEBUG") == "1")
	cliPath, _ := osext.Executable()
	fmt.Printf("CLI Path: %s\n", cliPath)
	fmt.Printf("Go version: %s\n", runtime.Version())
	fmt.Printf("OS/Arch: %s/%s\n", runtime.GOOS, runtime.GOARCH)
	fmt.Println("")
	fmt.Printf("RC files:\n")
	for _, filename := range conf.IncludedFiles() {
		fmt.Printf("- %s\n", filename)
	}
	fmt.Println("")
	fmt.Println("Statistics:")
	fmt.Printf("- %d hosts\n", len(conf.Hosts))
	fmt.Printf("- %d templates\n", len(conf.Templates))
	fmt.Printf("- %d included files\n", len(conf.IncludedFiles()))
	// FIXME: print info about connections/running processes
	// FIXME: print info about current config file version
}
