package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) AllocateAddress() (error, AWSAddressType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 allocate-address --region %s ", aws.Region))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp AWSAddressType

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) GetAddresses() (error, AWSEC2DescribeAddressesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-addresses --region %s ", aws.Region))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp AWSEC2DescribeAddressesResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}

func (aws *AWS) DisassociateAddress(associationId string) error {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 disassociate-address --region %s --association-id %s", aws.Region, associationId))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	return err
}
