package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) GetRouteTables() (error, EC2DescribeRouteTablesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-route-tables --region %s ", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2DescribeRouteTablesResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) CreateRouteTable(vpc AWSVPCType) (error, EC2CreateRouteTableResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-route-table --region %s --vpc-id %s", aws.Region, vpc.VpcId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2CreateRouteTableResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) AssociateRouteTable(routeTableId, subnetId string) (error, EC2AssociateRouteTableResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 associate-route-table --region %s --route-table-id %s --subnet-id %s", aws.Region, routeTableId, subnetId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2AssociateRouteTableResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) DisassociateRouteTable(associationId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 disassociate-route-table --region %s --association-id %s", aws.Region, associationId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}

func (aws *AWS) DeleteRouteTable(routeTableId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 delete-route-table --region %s --route-table-id %s", aws.Region, routeTableId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}

func (aws *AWS) CreateRouteWithInternetGateway(routeTableId, cidr, gatewayId string) (error, EC2CreateRouteResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-route --region %s --route-table-id %s --destination-cidr-block %s --gateway-id %s", aws.Region, routeTableId, cidr, gatewayId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2CreateRouteResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) CreateRouteWithNatGateway(routeTableId, cidr, gatewayId string) (error, EC2CreateRouteResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-route --region %s --route-table-id %s --destination-cidr-block %s --nat-gateway-id %s", aws.Region, routeTableId, cidr, gatewayId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2CreateRouteResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}
