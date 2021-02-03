package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
)

// this is just a comment
func (m *MMDeployContext) DeleteCluster() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteCluster() error")

	fmt.Println("Deleting cluster:", m.DeployConfig.ClusterName)

	err, stdout, stderr := Execute("kubectl delete -f deploy-aws-alb.yaml", true, true)

	fmt.Println("stdout:", stdout)
	fmt.Println("stderr:", stderr)

	if err != nil {
		return err
	}

	clusters, err := m.EKS.ListClusters(nil)

	if err != nil {
		return err
	}

	fmt.Println("clusters:", clusters)

	var cluster *eks.Cluster

	defer m.DeleteTargetGroups()

	for _, _cluster := range clusters.Clusters {
		if *_cluster == m.DeployConfig.ClusterName {
			__cluster, err := m.EKS.DescribeCluster(&eks.DescribeClusterInput{
				Name: clusters.Clusters[0],
			})

			cluster = __cluster.Cluster

			if err != nil {
				return err
			}

		}
	}

	fmt.Println("cluster:", cluster)

	if cluster == nil {
		return nil
	}

	m.EKSCluster = cluster

	var dbInstance *rds.DBInstance

	if dbs, err := m.RDS.DescribeDBInstances(nil); err == nil {
		// fmt.Println("dbs:", dbs)
	dbInstance:
		for _, db := range dbs.DBInstances {
			// fmt.Println("db:", db)
			for _, subnet := range cluster.ResourcesVpcConfig.SubnetIds {

				// fmt.Println("subnet:", *subnet)

				for _, _subnet := range *&db.DBSubnetGroup.Subnets {
					// fmt.Println("_subnet:", _subnet)
					if *_subnet.SubnetIdentifier == *subnet {
						dbInstance = db
						break dbInstance
					}
				}
			}
		}
	}

	fmt.Println("dbInstance:", dbInstance)

	delete_db_result := make(chan string, 1)
	delete_fg_profile_result := make(chan string, 1)

	if dbInstance != nil {
		go m.DeleteDB(*dbInstance.DBInstanceIdentifier, delete_db_result)
	} else {
		delete_db_result <- "done"
	}

	if cluster != nil {
		go m.DeleteFargateProfiles(delete_fg_profile_result)
	} else {
		delete_fg_profile_result <- "done"
	}

	fmt.Println("Delete db result: ", <-delete_db_result)
	fmt.Println("Delete Fargate Profiles result:", <-delete_fg_profile_result)

	_iam := m.IAM

	oidc_providers, err := _iam.ListOpenIDConnectProviders(nil)

	if err != nil {
		return err
	}

	fmt.Println("oidc_providers:", oidc_providers.OpenIDConnectProviderList)

	var iodc_provider *iam.OpenIDConnectProviderListEntry

	for _, _iodc_provider := range oidc_providers.OpenIDConnectProviderList {
		if strings.HasSuffix(*_iodc_provider.Arn, (*cluster.Identity.Oidc.Issuer)[len("https://"):]) {
			iodc_provider = _iodc_provider
		}
	}

	fmt.Println("iodc_provider:", iodc_provider)

	if iodc_provider != nil {
		fmt.Println("Deleting iodc_provider:", iodc_provider)

		_iam.DeleteOpenIDConnectProvider(&iam.DeleteOpenIDConnectProviderInput{
			OpenIDConnectProviderArn: iodc_provider.Arn,
		})
	}

	stacks, err := m.CF.ListStacks(&cloudformation.ListStacksInput{
		StackStatusFilter: []*string{aws.String("CREATE_COMPLETE"), aws.String("DELETE_FAILED")},
	})

	fmt.Println("stacks:", stacks)

	var stack *cloudformation.Stack

	for _, _stack := range stacks.StackSummaries {
		__stack, err := m.CF.DescribeStacks(&cloudformation.DescribeStacksInput{
			StackName: _stack.StackName,
		})

		stack = __stack.Stacks[0]

		if err != nil {
			return err
		}

		fmt.Println("stack:", stack)

		in_this_cluster := false
		process_stack := false

		for _, _tag := range stack.Tags {
			if *_tag.Key == "alpha.eksctl.io/iamserviceaccount-name" && *_tag.Value == "kube-system/aws-load-balancer-controller" {
				process_stack = true
			} else if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.DeployConfig.ClusterName {
				in_this_cluster = true
			}
		}

		if in_this_cluster && process_stack {
			fmt.Println("Will process stack:", stack)

			// stack.Stacks[0].Outputs
		}

		// roles, err := _iam.ListRoles(nil)

		// if err != nil {
		// 	return err
		// }

		// // fmt.Println("roles:", roles)

		// for _, role := range roles.Roles {
		// 	in_this_cluster := false
		// 	is_service_account := false

		// 	role, err := _iam.GetRole(&iam.GetRoleInput{
		// 		RoleName: role.RoleName,
		// 	})

		// 	// fmt.Println("role:", role)

		// 	if err != nil {
		// 		return err
		// 	}

		// 	for _, _tag := range role.Role.Tags {
		// 		if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.DeployConfig.ClusterName {
		// 			in_this_cluster = true
		// 		} else if *_tag.Key == "alpha.eksctl.io/iamserviceaccount-name" && *_tag.Value == "kube-system/aws-load-balancer-controller" {
		// 			is_service_account = true
		// 		}
		// 	}

		// 	if in_this_cluster && is_service_account {
		// 		fmt.Println("Deleting role:", role)

		// 		detach_output, err := _iam.DetachRolePolicy(&iam.DetachRolePolicyInput{
		// 			PolicyArn: role.Role.Arn,
		// 			RoleName:  role.Role.RoleName,
		// 		})

		// 		if err != nil {
		// 			return err
		// 		}

		// 		fmt.Println("detach_output:", detach_output)

		// 		delete_output, err := _iam.DeleteRole(&iam.DeleteRoleInput{
		// 			RoleName: role.Role.RoleName,
		// 		})

		// 		if err != nil {
		// 			return err
		// 		}

		// 		fmt.Println("delete_output:", delete_output)
		// 	}
		// }
	}

	delete_cluster_result := make(chan string, 1)
	if stack != nil {
		go func() {
			Execute(fmt.Sprintf("eksctl --region %v delete cluster --name %v -w", m.DeployConfig.Region, m.DeployConfig.ClusterName), true, true)
			delete_cluster_result <- "done"
		}()
	} else {
		delete_cluster_result <- "done"
	}

	<-delete_cluster_result

	fmt.Println("All done!")

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteOtherStacks() error {
	found_cluster_stack := true

	for found_cluster_stack {
		found_cluster_stack = false
		stacks, err := m.CF.ListStacks(&cloudformation.ListStacksInput{
			StackStatusFilter: []*string{aws.String("CREATE_COMPLETE"), aws.String("DELETE_FAILED"), aws.String("DELETE_IN_PROGRESS")},
		})

		if err != nil {
			return err
		}

		fmt.Println("stacks:", stacks)

		var stack *cloudformation.Stack

		for _, _stack := range stacks.StackSummaries {
			__stack, err := m.CF.DescribeStacks(&cloudformation.DescribeStacksInput{
				StackName: _stack.StackName,
			})

			stack = __stack.Stacks[0]

			if err != nil {
				return err
			}

			fmt.Println("stack:", stack)

			in_this_cluster := false

			for _, _tag := range stack.Tags {
				if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.DeployConfig.ClusterName {
					in_this_cluster = true
					found_cluster_stack = true
				}
			}

			if in_this_cluster && (*stack.StackStatus == "CREATE_COMPLETE" || *stack.StackStatus == "DELETE_FAILED") {
				fmt.Println("Deleting related stack:", stack)

				delete_stack_output, err := m.CF.DeleteStack(&cloudformation.DeleteStackInput{
					StackName: stack.StackName,
				})

				if err != nil {
					return err
				}

				fmt.Println("delete_stack_output:", delete_stack_output)
			}
		}

		if found_cluster_stack {
			time.Sleep(time.Second * 10)
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteTargetGroups() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteTargetGroups() error")

	tgs, err := m.ELB.DescribeTargetGroups(nil)
	if err != nil {
		return err
	}
	for _, tg := range tgs.TargetGroups {
		tags, err := m.ELB.DescribeTags(&elbv2.DescribeTagsInput{
			ResourceArns: []*string{tg.TargetGroupArn},
		})
		if err != nil {
			return err
		}
		fmt.Println("tags:", tags)

		is_in_cluster := false

		for _, _tag := range tags.TagDescriptions[0].Tags {
			fmt.Println("Key:", *_tag.Key, "Value:", *_tag.Value)
			if *_tag.Key == "kubernetes.io/cluster/"+m.DeployConfig.ClusterName && *_tag.Value == "owned" {
				is_in_cluster = true
			}
			fmt.Println("is_in_cluster:", is_in_cluster)
		}

		fmt.Println("--->>> is_in_cluster:", is_in_cluster, m.DeployConfig.ClusterName+"-alb-ingress")

		if is_in_cluster {
			delete_tg_resp, err := m.ELB.DeleteTargetGroup(&elbv2.DeleteTargetGroupInput{
				TargetGroupArn: tg.TargetGroupArn,
			})
			if err != nil {
				return err
			}

			fmt.Println("delete_tg_resp:", delete_tg_resp)
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteVPC(vpcId string) error {
	fmt.Println("------ func (m *MMDeployContext) DeleteVPC(vpcId string, result chan string) error")

	_ec2 := ec2.New(m.Session)

	delete_vpc_result, err := _ec2.DeleteVpc(&ec2.DeleteVpcInput{
		VpcId: aws.String(vpcId),
	})

	if err != nil {
		return err
	}

	fmt.Println("delete_vpc_result:", delete_vpc_result)

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteClusterVPCSecurityGroups(vpcId string) error {
	fmt.Println("------ func (m *MMDeployContext) DeleteClusterVPCSecurityGroups() error")

	sgs, err := m.EC2.DescribeSecurityGroups(nil)

	if err != nil {
		return err
	}

	for _, sg := range sgs.SecurityGroups {
		if *sg.VpcId == vpcId && *sg.GroupName != "default" {

			fmt.Println("Deleting security group:", fmt.Sprintf("Deleteing SecurityGroup: %v (%v)\n", sg.GroupName, sg.GroupId))

			_, err := m.EC2.DeleteSecurityGroup(&ec2.DeleteSecurityGroupInput{
				GroupId: sg.GroupId,
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteELBs() error {
	elbs, err := m.ELB.DescribeLoadBalancers(nil)

	if err != nil {
		return err
	}

	fmt.Println("elbs:", elbs)

	for _, _elb := range elbs.LoadBalancers {
		tags, err := m.ELB.DescribeTags(&elbv2.DescribeTagsInput{
			ResourceArns: []*string{_elb.LoadBalancerArn},
		})
		if err != nil {
			return err
		}
		fmt.Println("tags:", tags)

		is_in_cluster := false
		ingress_name := ""

		for _, _tag := range tags.TagDescriptions[0].Tags {
			fmt.Println("Key:", *_tag.Key, "Value:", *_tag.Value)
			if *_tag.Key == "ingress.k8s.aws/cluster" && *_tag.Value == m.DeployConfig.ClusterName {
				is_in_cluster = true
			}
			if *_tag.Key == "kubernetes.io/ingress-name" {
				ingress_name = *_tag.Value
			}
			fmt.Println("is_in_cluster:", is_in_cluster, "ingress_name:", ingress_name)
		}

		fmt.Println("--->>> is_in_cluster:", is_in_cluster, "ingress_name:", ingress_name, m.DeployConfig.ClusterName+"-alb-ingress", ingress_name == m.DeployConfig.ClusterName+"-alb-ingress")

		if is_in_cluster {
			delete_elb_resp, err := m.ELB.DeleteLoadBalancer(&elbv2.DeleteLoadBalancerInput{
				LoadBalancerArn: _elb.LoadBalancerArn,
			})
			if err != nil {
				return err
			}

			fmt.Println("delete_elb_resp:", delete_elb_resp)
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteSubnets() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteSubnets() error")

	subnets, err := m.EC2.DescribeSubnets(nil)

	if err != nil {
		return err
	}

	for _, subnet := range subnets.Subnets {
		in_this_cluster := false

		for _, _tag := range subnet.Tags {
			if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.DeployConfig.ClusterName {
				in_this_cluster = true
			}
		}

		if in_this_cluster {
			delete_subnet_resp, err := m.EC2.DeleteSubnet(&ec2.DeleteSubnetInput{
				SubnetId: subnet.SubnetId,
			})
			if err != nil {
				return err
			}

			fmt.Println("delete_subnet_resp:", delete_subnet_resp)
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteClusterVPC() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteClusterVPC() error")

	vpcs, err := m.EC2.DescribeVpcs(nil)

	if err != nil {
		return err
	}

	var vpc *ec2.Vpc

	for _, _vpc := range vpcs.Vpcs {
		for _, tag := range _vpc.Tags {
			if *tag.Key == "alpha.eksctl.io/cluster-name" && *tag.Value == m.DeployConfig.ClusterName {
				vpc = _vpc
			}
		}
	}

	fmt.Println("vpc:", vpc)

	err = m.DeleteELBs()
	if err != nil {
		return err
	}

	err = m.DeleteSubnets()
	if err != nil {
		return err
	}

	igs, err := m.EC2.DescribeInternetGateways(nil)
	if err != nil {
		return err
	}

	fmt.Println("igs:", igs)

	if vpc != nil {
		for _, ig := range igs.InternetGateways {
			in_this_cluster := false

			for _, tag := range ig.Tags {
				if *tag.Key == "alpha.eksctl.io/cluster-name" && *tag.Value == m.DeployConfig.ClusterName {
					in_this_cluster = true
				}
			}

			if in_this_cluster {
				for _, att := range ig.Attachments {
					if *att.VpcId == *vpc.VpcId {
						detach_igw_res, err := m.EC2.DetachInternetGateway(&ec2.DetachInternetGatewayInput{
							InternetGatewayId: ig.InternetGatewayId,
							VpcId:             vpc.VpcId,
						})

						if err != nil {
							return err
						}

						fmt.Println("detach_igw_res:", detach_igw_res)
					}
				}

				fmt.Println("Deleting internet gateway:", *ig.InternetGatewayId)

				delete_ig_resp, err := m.EC2.DeleteInternetGateway(&ec2.DeleteInternetGatewayInput{
					InternetGatewayId: ig.InternetGatewayId,
				})

				if err != nil {
					return err
				}

				fmt.Println("delete_ig_resp:", delete_ig_resp)
			}
		}

		m.DeleteClusterVPCSecurityGroups(*vpc.VpcId)

		fmt.Println("Deleting VPC:", *vpc.VpcId)

		_, err := m.EC2.DeleteVpc(&ec2.DeleteVpcInput{
			VpcId: vpc.VpcId,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteDB(dbInstanceIdentifier string, result chan string) error {
	fmt.Println("------ func (m *MMDeployContext) DeleteDB(dbInstanceIdentifier string, result chan string) error")

	fmt.Println("Deleting dbInstance:", dbInstanceIdentifier)

	for {
		fmt.Println("Checking instance status ... ")

		dbs, err := m.RDS.DescribeDBInstances(nil)

		if err != nil {
			return err
		}

		fmt.Println("dbs:", len(dbs.DBInstances))

		if len(dbs.DBInstances) == 0 {
			break
		}

		var dbInstance *rds.DBInstance

		for _, db := range dbs.DBInstances {
			if *db.DBInstanceIdentifier == m.DeployConfig.RDS.DBInstanceName {
				dbInstance = db
				break
			}
		}

		if dbInstance == nil {
			return nil
		}

		fmt.Println("Instance status:", *dbs.DBInstances[0].DBInstanceStatus)

		for _, db := range dbs.DBInstances {
			if *db.DBInstanceIdentifier == dbInstanceIdentifier && *db.DBInstanceStatus == "available" {
				delete_output, err := m.RDS.DeleteDBInstance(&rds.DeleteDBInstanceInput{
					DBInstanceIdentifier: aws.String(dbInstanceIdentifier),
					SkipFinalSnapshot:    aws.Bool(true),
				})

				if err != nil {
					return err
				}

				fmt.Println("delete_output:", delete_output)

				err = m.RDS.WaitUntilDBInstanceDeleted(&rds.DescribeDBInstancesInput{
					DBInstanceIdentifier: &dbInstanceIdentifier,
				})

				if err != nil {
					return err
				}
			}

			if *db.DBInstanceStatus == "deleting" || *db.DBInstanceStatus == "creating" || *db.DBInstanceStatus == "backing-up" {
				time.Sleep(time.Second * 10)
				continue
			}
		}
	}

	m.DeleteSubnetGroup()

	result <- "done"

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteSubnetGroup() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteSubnetGroup() error")

	subnet_groups, err := m.RDS.DescribeDBSubnetGroups(nil)

	if err != nil {
		return err
	}

	for _, subnet_group := range subnet_groups.DBSubnetGroups {
		if *subnet_group.DBSubnetGroupName == m.DeployConfig.ClusterName {
			m.RDS.DeleteDBSubnetGroup(&rds.DeleteDBSubnetGroupInput{
				DBSubnetGroupName: aws.String(m.DeployConfig.ClusterName),
			})
		}

	}

	return nil
}
