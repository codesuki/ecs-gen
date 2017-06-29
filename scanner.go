package main

import (
	"errors"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
)

type scanner struct {
	ec2 *ec2Client
	ecs *ecsClient

	cluster string
	hostVar string

	idAddressMap map[string]string

	nameNetworkBindingsMap map[string][]*ecs.NetworkBinding
}

func newScanner(cluster string, hostVar string, ec2 *ec2Client, ecs *ecsClient) *scanner {
	return &scanner{ec2: ec2, ecs: ecs, cluster: cluster, hostVar: hostVar}
}

func (s *scanner) scan() ([]*container, error) {
	log.Println("getting ECS cluster information")
	clusterInfo, err := s.ecs.describeCluster(s.cluster)
	if err != nil {
		return nil, err
	}
	if *clusterInfo.Status != "ACTIVE" {
		return nil, errClusterNotActive
	}
	s.idAddressMap, err = s.makeIDAddressMap()
	if err != nil {
		return nil, err
	}
	return s.extractContainers()
}

func (s *scanner) makeIDAddressMap() (map[string]string, error) {
	instances := make(map[string]string)
	arns, err := s.ecs.listContainerInstances(s.cluster)
	if err != nil {
		return nil, err
	}
	containerInstances, err := s.ecs.describeContainerInstances(s.cluster, arns)
	if err != nil {
		return nil, err
	}
	for i := range containerInstances {
		instance, err := s.ec2.describeInstance(containerInstances[i].Ec2InstanceId)
		if err != nil {
			return nil, err
		}
		if instance.PrivateIpAddress != nil {
			instances[*containerInstances[i].ContainerInstanceArn] = *instance.PrivateIpAddress
		}
	}
	return instances, nil
}

func (s *scanner) getTasks() ([]*ecs.Task, error) {
	arns, err := s.ecs.listTasks(s.cluster)
	if err != nil {
		return nil, err
	}
	return s.ecs.describeTasks(s.cluster, arns)
}

func (s *scanner) extractContainers() ([]*container, error) {
	tasks, err := s.getTasks()
	if err != nil {
		return nil, err
	}
	containers := make([]*container, 0, 10)
	for _, t := range tasks {
		s.nameNetworkBindingsMap = s.makeNameNetworkBindingsMap(t.Containers)
		taskDefinition, err := s.ecs.describeTaskDefinition(t.TaskDefinitionArn)
		if err != nil {
			return nil, err
		}
		for _, cd := range taskDefinition.ContainerDefinitions {
			container, err := s.extractContainer(t, cd)
			if err != nil {
				log.Println(err)
				continue
			}
			containers = append(containers, container)
		}
	}
	return containers, nil
}

func (s *scanner) makeNameNetworkBindingsMap(containers []*ecs.Container) map[string][]*ecs.NetworkBinding {
	networkBindings := make(map[string][]*ecs.NetworkBinding)
	for _, c := range containers {
		networkBindings[*c.Name] = c.NetworkBindings
	}
	return networkBindings
}

func (s *scanner) extractContainer(t *ecs.Task, cd *ecs.ContainerDefinition) (*container, error) {
	if strings.ToLower(*cd.Name) == *taskName {
		return nil, errors.New("container is own container. skipping")
	}
	if len(s.nameNetworkBindingsMap[*cd.Name]) == 0 {
		return nil, errors.New("container has no network bindings. skipping")
	}
	virtualHost, virtualPort, envVariables := extractVars(cd.Environment, s.hostVar)
	if virtualHost == "" {
		return nil, errors.New("[" + *cd.Name + "] " + s.hostVar + " environment variable not found. skipping")
	}
	port := ""
	if len(s.nameNetworkBindingsMap[*cd.Name]) == 1 {
		port = strconv.FormatInt(*s.nameNetworkBindingsMap[*cd.Name][0].HostPort, 10)
	} else if virtualPort != "" {
		port = extractHostPort(virtualPort, s.nameNetworkBindingsMap[*cd.Name])
	}
	if port == "" {
		return nil, errors.New("[" + *cd.Name + "] no valid port configuration found. skipping")
	}
	return &container{
		Host:    virtualHost,
		Port:    port,
		Env:     envVariables,
		Address: s.idAddressMap[*t.ContainerInstanceArn],
	}, nil
}

func extractHostPort(virtualPort string, nbs []*ecs.NetworkBinding) string {
	for _, nb := range nbs {
		if strconv.FormatInt(*nb.ContainerPort, 10) == virtualPort {
			return strconv.FormatInt(*nb.HostPort, 10)
		}
	}
	return ""
}

func extractVars(env []*ecs.KeyValuePair, hostVar string) (string, string, map[string]string) {
	envVariables := make(map[string]string)
	virtualHost := ""
	virtualPort := ""

	for _, e := range env {
		envVariables[*e.Name] = *e.Value
		if strings.ToLower(*e.Name) == strings.ToLower(hostVar) {
			virtualHost = *e.Value
		} else if strings.ToLower(*e.Name) == "virtual_port" {
			virtualPort = *e.Value
		}
	}
	return virtualHost, virtualPort, envVariables
}
