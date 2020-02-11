package aws

import (
	"fmt"
	"os"
)

func (aws *AWS) DeleteNatGateway(natGatewayId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 delete-nat-gateway --region %s --nat-gateway-id %s", aws.Region, natGatewayId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}
