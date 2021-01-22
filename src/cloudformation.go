package main

import (
	"fmt"

	aws_util "../aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
)

// this is just a comment
func (m *MMDeployContext) GetCloudFormationMainStack() (*cloudformation.Stack, error) {
	fmt.Println("------ func (m *MMDeployContext) GetCloudFormationMainStack(result chan string) error")

	stacks, err := m.CF.ListStacks(&cloudformation.ListStacksInput{
		StackStatusFilter: aws.StringSlice([]string{"CREATE_COMPLETE", "DELETE_FAILED", "CREATE_FAILED"}),
	})

	if err != nil {
		aws_util.ExitErrorf("Unable to create session, %v", err)
	}

	for _, _stack := range stacks.StackSummaries {
		if *_stack.StackName == "eksctl-"+m.DeployConfig.ClusterName+"-cluster" {
			stack, err := m.CF.DescribeStacks(&cloudformation.DescribeStacksInput{
				StackName: _stack.StackName,
			})

			if err != nil {
				aws_util.ExitErrorf("Unable to create session, %v", err)
			}

			return stack.Stacks[0], nil
		}
	}

	return nil, nil
}
