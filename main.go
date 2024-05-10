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
	flag string

	// Defines Flag shorthand: example -p
	shorthand string

	// A small text explaining the effect of given flag on the application
	usage string

	// User input from given flag, should be set by flag.StringVars
	content string
}

var projectFlag = flagConfig{
	flag:      "project",
	shorthand: "p",
	usage:     "The path to your project files (not needed if using a deployment file)",
}

var deploymentFlag = flagConfig{
	flag:      "deployment",
	shorthand: "d",
	usage:     "The path to your deployment file",
}

var serversFlag = flagConfig{
	flag:      "servers",
	shorthand: "s",
	usage:     "The path to your server files (not needed if using a deployment file)",
}

var varsFlag = flagConfig{
	flag:      "variables",
	shorthand: "v",
	usage:     "The path to your variables file (not needed if using a deployment file)",
}

var filterFlag = flagConfig{
	flag:      "filter",
	shorthand: "f",
	usage:     "Only scripts or commands with this tag will be executed. Example: SCRIPT.CMD ",
}

var ignorePrompt bool

func init() {
	flag.StringVar(&projectFlag.content, projectFlag.flag, "", projectFlag.usage)
	flag.StringVar(&projectFlag.content, projectFlag.shorthand, "", projectFlag.usage+" (shorthand)")
	flag.StringVar(&deploymentFlag.content, projectFlag.flag, "", deploymentFlag.usage)
	flag.StringVar(&deploymentFlag.content, projectFlag.shorthand, "", deploymentFlag.usage+" (shorthand)")
	flag.StringVar(&serversFlag.content, serversFlag.flag, "", serversFlag.usage)
	flag.StringVar(&serversFlag.content, serversFlag.shorthand, "", serversFlag.usage+" (shorthand)")
	flag.StringVar(&varsFlag.content, varsFlag.flag, "", varsFlag.usage)
	flag.StringVar(&varsFlag.content, varsFlag.shorthand, "", varsFlag.usage+" (shorthand)")
	flag.StringVar(&filterFlag.content, filterFlag.flag, "", filterFlag.usage)
	flag.StringVar(&filterFlag.content, filterFlag.flag, "", filterFlag.usage+" (shorthand)")

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
	if filterFlag.content != "" {
		splitFilter := strings.Split(filterFlag.content, ".")
		if len(splitFilter) < 2 {
			log.Println("You need to specify at least two filters seperated by a dot, example:  script.* or script.files etc..")
			os.Exit(1)
		}
		ScriptFilter = splitFilter[0]
		CMDFilter = splitFilter[1]
	}
	Deployment = new(D)
	if deploymentFlag.content != "" {
		LoadDeployments(deploymentFlag.content)
	}

	if serversFlag.content != "" {
		Deployment.Servers = serversFlag.content
	}
	if varsFlag.content != "" {
		Deployment.Vars = varsFlag.content
	}
	if projectFlag.content != "" {
		Deployment.Project = projectFlag.content
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
