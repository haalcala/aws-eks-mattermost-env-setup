package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	aws_util "../aws"
	stage2 "../stage2"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
)

type MMDeployEnvironment struct {
	ClusterName                            string                                 `json:"ClusterName"`
	AvailabilityZones                      []string                               `json:"AvailabilityZones"`
	Subnets                                []string                               `json:"Subnets`
	VpcId                                  string                                 `json:"VpcId"`
	Region                                 string                                 `json:"Region"`
	KubernetesVersion                      string                                 `json:"KubernetesVersion"`
	AWSLoadBalancerControllerIAMPolicyName string                                 `json:"AWSLoadBalancerControllerIAMPolicyName"`
	AWSLoadBalancerControllerIAMPolicyARN  string                                 `json:"AWSLoadBalancerControllerIAMPolicyARN"`
	RDS                                    MMDeployEnvironment_RDS                `json:"RDS"`
	MattermostInstance                     MMDeployEnvironment_MattermostInstance `json:"MattermostInstance"`
	InfraComponents                        MMDeployEnvironment_InfraComponents    `json:"InfraComponents"`
	OutputDir                              string                                 `json:"OutputDir"`
}

type MMDeployEnvironment_InfraComponents struct {
	PushProxy struct {
		URL string `json:"URL"`
	} `json:"PushProxy"`
}

type MMDeployEnvironment_RDS struct {
	RDSName             string `json:"RDSName"`
	RDSDeployAZ         string `json:"RDSDeployAZ"`
	DBSecurityGroupName string `json:"DBSecurityGroupName"`
	DBInstanceName      string `json:"DBInstanceName"`
}

type MMDeployEnvironment_MattermostInstance struct {
	Cluster       MMDeployEnvironment_MattermostInstance_Cluster `json:"Cluster"`
	PushServerUrl string                                         `json:"PushServerUrl"`
	DBHost        string                                         `json:"DBHost"`
	DBPort        string                                         `json:"DBPort"`
	AWSKey        string                                         `json:"AWSKey"`
	AWSSecret     string                                         `json:"AWSSecret"`
	SMTP          MMDeployEnvironment_MattermostInstance_SMTP    `json:"SMTP"`
	ListenPort    string                                         `json:"ListenPort"`
}

type MMDeployEnvironment_MattermostInstance_Cluster struct {
	Driver                 string `json:"Driver"`
	CustomClusterRedisHost string `json:"CustomClusterRedisHost"`
	CustomClusterRedisPort string `json:"CustomClusterRedisPort"`
	CustomClusterRedisUser string `json:"CustomClusterRedisUser"`
	CustomClusterRedisPass string `json:"CustomClusterRedisPass"`
}

type MMDeployEnvironment_MattermostInstance_SMTP struct {
	Host string `json:"Host"`
	Port string `json:"Port"`
	User string `json:"User"`
	Pass string `json:"Pass"`
}

// MMDeployContext bla bla bla
type MMDeployContext struct {
	Context    *MMDeployEnvironment
	Session    *session.Session
	EKSCluster *eks.Cluster
	EC2        *ec2.EC2
	EKS        *eks.EKS
	CF         *cloudformation.CloudFormation
	IAM        *iam.IAM
	RDS        *rds.RDS
	ConfigFile string
}

func MMDeployEnvironmentToJsonString(c *MMDeployEnvironment) (string, error) {
	b, err := json.MarshalIndent(c, "", "\t")

	return string(b), err
}

func MMDeployEnvironmentFromJson(_json string) (*MMDeployEnvironment, error) {
	c := &MMDeployEnvironment{}

	err := json.Unmarshal([]byte(_json), c)

	return c, err
}

func NewMMDeployEnvironment() *MMDeployEnvironment {
	env := &MMDeployEnvironment{
		KubernetesVersion:                      "1.17",
		OutputDir:                              "generated",
		RDS:                                    MMDeployEnvironment_RDS{},
		AWSLoadBalancerControllerIAMPolicyName: "",
		MattermostInstance: MMDeployEnvironment_MattermostInstance{
			PushServerUrl: "http://mattermost-push-proxy-svc:8066",
			ListenPort:    "8065",
		},
	}

	return env
}

// this is just a comment
func (m *MMDeployContext) LoadClusterConfig(conf string) error {
	fmt.Println("------ func (m *MMDeployContext) LoadClusterConfig(conf string) error")

	dat, err := ioutil.ReadFile(conf)

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	config, err := MMDeployEnvironmentFromJson(string(dat))

	dat2, err := MMDeployEnvironmentToJsonString(config)

	ioutil.WriteFile(os.Args[1], []byte(dat2), 0666)

	// d := json.NewDecoder(strings.NewReader(string(dat)))

	// d.UseNumber()

	// d.Decode(&config)

	if config.RDS.DBSecurityGroupName == "" {
		config.RDS.DBSecurityGroupName = config.ClusterName + "-dbaccess"
	}

	if config.RDS.DBInstanceName == "" {
		config.RDS.DBInstanceName = config.ClusterName
	}

	if config.AWSLoadBalancerControllerIAMPolicyName == "" {
		config.AWSLoadBalancerControllerIAMPolicyName = "AWSLoadBalancerControllerIAMPolicyName"
	}

	m.Context = config

	err, AWS_ACCESS_KEY_ID, stderr := aws_util.Execute("aws configure get aws_access_key_id", true, false)

	if err != nil {
		fmt.Println("stderr:", stderr)
		os.Exit(1)
	}

	err, AWS_ACCESS_SECRET, stderr := aws_util.Execute("aws configure get aws_secret_access_key", true, false)

	// err, stdout, stderr = aws.Execute("aws rds help", true, false)
	// log.Printf("%v %s %s", err, out, _err)

	fmt.Println("AWS_ACCESS_KEY_ID:", AWS_ACCESS_KEY_ID)
	fmt.Println("AWS_ACCESS_SECRET:", AWS_ACCESS_SECRET)

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(m.Context.Region),
		Credentials: credentials.NewSharedCredentials("", ""),
	})

	if err != nil {
		return err
	}

	m.Session = sess
	m.EC2 = ec2.New(sess)
	m.EKS = eks.New(sess)
	m.CF = cloudformation.New(sess)
	m.RDS = rds.New(sess)
	m.IAM = iam.New(sess)

	return nil
}

// this is just a comment
func (m *MMDeployContext) ProbeResources() error {
	fmt.Println("------ func (m *MMDeployContext) ProbeResources() error")

	eks_cluster, err := m.GetEKSCluster()

	if err != nil {
		return err
	}

	m.EKSCluster = eks_cluster

	iam_policy, err := m.GetAWSLoadBalancerControllerIAMPolicy()

	if err != nil {
		return err
	}

	if iam_policy != nil {
		m.Context.AWSLoadBalancerControllerIAMPolicyARN = *iam_policy.Arn
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
			if *db.DBInstanceIdentifier == m.Context.RDS.DBInstanceName {
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
		if *subnet_group.DBSubnetGroupName == m.Context.ClusterName {
			m.RDS.DeleteDBSubnetGroup(&rds.DeleteDBSubnetGroupInput{
				DBSubnetGroupName: aws.String(m.Context.ClusterName),
			})
		}

	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteCluster() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteCluster() error")

	fmt.Println("Deleting cluster:", m.Context.ClusterName)

	clusters, err := m.EKS.ListClusters(nil)

	if err != nil {
		return err
	}

	fmt.Println("clusters:", clusters)

	var cluster *eks.Cluster

	for _, _cluster := range clusters.Clusters {
		if *_cluster == m.Context.ClusterName {
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
			} else if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.Context.ClusterName {
				in_this_cluster = true
			}
		}

		if in_this_cluster && process_stack {
			fmt.Println("Will process stack:", stack)

			// stack.Stacks[0].Outputs
		}

		roles, err := _iam.ListRoles(nil)

		if err != nil {
			return err
		}

		// fmt.Println("roles:", roles)

		for _, role := range roles.Roles {
			in_this_cluster := false
			is_service_account := false

			role, err := _iam.GetRole(&iam.GetRoleInput{
				RoleName: role.RoleName,
			})

			// fmt.Println("role:", role)

			if err != nil {
				return err
			}

			for _, _tag := range role.Role.Tags {
				if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.Context.ClusterName {
					in_this_cluster = true
				} else if *_tag.Key == "alpha.eksctl.io/iamserviceaccount-name" && *_tag.Value == "kube-system/aws-load-balancer-controller" {
					is_service_account = true
				}
			}

			if in_this_cluster && is_service_account {
				fmt.Println("Deleting role:", role)

				detach_output, err := _iam.DetachRolePolicy(&iam.DetachRolePolicyInput{
					PolicyArn: role.Role.Arn,
				})

				if err != nil {
					return err
				}

				fmt.Println("detach_output:", detach_output)

				delete_output, err := _iam.DeleteRole(&iam.DeleteRoleInput{
					RoleName: role.Role.RoleName,
				})

				if err != nil {
					return err
				}

				fmt.Println("delete_output:", delete_output)
			}
		}
	}

	delete_cluster_result := make(chan string, 1)
	if stack != nil {
		go func() {
			aws_util.Execute(fmt.Sprintf("eksctl --region %v delete cluster --name %v -w", m.Context.Region, m.Context.ClusterName), true, true)
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
				if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.Context.ClusterName {
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
func (m *MMDeployContext) DeleteClusterVPC() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteClusterVPC() error")

	vpcs, err := m.EC2.DescribeVpcs(nil)

	if err != nil {
		return err
	}

	for _, vpc := range vpcs.Vpcs {
		in_this_cluster := false

		for _, tag := range vpc.Tags {
			if *tag.Key == "alpha.eksctl.io/cluster-name" && *tag.Value == m.Context.ClusterName {
				in_this_cluster = true
			}
		}

		if in_this_cluster {
			m.DeleteClusterVPCSecurityGroups(*vpc.VpcId)

			fmt.Println("Deleting VPC:", *vpc.VpcId)

			_, err := m.EC2.DeleteVpc(&ec2.DeleteVpcInput{
				VpcId: vpc.VpcId,
			})

			if err != nil {
				return err
			}
		}
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) EKSCreateCluster() error {
	fmt.Println("------ func (m *MMDeployContext) CreateCluster(result chan string) error")

	stack, err := m.GetCloudFormationMainStack()

	if err != nil {
		aws_util.ExitErrorf("Error getting stack, %v", err)
	}

	if stack != nil {
		return errors.New("CloudFormation stack already created")
	}

	cmd := []string{fmt.Sprintf("eksctl create cluster --name %v --fargate", m.Context.ClusterName)}

	if m.Context.Region != "" {
		cmd = append(cmd, fmt.Sprintf("--region %v", m.Context.Region))
	}

	if m.Context.KubernetesVersion != "" {
		cmd = append(cmd, fmt.Sprintf("--version %v", m.Context.KubernetesVersion))
	}

	if len(m.Context.AvailabilityZones) > 0 {
		cmd = append(cmd, fmt.Sprintf("--zones %v", strings.Join(m.Context.AvailabilityZones, ",")))
	}

	aws_util.Execute(strings.Join(cmd, " "), true, true)

	err, out1, out2 := aws_util.Execute("kubectl config current-context", true, true)

	fmt.Println("err:", err)
	fmt.Println("out1:", out1)
	fmt.Println("out2:", out2)

	os.Setenv("THIS_ENV_SET_BY_GO", "The quick brown fox jumps over the lazy dog")

	err, out1, out2 = aws_util.Execute("go run check_env.go", true, true)

	fmt.Println("err:", err)
	fmt.Println("out1:", out1)
	fmt.Println("out2:", out2)

	return nil
}

// this is just a comment
func (m *MMDeployContext) CreateApplicationLoadBalancer(result chan string) error {
	fmt.Println("------ func (m *MMDeployContext) CreateApplicationLoadBalancer(result chan string) error")

	fg_profiles, err := m.EKS.ListFargateProfiles(&eks.ListFargateProfilesInput{
		ClusterName: &m.Context.ClusterName,
	})

	if err != nil {
		return err
	}

	fmt.Println("fg_profiles:", fg_profiles)

	cert_manager_found := false

	for _, _fg_profile := range fg_profiles.FargateProfileNames {
		fg_profile, err := m.EKS.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
			ClusterName:        &m.Context.ClusterName,
			FargateProfileName: _fg_profile,
		})

		if err != nil {
			return err
		}

		// fmt.Println("fg_profile:", fg_profile)

		for _, sel := range fg_profile.FargateProfile.Selectors {
			if *sel.Namespace == "cert-manager" {
				cert_manager_found = true
			}
		}
	}

	if !cert_manager_found {
		create_fargate_profile_result, err := m.EKS.CreateFargateProfile(&eks.CreateFargateProfileInput{
			ClusterName:        &m.Context.ClusterName,
			FargateProfileName: aws.String("cert-manager"),
			Tags: aws.StringMap(map[string]string{
				"alpha.eksctl.io/cluster-name": m.Context.ClusterName,
			}),
		})

		if err != nil {
			return err
		}

		fmt.Println("create_fargate_profile_result:", create_fargate_profile_result)
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteDBSecurityGroup() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteDBSecurityGroup() error")

	return nil
}

// this is just a comment
func (m *MMDeployContext) GenerateEnvConfig() error {
	fmt.Println("------ func (m *MMDeployContext) GenerateEnvConfig() error")

	envConfig := &stage2.GenerateDeployEnvConfig{
		AWS_ACCESS_KEY_ID:                      "",
		AWS_SECRET_ACCESS_KEY:                  "",
		AWS_EKS_CLUSTER_NAME:                   m.Context.ClusterName,
		AWS_VPC_ID:                             "",
		AWS_REGION:                             m.Context.Region,
		DEPLOY_BUCKET:                          "",
		EKS_PUBLIC_SUBNETS:                     "",
		MM_DB_HOST:                             "",
		MM_DB_MASTER_USER:                      "",
		MM_DB_MASTER_PASS:                      "",
		NGINX_CONFIG_VERSION:                   "v1",
		MM_DEPLOY_VERSION:                      "v1",
		MM_CONF_PLUGIN_ENABLE_UPLOAD:           "true",
		SMTP_USER:                              "",
		SMTP_PASS:                              "",
		SMTP_HOST:                              "",
		SMTP_PORT:                              "",
		SMTP_FROM:                              "",
		MM_PROXY_PROXY_CONFIG_VERSION:          "v1",
		MATTERMOST_PUSH_NOTIFICATION_URL:       "",
		MATTERMOST_PUSH_PROXY_DOCKER_REPO:      "",
		MM_DOCKER_REPO:                         "",
		MM_CLUSTER_DRIVER:                      "redis",
		MM_CLUSTER_REDIS_HOST:                  "",
		MM_CLUSTER_REDIS_PORT:                  "",
		MM_CLUSTER_REDIS_PASS:                  "",
		VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME: "",
	}

	b, err := stage2.DeployEnvConfigToJsonString(envConfig)

	fmt.Println("b:", b)

	if err != nil {
		return err
	}

	ioutil.WriteFile("../stage2/env.json", []byte(b), 0666)

	if err != nil {
		return err
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) DeleteFargateProfiles(result chan string) error {
	fmt.Println("------ func (m *MMDeployContext) DeleteFargateProfiles(result chan string) error")

	for {
		fg_profiles, err := m.EKS.ListFargateProfiles(&eks.ListFargateProfilesInput{
			ClusterName: &m.Context.ClusterName,
		})

		if err != nil {
			return err
		}

		fmt.Println("fg_profiles:", fg_profiles)

		deleting_count := 0

		for _, _fg_profile := range fg_profiles.FargateProfileNames {
			fg_profile, err := m.EKS.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
				ClusterName:        &m.Context.ClusterName,
				FargateProfileName: _fg_profile,
			})

			if err != nil {
				return err
			}

			// fmt.Println("fg_profile:", fg_profile)

			if *fg_profile.FargateProfile.Status == "DELETING" {
				deleting_count = deleting_count + 1
			}
		}

		fmt.Println("deleting_count:", deleting_count)

		if deleting_count == 0 {
			for _, _fg_profile := range fg_profiles.FargateProfileNames {
				fg_profile, err := m.EKS.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
					ClusterName:        &m.Context.ClusterName,
					FargateProfileName: _fg_profile,
				})

				if err != nil {
					return err
				}

				// fmt.Println("fg_profile:", fg_profile)

				if *fg_profile.FargateProfile.Status == "ACTIVE" {
					_, err := m.EKS.DeleteFargateProfile(&eks.DeleteFargateProfileInput{
						ClusterName:        &m.Context.ClusterName,
						FargateProfileName: fg_profile.FargateProfile.FargateProfileName,
					})

					// fmt.Println("fg_delete_output:", fg_delete_output)

					if err != nil {
						return err
					}

					break
				}
			}
		}

		fmt.Println("len(fg_profiles.FargateProfileNames):", len(fg_profiles.FargateProfileNames))

		if len(fg_profiles.FargateProfileNames) == 0 {
			break
		} else {
			time.Sleep(time.Second * 10)
			continue
		}
	}

	result <- "done"

	return nil
}

func main() {
	var operation, config_file string

	if len(os.Args) > 1 {
		config_file = os.Args[1]
	}

	if len(os.Args) > 2 {
		operation = os.Args[2]
	}

	if config_file == "" {
		fmt.Println("Missing required parameter configuration file.")
		os.Exit(1)
	}

	if operation == "" {
		os.Setenv("THIS_ENV_SET_BY_GO", "The quick brown fox jumps over the lazy dog")

		err, out1, out2 := aws_util.Execute(fmt.Sprintf("./stage1 %v check_env", config_file), true, true)

		fmt.Println("err:", err)
		fmt.Println("out1:", out1)
		fmt.Println("out2:", out2)
	} else if operation == "check_env" {
		fmt.Println("THIS_ENV_SET_BY_GO:", os.Getenv("THIS_ENV_SET_BY_GO"))
	} else {
		mm_eks_env := &MMDeployContext{}

		mm_eks_env.LoadClusterConfig(config_file)

		fmt.Println("domains:", mm_eks_env, "operation:", operation)

		mm_eks_env.ProbeResources()

		if operation == "delete_cluster" {
			mm_eks_env.DeleteCluster()
			mm_eks_env.DeleteClusterVPC()
			mm_eks_env.DeleteOtherStacks()
		} else if operation == "create_cluster" {
			err := mm_eks_env.EKSCreateCluster()

			if err != nil {
				aws_util.ExitErrorf("Unable to create cluster, %v", err)
			}
		} else if operation == "fix_missing" {
			err := mm_eks_env.FixMissing()

			if err != nil {
				aws_util.ExitErrorf("Unable to create cluster, %v", err)
			}
		} else if operation == "generate_config_env" {
			err := mm_eks_env.GenerateEnvConfig()

			if err != nil {
				aws_util.ExitErrorf("Unable to create cluster, %v", err)
			}
		}
	}
}
