package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

func LoadDeployments(path string) {
	data, err := ioutil.ReadFile(path)
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
	configs := FindFiles(path, "server", ".json")
	for _, v := range configs {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			log.Println("Could not find config file:", v)
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
	for _, v := range configs {
		data, err := ioutil.ReadFile(v)
		if err != nil {
			log.Println("Could not find config file:", v)
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
	}
}
func LoadVariables(path string) {
	// configs := FindFiles(path, tag+".variables", ".json")
	// for _, v := range configs {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Println("Could not find config file:", path)
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
					data, err := ioutil.ReadFile(iiv.Template.Local)
					if err != nil {
						log.Println("Could not find config file:", v)
						os.Exit(1)
					}
					Servers[i].Scripts[ii].CMD[iii].Template.Data = make([]byte, len(data))
					Servers[i].Scripts[ii].CMD[iii].Template.Data = data
				}
			}
		}
	}
}
