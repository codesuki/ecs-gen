package main

import (
	"encoding/json"
	"errors"
	"html/template"
	"io/ioutil"
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

	freq     = app.Flag("frequency", "Time in seconds between polling. Must be >0.").Short('f').Default("30").Int()
	once     = app.Flag("once", "Only execute the template once and exit.").Bool()
	logLevel = app.Flag("log-level", "Set the logging level (info, warn, error)").Short('l').Default(levelWarn).Enum(levelInfo, levelWarn, levelError)
)

var version = ""
var logger = newLogger(levelWarn)

func main() {
	app.Version(version)
	app.DefaultEnvars()
	kingpin.MustParse(app.Parse(os.Args[1:]))
	logger = newLogger(*logLevel)
	sess, err := session.NewSession()
	if err != nil {
		logger.Error(err)
		logger.Fatal("unable to instantiate AWS session")
	}
	meta := NewEC2Metadata(sess)
	checkRegionFlag(meta)
	checkHostVarFlag()
	checkClusterFlag()
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

func checkHostVarFlag() {
	if *hostVar == "" {
		logger.Fatal("host-var must not be empty")
	}
}

func checkClusterFlag() {
	if *cluster == "" {
		var err error
		cluster, err = findClusterName()
		if err != nil || *cluster == "" {
			logger.Fatal("could not determine cluster name. please define using --cluster / -c")
		}
		logger.Info("found cluster name to be:", *cluster)
	}
}

func checkRegionFlag(meta *ec2Meta) {
	if *region == "" {
		r, err := meta.region()
		if err != nil {
			logger.Fatal("could not determine cluster region. please define using --region / -r")
		}
		*region = r
		logger.Info("found cluster region to be:", *region)
	}
}

func execute(ec2 *ec2Client, ecs *ecsClient) {
	updateAndWrite(ec2, ecs)
	var err error
	err = runSignal()
	if err != nil {
		logger.Warning("failed to run signal command")
		logger.Error(err)
		switch err := err.(type) {
		case *exec.ExitError:
			logger.Error(err)
		}
		os.Exit(1)
	}
}

func updateAndWrite(ec2 *ec2Client, ecs *ecsClient) {
	containers, err := newScanner(*cluster, *hostVar, ec2, ecs).scan()
	if err != nil {
		logger.Error(err)
	}
	err = writeConfig(containers)
	if err != nil {
		logger.Error(err)
	}
}

func runSignal() error {
	logger.Info("running signal command")
	output, err := exec.Command("/bin/sh", "-c", *signal).CombinedOutput()
	logger.Info("===== output start =====")
	logger.Info(string(output))
	logger.Info("===== output end =====")
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
