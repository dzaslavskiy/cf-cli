//go:build go1.13
// +build go1.13

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"code.cloudfoundry.org/cli/cf/cmd"
	"code.cloudfoundry.org/cli/command/common"
	"code.cloudfoundry.org/cli/util/command_parser"
	"code.cloudfoundry.org/cli/util/configv3"
	"code.cloudfoundry.org/cli/util/panichandler"
	plugin_util "code.cloudfoundry.org/cli/util/plugin"
	"code.cloudfoundry.org/cli/util/ui"
	"github.com/jessevdk/go-flags"
)

func main() {

	args, err := flags.Parse(&opts)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}

	if opts.Zone != "" {
		fmt.Printf("Lets set the zone\n")
		err := os.Setenv("CF_HOME", fmt.Sprintf("%s/.cfm/%s",os.ExpandEnv("$HOME"), opts.Zone))


		if err != nil {
			fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
			os.Exit(1)
		}


		zcmd := exec.Command("cp", "-rv", fmt.Sprintf("%s/.cf", os.ExpandEnv("$HOME")) , fmt.Sprintf("%s/", os.ExpandEnv("$CF_HOME")))
		zcmd.Env = os.Environ()

		
		fmt.Printf("cmd env var: %v\n", zcmd.Env)
		//fmt.Printf("env var: %s\n", os.Getenv("CF_HOME"))

		fmt.Printf("command: %s\n", zcmd.String())

		time.Sleep(1 * time.Second)

		stdout, err := zcmd.CombinedOutput()

		fmt.Printf("Zone: %s\n", stdout)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		}


		zcmd2 := exec.Command("cf", "target", "-s", opts.Zone)
		zcmd2.Env = os.Environ()
		stdout2, err := zcmd2.CombinedOutput()

		if err != nil {
			fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		}

		fmt.Printf("Zone: %s\n", stdout2)

	}


	fmt.Printf("Zone: %s\n", opts.Zone)
	fmt.Printf("Args: %s\n", args)
	fmt.Printf("END HACK\n\n")
	

	var exitCode int
	defer panichandler.HandlePanic()

	config, err := configv3.GetCFConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}

	commandUI, err := ui.NewUI(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}

	p, err := command_parser.NewCommandParser()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unexpected error: %s\n", err.Error())
		os.Exit(1)
	}

	exitCode, err = p.ParseCommandFromArgs(commandUI, os.Args[1:])
	if err == nil {
		os.Exit(exitCode)
	}

	if unknownCommandError, ok := err.(command_parser.UnknownCommandError); ok {
		plugin, commandIsPlugin := plugin_util.IsPluginCommand(os.Args[1:])

		switch {
		case commandIsPlugin:
			err = plugin_util.RunPlugin(plugin)
			if err != nil {
				exitCode = 1
			}

		case common.ShouldFallbackToLegacy:
			cmd.Main(os.Getenv("CF_TRACE"), os.Args)
			//NOT REACHED, legacy main will exit the process

		default:
			unknownCommandError.Suggest(plugin_util.PluginCommandNames())
			fmt.Fprintf(os.Stderr, "%s\n", unknownCommandError.Error())
			os.Exit(1)
		}
	}

	os.Exit(exitCode)
}


var opts struct {
	Zone string `short:"z" long:"zone" description:"A zone"`
}