package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"
)

// Conf object holds the server configuration
var Conf *Configuration

// Configuration dictates the schema of the yml conf file
type Configuration struct {
	Server struct {
		Listen string
		Dirs   struct {
			Templates string
			Static    string
		}
	}
}

// LoadConfig loads the configuration file in memory
func LoadConfig(filename string) *Configuration {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Println(err)
		return nil
	}
	conf := new(Configuration)
	err = yaml.Unmarshal(b, conf)
	if err != nil {
		log.Println(err)
		return nil
	}
	return conf
}

// Server entry point
func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Please provide conf file\n")
		os.Exit(1)
	}

	Conf = LoadConfig(os.Args[1])
	if Conf == nil {
		os.Exit(1)
	}
	fs := http.FileServer(http.Dir(Conf.Server.Dirs.Static))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/", uiHandler)
	http.ListenAndServe(Conf.Server.Listen, nil)
}
