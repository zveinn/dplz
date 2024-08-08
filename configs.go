package main

import (
	"encoding/json"
	"log"
	"os"
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

func LoadScript(path string) {
	var Services []Script
	data, err := os.ReadFile(path)
	if err != nil {
		log.Println("Could not find service config file:", path)
		os.Exit(1)
	}
	S := new(Script)
	err = json.Unmarshal(data, S)
	if err != nil {
		log.Println("Could not read/parse the config file ...", err, path)
		os.Exit(1)
	}
	Services = append(Services, *S)

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
