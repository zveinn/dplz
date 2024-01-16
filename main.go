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

func main() {
	deployment := flag.String("deployment", "", "The path to your deployment file")
	project := flag.String("project", "", "The path to your project files (not needed if using a deployment file)")
	servers := flag.String("servers", "", "The path to your server files (not needed if using a deployment file)")
	vars := flag.String("vars", "", "The path to your variables file (not needed if using a deployment file)")
	ignorePrompt := flag.Bool("ignorePrompt", false, "Add this flag to skipt the confirmation prompt")
	filter := flag.String("filter", "", "Only scripts or commands with this tag will be executed. Example: SCRIPT.CMD ")
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
	if *filter != "" {
		splitFilter := strings.Split(*filter, ".")
		if len(splitFilter) < 2 {
			log.Println("You need to specify at least two filters seperated by a dot, example:  script.* or script.files etc..")
			os.Exit(1)
		}
		ScriptFilter = splitFilter[0]
		CMDFilter = splitFilter[1]
	}
	Deployment = new(D)
	if *deployment != "" {
		LoadDeployments(*deployment)
	}

	if *servers != "" {
		Deployment.Servers = *servers
	}
	if *vars != "" {
		Deployment.Vars = *vars
	}
	if *project != "" {
		Deployment.Project = *project
	}
	if Deployment.Servers == "" || Deployment.Project == "" || Deployment.Vars == "" {
		color.Red("One or flags are missing, please verify your command line arguments..")
		fmt.Println()
		fmt.Println("-servers=" + Deployment.Servers)
		fmt.Println("-project=" + Deployment.Project)
		fmt.Println("-vars=" + Deployment.Vars)
		fmt.Println()
		os.Exit(1)
	}
	LoadServers(Deployment.Servers)
	LoadServices(Deployment.Project)
	LoadVariables(Deployment.Vars)
	LoadTemplates(Deployment.Project)
	InjectVariables()

	if !*ignorePrompt {
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
