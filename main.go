package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"

	"github.com/jessevdk/go-flags"
	"github.com/valyala/fastjson"
)

type Options struct {
	Build string `short:"b" long:"build" description:"Build link" required:"true"`
}

type Build struct {
	Class   string `json:"_class"`
	Actions []struct {
		Class  string `json:"_class"`
		Causes []struct {
			Class            string `json:"_class"`
			ShortDescription string `json:"shortDescription"`
			UpstreamBuild    int64  `json:"upstreamBuild"`
			UpstreamProject  string `json:"upstreamProject"`
			UpstreamURL      string `json:"upstreamUrl"`
		}
		Parameters []struct {
			Class string `json:"_class"`
			Name  string `json:"name"`
			Value string `json:"value"`
		}
	}
	Artifacts []struct {
		DisplayPath  interface{} `json:"displayPath"`
		FileName     string      `json:"fileName"`
		RelativePath string      `json:"relativePath"`
	}
	Building  bool   `json:"building"`
	BuiltOn   string `json:"builtOn"`
	ChangeSet struct {
		Class string        `json:"_class"`
		Items []interface{} `json:"items"`
		Kind  interface{}   `json:"kind"`
	}
	Culprits          []interface{} `json:"culprits"`
	Description       string        `json:"description"`
	DisplayName       string        `json:"displayName"`
	Duration          int64         `json:"duration"`
	EstimatedDuration int64         `json:"estimatedDuration"`
	Executor          interface{}   `json:"executor"`
	FullDisplayName   string        `json:"fullDisplayName"`
	ID                string        `json:"id"`
	KeepLog           bool          `json:"keepLog"`
	Number            int64         `json:"number"`
	QueueID           int64         `json:"queueId"`
	Result            string        `json:"result"`
	Timestamp         int64         `json:"timestamp"`
	URL               string        `json:"url"`
}

const api = "/api/json"

var (
	reURLSegment = regexp.MustCompile(`^(?:[^\/]*\/){5}[^\/]+`)
)

func main() {
	var opts Options

	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.Parse()
	if err != nil {
		os.Exit(1)
	}
	
	// http://jenkinsapi.readthedocs.io/en/latest
	
	url := opts.Build + api
	consuleText := opts.Build + "/consoleText"
	injectedEnvVars := opts.Build + "/injectedEnvVars" + api

	match := reURLSegment.FindString(url)

	res, err := http.Get(url)
	ct, err := http.Get(consuleText)
	envs, err := http.Get(injectedEnvVars)

	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)
	buildlog, err := ioutil.ReadAll(ct.Body)
	vars, err := ioutil.ReadAll(envs.Body)
	bs := string(buildlog)

	if err != nil {
		panic(err.Error())
	}

	data := &Build{}
	json.Unmarshal([]byte(body), data)
	fmt.Println("[Name]", data.FullDisplayName)
	fmt.Println("[Description]", data.Description)
	fmt.Println("[IsBuilding]", data.Building)
	fmt.Println("[Duration]", data.Duration/1000, "sec |", data.Duration/1000/60, "min")
	fmt.Println("[Build Timestamp]", data.Timestamp)
	fmt.Println("[URL]", data.URL)
	fmt.Println("[BuiltOn]", data.BuiltOn)
	fmt.Println("[Result]", data.Result)
	fmt.Println("\n")
	fmt.Println("[Parameters]")
	for _, i := range data.Actions {
		for _, j := range i.Parameters {
			fmt.Println(j)
		}
	}
	fmt.Println("\n")

	var p fastjson.Parser
	j, err := p.ParseBytes(vars)
	if err != nil {
		panic(err.Error())
	}

	fmt.Println("[Environment variables]")
	j.GetObject("envMap").Visit(func(envName []byte, envValue *fastjson.Value) {
		fmt.Printf("%s = %s\n", envName, envValue)
	})
	fmt.Println("\n")

	fmt.Println("[Artefacts]")
	for i := 0; i < len(data.Artifacts); i++ {
		fmt.Print(match, "/artifact/", data.Artifacts[i].RelativePath, "\n")
	}
	fmt.Println("\n")
	fmt.Println("[Console Log]")
	fmt.Print(bs)
	os.Exit(0)
}
