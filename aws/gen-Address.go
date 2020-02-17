
package aws
	
	import (
		"encoding/json"
		"fmt"
		"os"
		"strings"
	)

func (aws *AWS) AllocateAddress() (error, AWSAddressType) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 allocate-address --profile %s --region %s ", aws.Profile, aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}
	
		var resp AWSAddressType
	
		json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)	

	return err, resp
}
		


func (aws *AWS) GetAddresses() (error, AWSEC2DescribeAddressesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws ec2 describe-addresses --profile %s --region %s ", aws.Profile, aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}
	
		var resp AWSEC2DescribeAddressesResponse
	
		json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)	

	return err, resp
}
		


func (aws *AWS) DisassociateAddress(associationId string) (error) {
	err, _, stderr := Execute(fmt.Sprintf("aws ec2 disassociate-address --profile %s --region %s --association-id %s", aws.Profile, aws.Region, associationId), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}
		

	return err
}
		