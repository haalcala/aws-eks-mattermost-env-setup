package aws

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

type AWS struct {
	Region string
}

type EKS struct {
	aws        *AWS
	clusterKey string
}

func (aws *AWS) CreateTags(resourceIds, tags string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 create-tags --region %s --resources %s --tags %s", aws.Region, resourceIds, tags), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
	}

	return err
}

func (aws *AWS) GetStackWithTag(key, value string) (error, *AWSCFStackType) {
	err, describe_cloudformation_response, stderr := Execute(fmt.Sprintf("aws cloudformation describe-stacks --region %s", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
	}

	// fmt.Println("describe_cloudformation_response:", describe_cloudformation_response)

	var describe_cloudformation_json CFDescribeStacksResponse

	json.NewDecoder(strings.NewReader(describe_cloudformation_response)).Decode(&describe_cloudformation_json)

	fmt.Println("describe_cloudformation_json:", describe_cloudformation_json)

	resp, err := json.Marshal(describe_cloudformation_json)

	log.Println("resp:", string(resp))

	var stack *AWSCFStackType = nil

	if len(describe_cloudformation_json.Stacks) > 0 {
		for _, _stackPayload := range describe_cloudformation_json.Stacks {
			fmt.Println("_stackPayload:", _stackPayload)

			for _, _tag := range _stackPayload.Tags {
				fmt.Println("_tag:", _tag)
				if _tag.Key == key && _tag.Value == value {
					stack = &_stackPayload
					break
				}
			}

			if stack != nil {
				break
			}
		}
	}

	resp, err = json.Marshal(stack)

	log.Println("stack:", string(resp))

	return err, stack
}

func (aws *AWS) GetRDS() (error, RDSDescribeResponse) {
	err, get_rds_response, stderr := Execute("aws rds describe-db-instances --region "+aws.Region, true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var get_rds_json RDSDescribeResponse

	json.NewDecoder(strings.NewReader(get_rds_response)).Decode(&get_rds_json)

	fmt.Println("get_rds_response:", get_rds_response)

	return err, get_rds_json
}

func (aws *AWS) createRDS() {

}

func (aws *AWS) GetVPCs() (error, []AWSVPCType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-vpcs --region %s", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp DescribeVPCResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp.Vpcs
}

func (aws *AWS) CreateVPC(name, cidr_block string) (error, *AWSVPCType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-vpc --cidr-block %s --region %s", cidr_block, aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2CreateVPCResponse
	var vpc *AWSVPCType

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	vpc = resp.Vpc

	return err, vpc
}

func (aws *AWS) GetSubnetsByVPCId(vpcId string) (error, []AWSSubnetType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-subnets --region %s", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2DescribeSubnetsResponse
	var subnets = []AWSSubnetType{}

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	if len(resp.Subnets) > 0 {
		for _, _subnets := range resp.Subnets {
			if _subnets.VpcId == vpcId {
				subnets = append(subnets, _subnets)
			}
		}
	}

	return err, subnets
}

func (aws *AWS) CreateSubnet(vpcId, cidr, zone string) (error, AWSSubnetType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-subnet --region %s --vpc-id %s --cidr-block %s --availability-zone %s", aws.Region, vpcId, cidr, zone), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp *EC2CreateSubnetResponse = nil

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp.Subnet
}

func (aws *AWS) CreateInternetGateway() (error, *AWSInternetGatewayType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-internet-gateway --region %s", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp *AWSEC2CreateInternetGatewayResponse = nil

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp.InternetGateway
}

func (aws *AWS) CreateNatGateway(elasticIpAllocationId, subnetId string) (error, *AWSNatGatewayType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-nat-gateway --region %s --allocation-id %s --subnet-id %s", aws.Region, elasticIpAllocationId, subnetId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp *AWSEC2CreateNatGatewayResponse = nil

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp.NatGateway
}

func (aws *AWS) GetNatGateways() (error, []AWSNatGatewayType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-nat-gateways --region %s", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp *AWSEC2DescribeNatGatewaysResponse = nil

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp.NatGateways
}

func (aws *AWS) AttachInternetGatewayToVPC(vpcId, internetGatewayId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 attach-internet-gateway --region %s --internet-gateway-id %s --vpc-id %s", aws.Region, internetGatewayId, vpcId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}

// Availability Zones

func (aws *AWS) GetAvailabilityZones() (error, []AWSAvailabilityZoneType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-availability-zones --region %s", aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2DescribeAvailabilityZonesReponse
	var availabilityZones = []AWSAvailabilityZoneType{}

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	if len(resp.AvailabilityZones) > 0 {
		for _, _availabilityZone := range resp.AvailabilityZones {
			if _availabilityZone.RegionName == aws.Region {
				availabilityZones = append(availabilityZones, _availabilityZone)
			}
		}
	}

	return err, availabilityZones

}
