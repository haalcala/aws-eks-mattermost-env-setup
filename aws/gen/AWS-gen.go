//go:generate go run AWS-gen.go

// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
)

var commands = []Command{
	{"EC2", "EC2DeleteVPC", "aws ec2 delete-vpc", "vpcId string", "--vpc-id %s", "vpcId", ""},
	{"IAM", "IAMListRoles", "aws iam list-roles", "", "", "", "IAMListRolesResponse"},
	{"EKS", "EKSListFargateProfiles", "aws eks list-fargate-profiles", "clusterName string", "--cluster-name %s --max-items 99", "clusterName", "EKSListFargateProfilesResponse"},

	{"EKS", "EKSCreateFargateProfile", "aws eks create-fargate-profile", "clusterName, fargateProfile, executionRole, selectors, subnets, tags string", "--cluster-name %s --fargate-profile-name %s --pod-execution-role-arn %s --selectors %s --subnets %s --tags %s", "clusterName, fargateProfile, executionRole, selectors, subnets, tags", "EKSCreateFargateProfileResponse"},
	// selectors : namespace=string,labels={KeyName1=string,KeyName2=string} ...
	// tags : KeyName1=string,KeyName2=string

	{"EKS", "EKSDeleteFargateProfile", "aws eks delete-fargate-profile", "clusterName, fargateProfile string", "--cluster-name %s --fargate-profile-name %s", "clusterName, fargateProfile", "EKSListFargateProfilesResponse"},
	{"EKS", "EKSDescribeFargateProfile", "aws eks describe-fargate-profile", "clusterName, fargateProfile string", "--cluster-name %s --fargate-profile-name %s", "clusterName, fargateProfile", "EKSDescribeFargateProfileResponse"},
	{"EKS", "EKSCreateFargateCluster", "eksctl create cluster", "clusterName, privateSubnets, publicSubnets string", "--name %s --fargate --vpc-private-subnets %s --vpc-public-subnets %s", "clusterName, privateSubnets, publicSubnets", ""},
	{"EKS", "EKSListClusters", "aws eks list-clusters", "", "", "", "EKSListClustersResponse"},
	{"EKS", "EKSDescribeCluster", "aws eks describe-cluster", "clusterName string", "--name %s", "clusterName", "EKSDescribeClusterResponse"},
	{"EKS", "EKSDeleteCluster", "aws eks delete-cluster", "clusterName string", "--name %s", "clusterName", "AWSEKSClusterType"},
	{"CloudFormation", "CFDescribeStacks", "aws cloudformation describe-stacks", "", "", "", "CFDescribeStacksResponse"},
	{"CloudFormation", "CFDeleteStack", "aws cloudformation delete-stack", "stackName string", "--stack-name %s", "stackName", ""},
	{"Subnet", "CreateDefaultSubnet", "aws ec2 create-default-subnet", "availabilityZone string", "--availability-zone %s", "availabilityZone", "EC2CreateSubnetResponse"},
	{"Subnet", "DeleteSubnet", "aws ec2 delete-subnet", "subnetId string", "--subnet-id %s", "subnetId", ""},
	{"Address", "AllocateAddress", "aws ec2 allocate-address", "", "", "", "AWSAddressType"},
	{"Address", "GetAddresses", "aws ec2 describe-addresses", "", "", "", "AWSEC2DescribeAddressesResponse"},
	{"Address", "DisassociateAddress", "aws ec2 disassociate-address", "associationId string", "--association-id %s", "associationId", ""},
	{"Route", "GetRouteTables", "aws ec2 describe-route-tables", "", "", "", "EC2DescribeRouteTablesResponse"},
	{"Route", "CreateRouteTable", "aws ec2 create-route-table", "vpc AWSVPCType", "--vpc-id %s", "vpc.VpcId", "EC2CreateRouteTableResponse"},
	{"Route", "AssociateRouteTable", "aws ec2 associate-route-table", "routeTableId, subnetId string", "--route-table-id %s --subnet-id %s", "routeTableId, subnetId", "EC2AssociateRouteTableResponse"},
	{"Route", "DisassociateRouteTable", "aws ec2 disassociate-route-table", "associationId string", "--association-id %s", "associationId", ""},
	{"Route", "DeleteRouteTable", "aws ec2 delete-route-table", "routeTableId string", "--route-table-id %s", "routeTableId", ""},
	{"Route", "CreateRouteWithInternetGateway", "aws ec2 create-route", "routeTableId, cidr, gatewayId string", "--route-table-id %s --destination-cidr-block %s --gateway-id %s", "routeTableId, cidr, gatewayId", "EC2CreateRouteResponse"},
	{"Route", "CreateRouteWithNatGateway", "aws ec2 create-route", "routeTableId, cidr, gatewayId string", "--route-table-id %s --destination-cidr-block %s --nat-gateway-id %s", "routeTableId, cidr, gatewayId", "EC2CreateRouteResponse"},
	{"NAT", "DeleteNatGateway", "aws ec2 delete-nat-gateway", "natGatewayId string", "--nat-gateway-id %s", "natGatewayId", ""},
	{"RDS", "RDSDescribeDBInstances", "aws rds describe-db-instances", "", "", "", "RDSDescribeDBInstancesResponse"},
	{"RDS", "RDSDeleteDBInstance", "aws rds delete-db-instance", "dbInstanceIdentifier string", "--db-instance-identifier %s --skip-final-snapshot", "dbInstanceIdentifier", "RDSCreateDBInstanceResponse"},
	{"RDS", "RDSCreateDBInstance", "aws rds create-db-instance", "instanceName, dbIdentifier, masterUsername, masterPassword, availabilityZone, subnetGroupName, port, engineVersion, dbInstanceClass, storageSize string", "--db-name %s --db-instance-identifier %s --master-username %s --master-user-password %s --availability-zone %s --db-subnet-group-name %s --port %s --engine-version %s --no-publicly-accessible --engine mysql --db-instance-class %s --allocated-storage %s", "instanceName, dbIdentifier, masterUsername, masterPassword, availabilityZone, subnetGroupName, port, engineVersion, dbInstanceClass, storageSize", "RDSCreateDBInstanceResponse"},
	{"RDS", "RDSDescribeSubnetGroups", "aws rds describe-db-subnet-groups", "", "", "", "RDSDescribeSubnetGroupsResponse"},
	{"RDS", "RDSCreateSubnetGroup", "aws rds create-db-subnet-group", "subnetGroupName, subnetGroupDescription, subnetIds string", "--db-subnet-group-name %s --db-subnet-group-description %s --subnet-ids %s", "subnetGroupName, subnetGroupDescription, subnetIds", "RDSCreateDBSubnetGroupResponse"},
	{"RDS", "RDSDeleteSubnetGroup", "aws rds delete-db-subnet-group", "subnetGroupName string", "--db-subnet-group-name %s", "subnetGroupName", ""}}

type Command struct {
	file, method, command, params, cliParams, cliParamsParams, returnType string
}

func AppendFile(filename, data string) {
	file, err := os.OpenFile(filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Fatalf("failed opening file: %s", err)
	}
	defer file.Close()

	len, err := file.WriteString(data)
	if err != nil {
		log.Fatalf("failed writing to file: %s", err)
	}
	fmt.Printf("\nLength: %d bytes", len)
	fmt.Printf("\nFile Name: %s", file.Name())
}

func main() {
	for _, command := range commands {
		file := "../gen-" + command.file + ".go"

		_, err := os.Stat(file)

		if err == nil {
			os.Remove(file)
		}
	}

	for _, command := range commands {
		file := "../gen-" + command.file + ".go"

		_package := ""

		_, err := os.Stat(file)

		// fmt.Println("err", err)

		if err != nil {
			if command.returnType != "" {
				_package = `package aws
	
	import (
		"encoding/json"
		"fmt"
		"os"
		"strings"
	)`
			} else {
				_package = `package aws
			import (
				"fmt"
				"os"
			)`
			}
		}

		_returnType := "error"

		if command.returnType != "" {
		}

		cliParams := "aws.Region"

		if command.cliParamsParams != "" {
			cliParams = cliParams + ", " + command.cliParamsParams
		}

		_return := "err"
		_returnBody := ""
		_returnVars := "err"

		if command.returnType != "" {
			_returnType = _returnType + ", " + command.returnType
			_return = _return + ", resp"
			_returnBody = `
		var resp ` + command.returnType + `
	
		json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)`
			_returnVars = _returnVars + ", _resp"
		} else {
			_returnVars = _returnVars + ", _"

		}

		_returnVars = _returnVars + ", stderr"

		data := fmt.Sprintf(`
%s

func (aws *AWS) %s(%s) (%s) {
	%s := Execute(fmt.Sprintf("%s --region %%s %s", %s))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}
	%s	

	return %s
}
		`, _package, command.method, command.params, _returnType, _returnVars, command.command, command.cliParams, cliParams, _returnBody, _return)

		AppendFile("../gen-"+command.file+".go", data)
	}
}
