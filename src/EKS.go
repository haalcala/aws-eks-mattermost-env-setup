package main

import (
	"fmt"

	aws_util "../aws"
	"github.com/aws/aws-sdk-go/service/eks"
)

// bla bla bla
func (m *MMDeployContext) WaitIfClusterCreating() error {
	fmt.Println("------ func (m *MMDeployContext) WaitIfClusterCreating() error")

	aws_util.WaitUntilTrue(func() bool {
		cluster, err := m.GetEKSCluster()

		if err != nil {
			return true
		}

		m.EKSCluster = cluster

		return *m.EKSCluster.Status == aws_util.EKS_STATUS_CREATING
	})

	return nil
}

// bla bla bla
func (m *MMDeployContext) GetEKSCluster() (*eks.Cluster, error) {
	fmt.Println("------ func (m *MMDeployContext) GetEKSCluster() error")

	cluster, err := m.EKS.DescribeCluster(&eks.DescribeClusterInput{
		Name: &m.DeployConfig.ClusterName,
	})

	return cluster.Cluster, err
}
