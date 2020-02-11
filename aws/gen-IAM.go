package aws

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func (aws *AWS) IAMListRoles() (error, IAMListRolesResponse) {
	err, _resp, stderr := Execute(fmt.Sprintf("aws iam list-roles --region %s ", aws.Region))

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	var resp IAMListRolesResponse

	json.NewDecoder(strings.NewReader(_resp)).Decode(&resp)

	return err, resp
}
