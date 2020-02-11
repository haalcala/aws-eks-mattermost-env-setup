package aws

import (
	"fmt"
	"os"
)

func (aws *AWS) EC2DeleteVPC(vpcId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 delete-vpc --region %s --vpc-id %s", aws.Region, vpcId))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}
