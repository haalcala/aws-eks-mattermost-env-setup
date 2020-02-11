package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) EKSListFargateProfiles(clusterName string) (error, EKSListFargateProfilesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks list-fargate-profiles --region %s --cluster-name %s --max-items 99", aws.Region, clusterName))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EKSListFargateProfilesResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EKSCreateFargateProfile(clusterName, fargateProfile, executionRole, selectors, subnets, tags string) (error, EKSCreateFargateProfileResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks create-fargate-profile --region %s --cluster-name %s --fargate-profile-name %s --pod-execution-role-arn %s --selectors %s --subnets %s --tags %s", aws.Region, clusterName, fargateProfile, executionRole, selectors, subnets, tags))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EKSCreateFargateProfileResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EKSDeleteFargateProfile(clusterName, fargateProfile string) (error, EKSListFargateProfilesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks delete-fargate-profile --region %s --cluster-name %s --fargate-profile-name %s", aws.Region, clusterName, fargateProfile))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EKSListFargateProfilesResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EKSDescribeFargateProfile(clusterName, fargateProfile string) (error, EKSDescribeFargateProfileResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks describe-fargate-profile --region %s --cluster-name %s --fargate-profile-name %s", aws.Region, clusterName, fargateProfile))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EKSDescribeFargateProfileResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EKSCreateFargateCluster(clusterName, privateSubnets, publicSubnets string) error {
	err, _, stderr := Execute(fmt.Sprintf("eksctl create cluster --region %s --name %s --fargate --vpc-private-subnets %s --vpc-public-subnets %s", aws.Region, clusterName, privateSubnets, publicSubnets))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}

func (aws *AWS) EKSListClusters() (error, EKSListClustersResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks list-clusters --region %s ", aws.Region))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EKSListClustersResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EKSDescribeCluster(clusterName string) (error, EKSDescribeClusterResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks describe-cluster --region %s --name %s", aws.Region, clusterName))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EKSDescribeClusterResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EKSDeleteCluster(clusterName string) (error, AWSEKSClusterType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws eks delete-cluster --region %s --name %s", aws.Region, clusterName))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp AWSEKSClusterType

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}
