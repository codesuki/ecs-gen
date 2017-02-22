package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"fmt"
)

type ecsClient struct {
	ecs *ecs.ECS
}

func newECS(region string, sess *session.Session) *ecsClient {
	return &ecsClient{ecs.New(sess, aws.NewConfig().WithRegion(region))}
}

func (e *ecsClient) describeCluster(cluster string) (*ecs.Cluster, error) {
	params := &ecs.DescribeClustersInput{
		Clusters: []*string{
			aws.String(cluster),
		},
	}
	resp, err := e.ecs.DescribeClusters(params)
	if err != nil {
		return nil, err
	}
	if len(resp.Clusters) == 0 {
		return nil, fmt.Errorf("Cluster %s not found. Did you specify the correct --cluster and --region?", cluster)
	}
	return resp.Clusters[0], nil
}

func (e *ecsClient) listContainerInstances(cluster string) ([]*string, error) {
	params := &ecs.ListContainerInstancesInput{
		Cluster: aws.String(cluster),
	}
	resp, err := e.ecs.ListContainerInstances(params)
	if err != nil {
		return nil, err
	}
	return resp.ContainerInstanceArns, nil
}

func (e *ecsClient) describeContainerInstances(cluster string, arns []*string) ([]*ecs.ContainerInstance, error) {
	params := &ecs.DescribeContainerInstancesInput{
		ContainerInstances: arns,
		Cluster:            aws.String(cluster),
	}
	resp, err := e.ecs.DescribeContainerInstances(params)
	if err != nil {
		return nil, err
	}
	return resp.ContainerInstances, nil
}

func (e *ecsClient) listServices(cluster string) ([]*string, error) {
	params := &ecs.ListServicesInput{
		Cluster: aws.String(cluster),
	}
	resp, err := e.ecs.ListServices(params)
	if err != nil {
		return nil, err
	}
	return resp.ServiceArns, nil
}

func (e *ecsClient) describeServices(cluster string, arns []*string) ([]*ecs.Service, error) {
	params := &ecs.DescribeServicesInput{
		Services: arns,
		Cluster:  aws.String(cluster),
	}
	resp, err := e.ecs.DescribeServices(params)
	if err != nil {
		return nil, err
	}
	return resp.Services, nil
}

func (e *ecsClient) listTasks(cluster string) ([]*string, error) {
	params := &ecs.ListTasksInput{
		Cluster: aws.String(cluster),
		//	ContainerInstance: aws.String("String"),
		//	DesiredStatus:     aws.String("DesiredStatus"),
		//		Family:            aws.String("String"),
		//		MaxResults:        aws.Int64(1),
		//		NextToken:         aws.String("String"),
		//		ServiceName:       aws.String("String"),
		//		StartedBy:         aws.String("String"),
	}
	resp, err := e.ecs.ListTasks(params)
	if err != nil {
		return nil, err
	}
	return resp.TaskArns, nil
}

func (e *ecsClient) describeTasks(cluster string, arns []*string) ([]*ecs.Task, error) {
	params := &ecs.DescribeTasksInput{
		Cluster: aws.String(cluster),
		Tasks:   arns,
	}
	tasks, err := e.ecs.DescribeTasks(params)
	if err != nil {
		return nil, err
	}
	return tasks.Tasks, nil
}

func (e *ecsClient) describeTaskDefinition(arn *string) (*ecs.TaskDefinition, error) {
	params := &ecs.DescribeTaskDefinitionInput{
		TaskDefinition: arn,
	}
	resp, err := e.ecs.DescribeTaskDefinition(params)
	if err != nil {
		return nil, err
	}
	return resp.TaskDefinition, nil
}
