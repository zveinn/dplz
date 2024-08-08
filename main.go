package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/opensourcez/logger"
)

// Defines a group of user-defined information through a flag.
type flagConfig struct {
	// Defines the Flag: example -project
	Flag string

	// Defines Flag Shorthand: example -p
	Shorthand string

	// A small text explaining the effect of given flag on the application
	Usage string

	// User input from given flag, should be set by flag.StringVars
	Value string
}

var scriptFlag = flagConfig{
	Flag:      "script",
	Shorthand: "sc",
	Usage:     "The path to your script file",
}

var serversFlag = flagConfig{
	Flag:      "servers",
	Shorthand: "s",
	Usage:     "The path to your server file",
}

var varsFlag = flagConfig{
	Flag:      "variables",
	Shorthand: "v",
	Usage:     "The path to your variables file",
}

var filterFlag = flagConfig{
	Flag:      "filter",
	Shorthand: "f",
	Usage:     "Only scripts or commands with this tag will be executed. Example: SCRIPT.CMD ",
}

var ignorePrompt bool

func init() {
	flag.StringVar(&scriptFlag.Value, scriptFlag.Flag, "", scriptFlag.Usage)
	flag.StringVar(&scriptFlag.Value, scriptFlag.Shorthand, "", scriptFlag.Usage+" (shorthand)")
	flag.StringVar(&serversFlag.Value, serversFlag.Flag, "", serversFlag.Usage)
	flag.StringVar(&serversFlag.Value, serversFlag.Shorthand, "", serversFlag.Usage+" (shorthand)")
	flag.StringVar(&varsFlag.Value, varsFlag.Flag, "", varsFlag.Usage)
	flag.StringVar(&varsFlag.Value, varsFlag.Shorthand, "", varsFlag.Usage+" (shorthand)")
	flag.StringVar(&filterFlag.Value, filterFlag.Flag, "", filterFlag.Usage)
	flag.StringVar(&filterFlag.Value, filterFlag.Shorthand, "", filterFlag.Usage+" (shorthand)")

	flag.BoolVar(&ignorePrompt, "ignorePrompt", false, "Add this flag to skipt the confirmation prompt")
}

func main() {
	flag.Parse()

	err, _ := logger.Init(&logger.LoggingConfig{
		DefaultLogTag:   "testing-logs",
		DefaultLogLevel: "INFO",
		WithTrace:       true,
		SimpleTrace:     true,
		PrettyPrint:     true,
		Colors:          true,
		Type:            "stdout",
	})
	if err != nil {
		panic(err)
	}
	if filterFlag.Value != "" {
		splitFilter := strings.Split(filterFlag.Value, ".")
		if len(splitFilter) < 2 {
			log.Println("You need to specify at least two filters seperated by a dot, example:  script.* or script.files etc..")
			os.Exit(1)
		}
		ScriptFilter = splitFilter[0]
		CMDFilter = splitFilter[1]
	}
	Deployment = new(D)

	if serversFlag.Value != "" {
		Deployment.Server = serversFlag.Value
	}
	if varsFlag.Value != "" {
		Deployment.Vars = varsFlag.Value
	}
	if scriptFlag.Value != "" {
		Deployment.Script = scriptFlag.Value
	}
	if Deployment.Server == "" || Deployment.Script == "" {
		color.Red("One or flags are missing, please verify your command line arguments..")
		fmt.Println()
		fmt.Println("-s=" + Deployment.Server)
		fmt.Println("-sc=" + Deployment.Script)
		fmt.Println("-v=" + Deployment.Vars)
		fmt.Println()
		os.Exit(1)
	}

	LoadServers(Deployment.Server)
	LoadScript(Deployment.Script)
	LoadVariables(Deployment.Vars)
	LoadTemplates(Deployment.Script)
	InjectVariables()

	if !ignorePrompt {
		fmt.Println()
		color.Green("You are about to run this deployment")
		fmt.Println()
		fmt.Println(color.GreenString("Server:"), Deployment.Server)
		fmt.Println(color.GreenString("Script:"), Deployment.Script)
		fmt.Println(color.GreenString("Variables:"), Deployment.Vars)
		fmt.Println()
		fmt.Println("Press N to cancel ..")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() == "N" || input.Text() == "n" {
			os.Exit(1)
		}
	} else {
		fmt.Println()
		color.Green("Running deployment")
		fmt.Println()
		fmt.Println(color.GreenString("Server:"), Deployment.Server)
		fmt.Println(color.GreenString("Script:"), Deployment.Script)
		fmt.Println(color.GreenString("Variables:"), Deployment.Vars)
		fmt.Println()
	}

	for i := range Servers {
		OpenSessionsAndRunCommands(Servers[i])
	}

	ParseWaitGroup.Wait()
}
