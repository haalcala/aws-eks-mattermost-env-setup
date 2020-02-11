package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) EC2DescribeInternetGateways() (error, AWSEC2DescribeInternetGatewaysResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-internet-gateways --region %s ", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp AWSEC2DescribeInternetGatewaysResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) EC2DetachInternetGateway(internetGatewayId, vpcId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 detach-internet-gateway --region %s --internet-gateway-id %s --vpc-id %s", aws.Region, internetGatewayId, vpcId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}

func (aws *AWS) EC2DeleteInternetGateway(internetGatewayId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 delete-internet-gateway --region %s --internet-gateway-id %s", aws.Region, internetGatewayId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}

func (aws *AWS) EC2DeleteVPC(vpcId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 delete-vpc --region %s --vpc-id %s", aws.Region, vpcId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}
