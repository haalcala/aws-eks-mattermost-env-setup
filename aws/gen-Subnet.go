package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) CreateDefaultSubnet(availabilityZone string) (error, EC2CreateSubnetResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 create-default-subnet --region %s --availability-zone %s", aws.Region, availabilityZone))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp EC2CreateSubnetResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) DeleteSubnet(subnetId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 delete-subnet --region %s --subnet-id %s", aws.Region, subnetId))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}
