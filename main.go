package main

import (
	"errors"
	"html/template"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"gopkg.in/alecthomas/kingpin.v2"
)

type container struct {
	Host    string
	Port    string
	Address string
}

var (
	errClusterNotActive = errors.New("ecs-nginx-proxy: cluster is not active")
)

var (
	app = kingpin.New("ecs-gen", "docker-gen for AWS ECS.")

	region       = app.Flag("region", "AWS region.").Short('r').Default("ap-northeast-1").String()
	cluster      = app.Flag("cluster", "ECS cluster name.").Short('c').Required().String()
	templateFile = app.Flag("template", "Path to template file.").Short('t').Required().ExistingFile()
	outputFile   = app.Flag("output", "Path to output file.").Short('o').Required().String()
	taskName     = app.Flag("task", "Name of ECS task containing nginx.").Default("ecs-nginx-proxy").String()

	signal = app.Flag("signal", "Command to run to signal change.").Short('s').Default("nginx -s reload").String()

	freq = app.Flag("frequency", "Time in seconds between polling. Must be >0.").Short('f').Default("30").Int()
	once = app.Flag("once", "Only execute the template once and exit.").Bool()
)

var version = ""

func main() {
	app.Version(version)
	app.DefaultEnvars()
	kingpin.MustParse(app.Parse(os.Args[1:]))
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
	for range time.Tick(time.Second * time.Duration(*freq)) {
		updateAndWrite(ec2, ecs)
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
}

func updateAndWrite(ec2 *ec2Client, ecs *ecsClient) {
	containers, err := newScanner(*cluster, ec2, ecs).scan()
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

func writeConfig(params []*container) error {
	tmpl, err := template.ParseFiles(*templateFile)
	if err != nil {
		return err
	}
	f, err := os.Create(*outputFile)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, params)
}
