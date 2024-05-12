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

var projectFlag = flagConfig{
	Flag:      "project",
	Shorthand: "p",
	Usage:     "The path to your project files (not needed if using a deployment file)",
}

var deploymentFlag = flagConfig{
	Flag:      "deployment",
	Shorthand: "d",
	Usage:     "The path to your deployment file",
}

var serversFlag = flagConfig{
	Flag:      "servers",
	Shorthand: "s",
	Usage:     "The path to your server files (not needed if using a deployment file)",
}

var varsFlag = flagConfig{
	Flag:      "variables",
	Shorthand: "v",
	Usage:     "The path to your variables file (not needed if using a deployment file)",
}

var filterFlag = flagConfig{
	Flag:      "filter",
	Shorthand: "f",
	Usage:     "Only scripts or commands with this tag will be executed. Example: SCRIPT.CMD ",
}

var ignorePrompt bool

func init() {
	flag.StringVar(&projectFlag.Value, projectFlag.Flag, "", projectFlag.Usage)
	flag.StringVar(&projectFlag.Value, projectFlag.Shorthand, "", projectFlag.Usage+" (shorthand)")
	flag.StringVar(&deploymentFlag.Value, deploymentFlag.Flag, "", deploymentFlag.Usage)
	flag.StringVar(&deploymentFlag.Value, deploymentFlag.Shorthand, "", deploymentFlag.Usage+" (shorthand)")
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
	if deploymentFlag.Value != "" {
		LoadDeployments(deploymentFlag.Value)
	}

	if serversFlag.Value != "" {
		Deployment.Servers = serversFlag.Value
	}
	if varsFlag.Value != "" {
		Deployment.Vars = varsFlag.Value
	}
	if projectFlag.Value != "" {
		Deployment.Project = projectFlag.Value
	}
	if Deployment.Servers == "" || Deployment.Project == "" {
		color.Red("One or flags are missing, please verify your command line arguments..")
		fmt.Println()
		fmt.Println("-servers=" + Deployment.Servers)
		fmt.Println("-project=" + Deployment.Project)
		fmt.Println()
		os.Exit(1)
	}
	LoadServers(Deployment.Servers)
	LoadServices(Deployment.Project)
	LoadVariables(Deployment.Vars)
	LoadTemplates(Deployment.Project)
	InjectVariables()

	if !ignorePrompt {
		fmt.Println()
		color.Green("You are about to run this deployment")
		fmt.Println()
		fmt.Println(color.GreenString("Servers:"), Deployment.Servers)
		fmt.Println(color.GreenString("Project"), Deployment.Project)
		fmt.Println(color.GreenString("Variables File:"), Deployment.Vars)
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
		fmt.Println(color.GreenString("Servers:"), Deployment.Servers)
		fmt.Println(color.GreenString("Project"), Deployment.Project)
		fmt.Println(color.GreenString("Variables File:"), Deployment.Vars)
		fmt.Println()
	}

	for i := range Servers {
		OpenSessionsAndRunCommands(Servers[i])
	}

	ParseWaitGroup.Wait()
}
