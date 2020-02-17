
package aws
	
	import (
		"encoding/json"
		"fmt"
		"os"
		"strings"
	)

func (aws *AWS) CFDescribeStacks() (error, CFDescribeStacksResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws cloudformation describe-stacks --profile %s --region %s ", aws.Profile, aws.Region), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}
	
		var resp CFDescribeStacksResponse
	
		json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)	

	return err, resp
}
		


func (aws *AWS) CFDeleteStack(stackName string) (error) {
	err, _, stderr := Execute(fmt.Sprintf("aws cloudformation delete-stack --profile %s --region %s --stack-name %s", aws.Profile, aws.Region, stackName), true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}
		

	return err
}
		