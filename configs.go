package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

func LoadDeployments(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Could not find config file:", path)
		os.Exit(1)
	}
	err = json.Unmarshal(data, Deployment)
	if err != nil {
		log.Println("Could not read/parse the config file ...", err, path)
		os.Exit(1)
	}
}

func LoadServers(path string) {
	if strings.Contains(path, "server.json") {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Println("Could not find server config file:", path)
			os.Exit(1)
		}
		S := new(Server)
		err = json.Unmarshal(data, S)
		if err != nil {
			log.Println("Could not read/parse the config file ...", err, path)
			os.Exit(1)
		}
		Servers = append(Servers, S)
		return
	}

	configs := FindFiles(path, "server", ".json")
	for i, v := range configs {
		data, err := os.ReadFile(v)
		if err != nil {
			log.Println("Could not find server config file:", i)
			os.Exit(1)
		}
		S := new(Server)
		err = json.Unmarshal(data, S)
		if err != nil {
			log.Println("Could not read/parse the config file ...", err, v)
			os.Exit(1)
		}
		Servers = append(Servers, S)
	}
}

func LoadServices(path string) {
	configs := FindFiles(path, "script", ".json")
	var Services []Script
	for i, v := range configs {
		data, err := os.ReadFile(v)
		if err != nil {
			log.Println("Could not find service config file:", v, i)
			os.Exit(1)
		}
		S := new(Script)
		err = json.Unmarshal(data, S)
		if err != nil {
			log.Println("Could not read/parse the config file ...", err, v)
			os.Exit(1)
		}
		Services = append(Services, *S)
	}

	for i := range Servers {
		newScripts := make([]Script, len(Services))
		copy(newScripts, Services)
		Servers[i].Scripts = newScripts
		// fmt.Println("Loaded", len(Servers[i].Scripts), "scripts for server", Servers[i].Hostname)
	}
}

func LoadVariables(path string) {
	if path == "" {
		return
	}
	// configs := FindFiles(path, tag+".variables", ".json")
	// for _, v := range configs {
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Could not find variables config file:", path)
		os.Exit(1)
	}
	err = json.Unmarshal(data, &Variables)
	if err != nil {
		log.Println("Could not read/parse the config file ...", err, path)
		os.Exit(1)
	}
	// }
}

func LoadTemplates(path string) {
	for i, v := range Servers {
		for ii, iv := range v.Scripts {
			for iii, iiv := range iv.CMD {
				if iiv.Template != nil {
					data, err := os.ReadFile(iiv.Template.Local)
					if err != nil {
						log.Println("Could not find config template file:", iiv.Template.Local)
						os.Exit(1)
					}
					Servers[i].Scripts[ii].CMD[iii].Template.Data = make([]byte, len(data))
					Servers[i].Scripts[ii].CMD[iii].Template.Data = data
				}
			}
		}
	}
}
