package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"gopkg.in/alecthomas/kingpin.v2"
)

type container struct {
	Host    string
	Port    string
	Address string
	Env     map[string]string
}

type metadata struct {
	Cluster              string
	ContainerInstanceArn string
	Version              string
}

type document struct {
	Region           string
	AvailabilityZone string
}


var (
	errClusterNotActive = errors.New("ecs-nginx-proxy: cluster is not active")
)

var (
	app = kingpin.New("ecs-gen", "docker-gen for AWS ECS.")

	region       = app.Flag("region", "AWS region.").Short('r').String()
	cluster      = app.Flag("cluster", "ECS cluster name.").Short('c').String()
	templateFile = app.Flag("template", "Path to template file.").Short('t').Required().ExistingFile()
	outputFile   = app.Flag("output", "Path to output file.").Short('o').Required().String()
	taskName     = app.Flag("task", "Name of ECS task containing nginx.").Default("ecs-nginx-proxy").String()
	hostVar      = app.Flag("host-var", "Which ENV var to use for the hostname.").Default("virtual_host").String()

	signal = app.Flag("signal", "Command to run to signal change.").Short('s').Default("nginx -s reload").String()

	freq = app.Flag("frequency", "Time in seconds between polling. Must be >0.").Short('f').Default("30").Int()
	once = app.Flag("once", "Only execute the template once and exit.").Bool()
)

var version = ""

func main() {
	app.Version(version)
	app.DefaultEnvars()
	kingpin.MustParse(app.Parse(os.Args[1:]))
	if *hostVar == "" {
		log.Fatalf("host-var must not be empty")
	}
	if *cluster == "" {
		var err error
		cluster, err = findClusterName()
		if err != nil || *cluster == "" {
			log.Fatalf("could not determine cluster name. please define using --cluster / -c")
		}
		log.Println("found cluster name to be:", *cluster)
	}

	if *region == "" {
		var err error
		region, err = findRegion()
		if err != nil || *region == "" {
			panic("could not determine region name. please define using --region / -r.")
		}
		log.Println("found Region name to be:", *region)
	}

	sess, err := session.NewSession()
	if err != nil {
		panic(err)
	}
	ec2 := newEC2(*region, sess)
	ecs := newECS(*region, sess)
	if *once {
		updateAndWrite(ec2, ecs)
		return
	}
	execute(ec2, ecs)
	for range time.Tick(time.Second * time.Duration(*freq)) {
		execute(ec2, ecs)
	}
}

func execute(ec2 *ec2Client, ecs *ecsClient) {
	updateAndWrite(ec2, ecs)
	var err error
	err = runSignal()
	if err != nil {
		log.Println("failed to run signal command")
		log.Println("error: ", err)
		switch err := err.(type) {
		case *exec.ExitError:
			log.Println(err.Stderr)
		}
		os.Exit(1)
	}
}

func updateAndWrite(ec2 *ec2Client, ecs *ecsClient) {
	containers, err := newScanner(*cluster, *hostVar, ec2, ecs).scan()
	if err != nil {
		log.Println(err)
	}
	err = writeConfig(containers)
	if err != nil {
		log.Println(err)
	}
}

func runSignal() error {
	log.Println("running signal command")
	output, err := exec.Command("/bin/sh", "-c", *signal).CombinedOutput()
	log.Println("===== output start =====")
	log.Println(string(output))
	log.Println("===== output end =====")
	return err
}

func newTemplate(name string) *template.Template {
	tmpl := template.New(name).Funcs(template.FuncMap{
		"replace": strings.Replace,
		"split":   strings.Split,
		"splitN":  strings.SplitN,
	})
	return tmpl
}

func writeConfig(params []*container) error {
	var containers map[string][]*container
	containers = make(map[string][]*container)
	// Remap the containers as a 2D array with the domain as the index
	for _, v := range params {
		if _, ok := containers[v.Host]; !ok {
			containers[v.Host] = make([]*container, 0)
		}
		containers[v.Host] = append(containers[v.Host], v)
	}
	tmpl, err := newTemplate(filepath.Base(*templateFile)).ParseFiles(*templateFile)
	if err != nil {
		return err
	}
	f, err := os.Create(*outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, containers)
}

func findClusterName() (*string, error) {
	ip, err := findHostIP()
	if err != nil {
		return nil, err
	}
	meta, err := fetchMetadata(ip)
	if err != nil {
		return nil, err
	}
	return &meta.Cluster, nil
}

func findHostIP() (string, error) {
	result, err := sendHTTRequest("http://169.254.169.254/latest/meta-data/local-ipv4")
	return string(result), err
}

func fetchMetadata(host string) (*metadata, error) {
	result, err := sendHTTRequest("http://" + host + ":51678/v1/metadata")
	var meta metadata
	err = json.Unmarshal(result, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil
}

func findRegion() (*string, error) {
	meta, err := fetchDocument()
	if err != nil {
		return nil, err
	}
	return &meta.Region, nil
}

func fetchDocument()(*document, error) {
	result, err := sendHTTRequest("http://169.254.169.254/latest/dynamic/instance-identity/document")
	var meta document
	err = json.Unmarshal(result, &meta)
	if err != nil {
		return nil, err
	}
	return &meta, nil

}

func sendHTTRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return body, nil
}
