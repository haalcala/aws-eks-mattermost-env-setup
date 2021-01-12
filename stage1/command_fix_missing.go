package main

import (
	"fmt"
	"io/ioutil"
	"strings"

	aws_util "../aws"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
)

// this is just a comment
func (m *MMDeployContext) GetOrCreateDBVPCSecurityGroup() (*ec2.SecurityGroup, error) {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateDBVPCSecurityGroup() (*ec2.SecurityGroup, error)")

	sg, err := m.EC2.CreateSecurityGroup(&ec2.CreateSecurityGroupInput{
		GroupName:   aws.String(m.DeployConfig.RDS.DBSecurityGroupName),
		Description: aws.String(m.DeployConfig.RDS.DBSecurityGroupName),
		VpcId:       m.EKSCluster.ResourcesVpcConfig.VpcId,

		TagSpecifications: []*ec2.TagSpecification{
			&ec2.TagSpecification{
				ResourceType: aws.String("security-group"),
				Tags: []*ec2.Tag{
					&ec2.Tag{Key: aws.String("alpha.eksctl.io/cluster-name"), Value: aws.String(m.DeployConfig.ClusterName)},
					&ec2.Tag{Key: aws.String("eksctl.cluster.k8s.io/v1alpha1/cluster-name"), Value: aws.String(m.DeployConfig.ClusterName)},
					&ec2.Tag{Key: aws.String("harold-cluster-create-tool-version"), Value: aws.String("0.0.1")},
					&ec2.Tag{Key: aws.String("Name"), Value: aws.String(m.DeployConfig.RDS.DBSecurityGroupName)},
				},
			},
		},
	})
	if err != nil && !strings.Contains(err.Error(), "already exists") {
		aws_util.ExitErrorf("Unable to create cluster, %v", err)
	}

	fmt.Println("sg:", sg, sg != nil)

	if sg != nil && sg.GroupId != nil {
		authorise_sg_output, err := m.EC2.AuthorizeSecurityGroupIngress(&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId: sg.GroupId,
			// GroupName:  aws.String(m.DeployConfig.RDS.DBSecurityGroupName),
			// IpProtocol: aws.String("tcp"),
			// ToPort:     aws.Int64(3306),
			// CidrIp:     aws.String("0.0.0.0/0"),
			IpPermissions: []*ec2.IpPermission{
				// Can use setters to simplify seting multiple values without the
				// needing to use aws.String or associated helper utilities.
				(&ec2.IpPermission{}).
					SetIpProtocol("tcp").
					SetFromPort(3306).
					SetToPort(3306).
					SetIpRanges([]*ec2.IpRange{
						{CidrIp: aws.String("0.0.0.0/0")},
					}),
			},
		})
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			aws_util.ExitErrorf("Unable to authorize security group ingress, %v", err)
		}

		fmt.Println("authorise_sg_output:", authorise_sg_output)
	}

	sg2, err := m.EC2.DescribeSecurityGroups(nil)

	if err != nil && !strings.Contains(err.Error(), "already exists") {
		aws_util.ExitErrorf("Unable to create cluster, %v", err)
	}

	fmt.Println("sg2:", sg2)

	for _, _sg := range sg2.SecurityGroups {
		if *_sg.GroupName == m.DeployConfig.RDS.DBSecurityGroupName {
			return _sg, nil
		}
	}

	return nil, nil
}

// this is just a comment
func (m *MMDeployContext) GetOrCreateDBSubnetGroup() (*rds.DBSubnetGroup, error) {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateDBSubnetGroup() (*rds.DBSubnetGroup, error)")

	create_subnetgroup_output, err := m.RDS.CreateDBSubnetGroup(&rds.CreateDBSubnetGroupInput{
		DBSubnetGroupName:        aws.String(m.DeployConfig.ClusterName),
		SubnetIds:                m.EKSCluster.ResourcesVpcConfig.SubnetIds,
		DBSubnetGroupDescription: aws.String(m.DeployConfig.ClusterName),
	})

	if err != nil && !strings.Contains(err.Error(), "already exists") {
		aws_util.ExitErrorf("Unable to create cluster, %v", err)
	}

	fmt.Println("create_subnetgroup_output:", create_subnetgroup_output)

	db_subnet_groups, err := m.RDS.DescribeDBSubnetGroups(&rds.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(m.DeployConfig.ClusterName),
	})

	for _, subnet_group := range db_subnet_groups.DBSubnetGroups {
		if *subnet_group.DBSubnetGroupName == m.DeployConfig.ClusterName {
			return subnet_group, nil
		}
	}

	return nil, nil
}

// this is just a comment
func (m *MMDeployContext) GetOrCreateDBParameterGroup() (*rds.DBParameterGroup, error) {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateDBParameterGroup() (*rds.DBParameterGroup,error)")

	m.RDS.CreateDBClusterParameterGroup(&rds.CreateDBClusterParameterGroupInput{
		DBClusterParameterGroupName: aws.String("default.mysql8.0"),
		DBParameterGroupFamily:      aws.String(""),
	})

	return nil, nil
}

// this is just a comment
func (m *MMDeployContext) GetDBInstance() (*rds.DBInstance, error) {
	fmt.Println("------ func (m *MMDeployContext) GetDBInstance() (*rds.DBInstance,error)")

	dbs, err := m.RDS.DescribeDBInstances(nil)

	if err != nil {
		aws_util.ExitErrorf("Unable to create cluster, %v", err)
	}

	fmt.Println("dbs:", dbs)

	for _, db := range dbs.DBInstances {
		if *db.DBInstanceIdentifier == m.DeployConfig.RDS.DBInstanceName {
			return db, nil
		}
	}

	return nil, nil
}

// this is just a comment
func (m *MMDeployContext) GetOrCreateDB() (*rds.DBInstance, error) {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateDB() (*rds.DBInstance,error)")

	db, err := m.GetDBInstance()

	if err != nil {
		return nil, err
	}

	if db == nil {
		fmt.Println("m.EKSCluster:", m.EKSCluster)
		fmt.Println("m.EKSCluster.ResourcesVpcConfig.VpcId:", *m.EKSCluster.ResourcesVpcConfig.VpcId)

		fmt.Println("SubnetIds:", m.EKSCluster.ResourcesVpcConfig.SubnetIds)

		subnets, err := m.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: m.EKSCluster.ResourcesVpcConfig.SubnetIds,
		})

		if err != nil {
			aws_util.ExitErrorf("Unable to create cluster, %v", err)
		}

		fmt.Println("subnets:", subnets)

		sg, err := m.GetOrCreateDBVPCSecurityGroup()

		if err != nil && !strings.Contains(err.Error(), "already exists") {
			aws_util.ExitErrorf("Unable to create cluster, %v", err)
		}

		fmt.Println("sg:", sg)

		db_subnet_group, err := m.GetOrCreateDBSubnetGroup()
		if err != nil && !strings.Contains(err.Error(), "already exists") {
			aws_util.ExitErrorf("Unable to create cluster, %v", err)
		}

		fmt.Println("db_subnet_group:", db_subnet_group)

		create_instance_output, err := m.RDS.CreateDBInstance(&rds.CreateDBInstanceInput{
			AllocatedStorage: aws.Int64(20),
			// DBClusterIdentifier:     &m.DeployConfig.ClusterName,
			DBName:               aws.String("test"),
			DBInstanceClass:      aws.String("db.t2.micro"),
			DBInstanceIdentifier: aws.String(m.DeployConfig.RDS.DBInstanceName),
			// DBSecurityGroups:        aws.StringSlice([]string{*sg.GroupId}),
			Engine:                  aws.String("MySQL"),
			EngineVersion:           aws.String("8.0.21"),
			MasterUserPassword:      aws.String("vcube2192"),
			MasterUsername:          aws.String("admin"),
			Port:                    aws.Int64(3306),
			StorageType:             aws.String("gp2"),
			MaxAllocatedStorage:     aws.Int64(1000),
			CopyTagsToSnapshot:      aws.Bool(true),
			AutoMinorVersionUpgrade: aws.Bool(true),
			DBParameterGroupName:    aws.String("default.mysql8.0"),
			OptionGroupName:         aws.String("default:mysql-8-0"),
			AvailabilityZone:        aws.String("us-east-2a"),
			VpcSecurityGroupIds:     aws.StringSlice([]string{*sg.GroupId}),
			DBSubnetGroupName:       db_subnet_group.DBSubnetGroupName,
			Tags: []*rds.Tag{
				&rds.Tag{Key: aws.String("alpha.eksctl.io/cluster-name"), Value: aws.String(m.DeployConfig.ClusterName)},
				&rds.Tag{Key: aws.String("eksctl.cluster.k8s.io/v1alpha1/cluster-name"), Value: aws.String(m.DeployConfig.ClusterName)},
				&rds.Tag{Key: aws.String("harold-cluster-create-tool-version"), Value: aws.String("0.0.1")},
				&rds.Tag{Key: aws.String("Name"), Value: aws.String(m.DeployConfig.RDS.DBInstanceName)},
			},
			MultiAZ:            aws.Bool(false),
			PubliclyAccessible: aws.Bool(false),
			StorageEncrypted:   aws.Bool(false),
			// ProcessorFeatures: []*rds.ProcessorFeature{
			// 	&rds.ProcessorFeature{Name: aws.String("cpu"), Value: aws.String("1")},
			// },
		})

		if err != nil {
			aws_util.ExitErrorf("Unable to create cluster, %v", err)
		}

		fmt.Println("create_instance_output:", create_instance_output)

		m.RDS.WaitUntilDBInstanceAvailable(&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(m.DeployConfig.RDS.DBInstanceName),
		})
	}

	db, err = m.GetDBInstance()

	if *db.DBInstanceStatus == aws_util.RDS_DB_INSTANCE_STATUS_CREATING {
		m.RDS.WaitUntilDBInstanceAvailable(&rds.DescribeDBInstancesInput{
			DBInstanceIdentifier: aws.String(m.DeployConfig.RDS.DBInstanceName),
		})
	}

	return db, err
}

// this is just a comment
func (m *MMDeployContext) GetAWSLoadBalancerControllerIAMPolicy() (*iam.Policy, error) {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateALBIAMPolicy() error")

	if policies, err := m.IAM.ListPolicies(nil); err != nil {
		return nil, err
	} else {
		for _, policy := range policies.Policies {
			if *policy.PolicyName == m.DeployConfig.AWSLoadBalancerControllerIAMPolicyName {
				return policy, nil
			}
		}
	}

	return nil, nil
}

// this is just a comment
func (m *MMDeployContext) GetOrCreateALBIAMPolicy() (*iam.Policy, error) {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateALBIAMPolicy() error")

	iam_policy, err := m.GetAWSLoadBalancerControllerIAMPolicy()

	if err != nil {
		return nil, err
	}

	err, out1, out2 := aws_util.Execute(fmt.Sprintf("eksctl utils associate-iam-oidc-provider --region %v --cluster %v --approve", m.DeployConfig.Region, m.DeployConfig.ClusterName), true, true)

	fmt.Println("err:", err)
	fmt.Println("out1:", out1)
	fmt.Println("out2:", out2)

	if iam_policy == nil {
		err, out1, out2 = aws_util.Execute("curl -o iam-policy.json https://raw.githubusercontent.com/kubernetes-sigs/aws-load-balancer-controller/v2.1.0/docs/install/iam_policy.json", true, true)

		fmt.Println("err:", err)
		fmt.Println("out1:", out1)
		fmt.Println("out2:", out2)

		policy_document, err := ioutil.ReadFile("iam-policy.json")

		if err != nil {
			return nil, err
		}

		create_policy_output, err := m.IAM.CreatePolicy(&iam.CreatePolicyInput{
			PolicyName:     aws.String(m.DeployConfig.AWSLoadBalancerControllerIAMPolicyName),
			PolicyDocument: aws.String(string(policy_document)),
		})

		if err != nil && !strings.Contains(err.Error(), "already exists") {
			return nil, err
		}

		fmt.Println("create_policy_output:", create_policy_output)
	}

	return m.GetAWSLoadBalancerControllerIAMPolicy()
}

// this is just a comment
func (m *MMDeployContext) GetOrCreateEKSServiceAccount() error {
	fmt.Println("------ func (m *MMDeployContext) GetOrCreateEKSServiceAccount() error")

	err, out1, out2 := aws_util.Execute(
		fmt.Sprintf("eksctl --region=%v create iamserviceaccount --cluster=%v --namespace=kube-system --name=aws-load-balancer-controller --attach-policy-arn=%v --approve",
			m.DeployConfig.Region, m.DeployConfig.ClusterName, m.DeployConfig.AWSLoadBalancerControllerIAMPolicyARN), true, true)

	fmt.Println("err:", err)
	fmt.Println("out1:", out1)
	fmt.Println("out2:", out2)

	return nil
}

// this is just a comment
func (m *MMDeployContext) FixMissing() error {
	fmt.Println("------ func (m *MMDeployContext) FixMissing() error")

	db, err := m.GetOrCreateDB()

	if err != nil {
		return err
	}

	fmt.Println("db:", db)

	iam_policy, err := m.GetOrCreateALBIAMPolicy()

	if err != nil {
		return err
	}

	fmt.Println("iam_policy:", iam_policy)

	m.GetOrCreateEKSServiceAccount()

	// err, out1, out2 := aws_util.Execute("cd ../mattermost-env-setup-stage-2 ; source .staging_env ; go run generate_deployment.go common.go", true, true)

	// fmt.Println("err:", err)
	// fmt.Println("out1:", out1)
	// fmt.Println("out2:", out2)

	err, out1, out2 := aws_util.Execute("go run check_env.go", true, true)

	fmt.Println("err:", err)
	fmt.Println("out1:", out1)
	fmt.Println("out2:", out2)

	return nil
}
