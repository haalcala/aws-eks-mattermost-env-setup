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
	Subnets                                []string                               `json:"Subnets"`
	PrivateSubnets                         []string                               `json:"PrivateSubnets"`
	PublicSubnets                          []string                               `json:"PublicSubnets"`
	VpcId                                  string                                 `json:"VpcId"`
	Region                                 string                                 `json:"Region"`
	KubernetesVersion                      string                                 `json:"KubernetesVersion"`
	AWSLoadBalancerControllerIAMPolicyName string                                 `json:"AWSLoadBalancerControllerIAMPolicyName"`
	AWSLoadBalancerControllerIAMPolicyARN  string                                 `json:"AWSLoadBalancerControllerIAMPolicyARN"`
	RDS                                    MMDeployEnvironment_RDS                `json:"RDS"`
	MattermostInstance                     MMDeployEnvironment_MattermostInstance `json:"MattermostInstance"`
	InfraComponents                        MMDeployEnvironment_InfraComponents    `json:"InfraComponents"`
	OutputDir                              string                                 `json:"OutputDir"`
	AWSCredentialProfile                   string                                 `json:"AWSCredentialProfile"`
	DeployBucket                           string                                 `json:"DeployBucket"`
	Containers                             MMDeployEnvironment_Containers         `json:"Containers"`
	VcubeOauth                             MMDeployEnvironment_VcubeOauth         `json:"VcubeOauth"`
	ImportBucketRegion                     string                                 `json:"ImportBucketRegion"`
	ImportBucket                           string                                 `json:"ImportBucket"`
	KubernetesContext                      string                                 `json:"KubernetesContext"`
}

type MMDeployEnvironment_VcubeOauth struct {
	InitialAdminUser  string `json:"InitialAdminUser"`
	InitialAdminPass  string `json:"InitialAdminPass"`
	SessionSecret     string `json:"SessionSecret"`
	VMeetingUrl       string `json:"VMeetingUrl"`
	VIDConsumerKey    string `json:"VIDConsumerKey"`
	VIDRestPwd        string `json:"VIDRestPwd"`
	VIDRestUrl        string `json:"VIDRestUrl"`
	VIDSecretAuthCode string `json:"VIDSecretAuthCode"`
	DBName            string `json:"DBName"`
	DBPort            string `json:"DBPort"`
	DBUser            string `json:"DBUser"`
	DBPass            string `json:"DBPass"`
}

type MMDeployEnvironment_Containers struct {
	Mattermost     MMDeployEnvironment_Container `json:"Mattermost"`
	PushProxy      MMDeployEnvironment_Container `json:"PushProxy"`
	Nginx          MMDeployEnvironment_Container `json:"Nginx"`
	JaegerQuery    MMDeployEnvironment_Container `json:"JaegerQuery"`
	JaegeCollector MMDeployEnvironment_Container `json:"JaegeCollector"`
	Redis          MMDeployEnvironment_Container `json:"Redis"`
	VcubeOauth     MMDeployEnvironment_Container `json:"VcubeOauth"`
}

type MMDeployEnvironment_Container struct {
	ImageName string `json:"ImageName"`
	Repo      string `json:"Repo"`
	Version   string `json:"Version"`
}

type MMDeployEnvironment_InfraComponents struct {
	PushProxy struct {
		URL string `json:"URL"`
	} `json:"PushProxy"`
}

type MMDeployEnvironment_RDS struct {
	RDSDeployAZ         string `json:"RDSDeployAZ"`
	DBSecurityGroupName string `json:"DBSecurityGroupName"`
	DBInstanceName      string `json:"DBInstanceName"`
}

type MMDeployEnvironment_MattermostInstance struct {
	Cluster       MMDeployEnvironment_MattermostInstance_Cluster `json:"Cluster"`
	PushServerUrl string                                         `json:"PushServerUrl"`
	DBHost        string                                         `json:"DBHost"`
	DBPort        string                                         `json:"DBPort"`
	DBUser        string                                         `json:"DBUser"`
	DBPass        string                                         `json:"DBPass"`
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
	From string `json:"From"`
}

// MMDeployContext bla bla bla
type MMDeployContext struct {
	DeployConfig   MMDeployEnvironment
	Session        *session.Session
	EKSCluster     *eks.Cluster
	EC2            *ec2.EC2
	EKS            *eks.EKS
	CF             *cloudformation.CloudFormation
	IAM            *iam.IAM
	RDS            *rds.RDS
	ConfigFile     string
	Subnets        []*ec2.Subnet
	PrivateSubnets []*ec2.Subnet
	PublicSubnets  []*ec2.Subnet
}

func (c *MMDeployEnvironment) MMDeployConfigToJsonString() (string, error) {
	b, err := json.MarshalIndent(c, "", "\t")

	return string(b), err
}

func (c *MMDeployEnvironment) ApplyDefaults() {
	if c.MattermostInstance.DBPort == "" {
		c.MattermostInstance.DBPort = "3306"
	}

	if c.MattermostInstance.ListenPort == "" {
		c.MattermostInstance.ListenPort = "8065"
	}

	if c.MattermostInstance.PushServerUrl == "" {
		c.MattermostInstance.PushServerUrl = "http://mattermost-push-proxy-svc:8066"
	}

	if c.MattermostInstance.Cluster.Driver == "" {
		c.MattermostInstance.Cluster.Driver = "redis"
	}
	if c.MattermostInstance.Cluster.CustomClusterRedisHost == "" {
		c.MattermostInstance.Cluster.CustomClusterRedisHost = "svc-redis"
	}
	if c.MattermostInstance.Cluster.CustomClusterRedisPort == "" {
		c.MattermostInstance.Cluster.CustomClusterRedisPort = "6357"
	}
	if c.OutputDir == "" {
		c.OutputDir = "generated_config"
	}

	getDefaultContainerProps := func(container *MMDeployEnvironment_Container, name, repo string) {
		if container.ImageName == "" {
			container.ImageName = name
		}

		if container.Repo == "" {
			container.Repo = repo
		}

		if container.Version == "" {
			container.Version = "v1"
		}
	}

	getDefaultContainerProps(&c.Containers.Mattermost, "mattermost", "946808171471.dkr.ecr.ap-northeast-1.amazonaws.com/mattermost-prod")
	getDefaultContainerProps(&c.Containers.PushProxy, "pushproxy", "946808171471.dkr.ecr.ap-northeast-1.amazonaws.com/mattermost-push-proxy")
	getDefaultContainerProps(&c.Containers.JaegeCollector, "jaeger-collector", "")
	getDefaultContainerProps(&c.Containers.JaegerQuery, "jaeger-query", "")
	getDefaultContainerProps(&c.Containers.Nginx, "nginx", "")
	getDefaultContainerProps(&c.Containers.VcubeOauth, "vcube-oauth", "946808171471.dkr.ecr.ap-northeast-1.amazonaws.com/vmeeting-oauth2-wrapper:harold-test")
}

func MMDeployConfigFromJson(_json string) (*MMDeployEnvironment, error) {
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
func (m *MMDeployContext) SaveDeployConfig() error {
	fmt.Println("------ func (m *MMDeployContext) SaveDeployConfig() error")

	dat2, err := m.DeployConfig.MMDeployConfigToJsonString()

	if err != nil {
		return err
	}

	ioutil.WriteFile(m.ConfigFile, []byte(dat2), 0666)

	return nil
}

// this is just a comment
func (m *MMDeployContext) LoadDeployConfig(conf string) error {
	fmt.Println("------ func (m *MMDeployContext) LoadDeployConfig(conf string) error")

	dat, err := ioutil.ReadFile(conf)

	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			fmt.Printf("Configuration file %v cannot be found. Generating configuration structure...\n", conf)

			m.DeployConfig.ApplyDefaults()

			err := m.SaveDeployConfig()

			if err != nil {
				fmt.Println("err:", err)
			}

			fmt.Println("Please try again.")
		} else {
			fmt.Println("err:", err)
		}
		os.Exit(1)
	}

	config, err := MMDeployConfigFromJson(string(dat))

	m.DeployConfig = *config

	config.ApplyDefaults()

	m.SaveDeployConfig()

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
		Region:      aws.String(m.DeployConfig.Region),
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
	fmt.Println("------ func (m *MMDeployContext) ProbeResources() error")

	eks_cluster, err := m.GetEKSCluster()

	if err != nil {
		return err
	}

	m.EKSCluster = eks_cluster

	m.DeployConfig.KubernetesVersion = *m.EKSCluster.Version

	if m.DeployConfig.KubernetesContext == "" {
		err, stdout, stderr := aws_util.Execute("kubectl config current-context", true, true)

		if err != nil {
			return err
		}

		fmt.Println("stdout:", stdout)
		fmt.Println("stderr:", stderr)

		m.DeployConfig.KubernetesContext = strings.Trim(stdout, "\r")
	}

	subnets, err := m.EC2.DescribeSubnets(&ec2.DescribeSubnetsInput{
		SubnetIds: m.EKSCluster.ResourcesVpcConfig.SubnetIds,
	})

	if err != nil {
		return err
	}

	public_subnets, private_subnets := []*ec2.Subnet{}, []*ec2.Subnet{}

	for _, subnet := range subnets.Subnets {
		for _, tag := range subnet.Tags {
			if *tag.Key == "kubernetes.io/role/elb" && *tag.Value == "1" {
				public_subnets = append(public_subnets, subnet)
			} else if *tag.Key == "kubernetes.io/role/internal-elb" && *tag.Value == "1" {
				private_subnets = append(private_subnets, subnet)
			}
		}
	}

	m.Subnets = append(public_subnets, private_subnets...)
	m.PrivateSubnets = private_subnets
	m.PublicSubnets = public_subnets

	iam_policy, err := m.GetAWSLoadBalancerControllerIAMPolicy()

	if err != nil {
		return err
	}

	if iam_policy != nil {
		m.DeployConfig.AWSLoadBalancerControllerIAMPolicyARN = *iam_policy.Arn
	}

	DBInstanceName_assumed := false

	if m.DeployConfig.RDS.DBInstanceName == "" {
		m.DeployConfig.RDS.DBInstanceName = m.DeployConfig.ClusterName
		DBInstanceName_assumed = true
	}

	dbs, err := m.RDS.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: &m.DeployConfig.RDS.DBInstanceName,
	})

	if err != nil {
		return err
	}

	if len(dbs.DBInstances) > 0 {
		db := dbs.DBInstances[0]

		m.DeployConfig.MattermostInstance.DBHost = *db.Endpoint.Address
		m.DeployConfig.MattermostInstance.DBPort = fmt.Sprintf("%v", *db.Endpoint.Port)
		m.DeployConfig.MattermostInstance.DBUser = *db.MasterUsername
	} else if DBInstanceName_assumed {
		m.DeployConfig.RDS.DBInstanceName = ""
	}

	return nil
}

// this is just a comment
func (m *MMDeployContext) PatchDeployConfig() error {
	fmt.Println("------ func (m *MMDeployContext) PatchDeployConfig() error")

	m.DeployConfig.VpcId = *m.EKSCluster.ResourcesVpcConfig.VpcId

	m.DeployConfig.Subnets = aws.StringValueSlice(m.EKSCluster.ResourcesVpcConfig.SubnetIds)

	m.DeployConfig.PublicSubnets = []string{}
	m.DeployConfig.PrivateSubnets = []string{}

	for _, subnet := range m.PublicSubnets {
		m.DeployConfig.PublicSubnets = append(m.DeployConfig.PublicSubnets, *subnet.SubnetId)
	}
	for _, subnet := range m.PrivateSubnets {
		m.DeployConfig.PrivateSubnets = append(m.DeployConfig.PrivateSubnets, *subnet.SubnetId)
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

// this is just a comment
func (m *MMDeployContext) DeleteCluster() error {
	fmt.Println("------ func (m *MMDeployContext) DeleteCluster() error")

	fmt.Println("Deleting cluster:", m.DeployConfig.ClusterName)

	clusters, err := m.EKS.ListClusters(nil)

	if err != nil {
		return err
	}

	fmt.Println("clusters:", clusters)

	var cluster *eks.Cluster

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
				if *_tag.Key == "alpha.eksctl.io/cluster-name" && *_tag.Value == m.DeployConfig.ClusterName {
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
			aws_util.Execute(fmt.Sprintf("eksctl --region %v delete cluster --name %v -w", m.DeployConfig.Region, m.DeployConfig.ClusterName), true, true)
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
			if *tag.Key == "alpha.eksctl.io/cluster-name" && *tag.Value == m.DeployConfig.ClusterName {
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

	cmd := []string{fmt.Sprintf("eksctl create cluster --name %v --fargate", m.DeployConfig.ClusterName)}

	if m.DeployConfig.Region != "" {
		cmd = append(cmd, fmt.Sprintf("--region %v", m.DeployConfig.Region))
	}

	if m.DeployConfig.KubernetesVersion != "" {
		cmd = append(cmd, fmt.Sprintf("--version %v", m.DeployConfig.KubernetesVersion))
	}

	if len(m.DeployConfig.AvailabilityZones) > 0 {
		cmd = append(cmd, fmt.Sprintf("--zones %v", strings.Join(m.DeployConfig.AvailabilityZones, ",")))
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
		ClusterName: &m.DeployConfig.ClusterName,
	})

	if err != nil {
		return err
	}

	fmt.Println("fg_profiles:", fg_profiles)

	cert_manager_found := false

	for _, _fg_profile := range fg_profiles.FargateProfileNames {
		fg_profile, err := m.EKS.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
			ClusterName:        &m.DeployConfig.ClusterName,
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
			ClusterName:        &m.DeployConfig.ClusterName,
			FargateProfileName: aws.String("cert-manager"),
			Tags: aws.StringMap(map[string]string{
				"alpha.eksctl.io/cluster-name": m.DeployConfig.ClusterName,
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
func (m *MMDeployContext) GenerateSaveEnvConfig() error {
	fmt.Println("------ func (m *MMDeployContext) GenerateSaveEnvConfig() error")

	cred, err := credentials.NewSharedCredentials("", m.DeployConfig.AWSCredentialProfile).Get()

	if err != nil {
		return err
	}

	envConfig := &stage2.DeploymentEnvironment{
		AWS_ACCESS_KEY_ID:                      cred.AccessKeyID,
		AWS_SECRET_ACCESS_KEY:                  cred.SecretAccessKey,
		AWS_EKS_CLUSTER_NAME:                   m.DeployConfig.ClusterName,
		AWS_VPC_ID:                             m.DeployConfig.VpcId,
		AWS_REGION:                             m.DeployConfig.Region,
		DEPLOY_BUCKET:                          m.DeployConfig.DeployBucket,
		EKS_PUBLIC_SUBNETS:                     strings.Join(m.DeployConfig.PublicSubnets, ","),
		MM_DB_HOST:                             m.DeployConfig.MattermostInstance.DBHost,
		MM_DB_PORT:                             m.DeployConfig.MattermostInstance.DBPort,
		MM_DB_MASTER_USER:                      m.DeployConfig.MattermostInstance.DBUser,
		MM_DB_MASTER_PASS:                      m.DeployConfig.MattermostInstance.DBPass,
		NGINX_CONFIG_VERSION:                   m.DeployConfig.Containers.Nginx.Version,
		MM_DEPLOY_VERSION:                      m.DeployConfig.Containers.Mattermost.Version,
		MM_CONF_PLUGIN_ENABLE_UPLOAD:           "true",
		SMTP_USER:                              m.DeployConfig.MattermostInstance.SMTP.User,
		SMTP_PASS:                              m.DeployConfig.MattermostInstance.SMTP.Pass,
		SMTP_HOST:                              m.DeployConfig.MattermostInstance.SMTP.Host,
		SMTP_PORT:                              m.DeployConfig.MattermostInstance.SMTP.Port,
		SMTP_FROM:                              m.DeployConfig.MattermostInstance.SMTP.From,
		MM_PROXY_PROXY_CONFIG_VERSION:          m.DeployConfig.Containers.PushProxy.Version,
		MATTERMOST_PUSH_NOTIFICATION_URL:       m.DeployConfig.MattermostInstance.PushServerUrl,
		MATTERMOST_PUSH_PROXY_DOCKER_REPO:      m.DeployConfig.Containers.PushProxy.Repo,
		MM_DOCKER_REPO:                         m.DeployConfig.Containers.Mattermost.Repo,
		MM_CLUSTER_DRIVER:                      m.DeployConfig.MattermostInstance.Cluster.Driver,
		MM_CLUSTER_REDIS_HOST:                  m.DeployConfig.MattermostInstance.Cluster.CustomClusterRedisHost,
		MM_CLUSTER_REDIS_PORT:                  m.DeployConfig.MattermostInstance.Cluster.CustomClusterRedisPort,
		MM_CLUSTER_REDIS_PASS:                  m.DeployConfig.MattermostInstance.Cluster.CustomClusterRedisPass,
		VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME: m.DeployConfig.VcubeOauth.InitialAdminUser,
		VCUBE_VID_OAUTH_INITIAL_ADMIN_PASSWORD: m.DeployConfig.VcubeOauth.InitialAdminPass,
		VCUBE_VID_OAUTH_EXPRESS_SESSION_SECRET: m.DeployConfig.VcubeOauth.SessionSecret,
		VCUBE_VID_OAUTH_VMEETING_URL:           m.DeployConfig.VcubeOauth.VMeetingUrl,
		VCUBE_VID_OAUTH_VID_CONSUMER_KEY:       m.DeployConfig.VcubeOauth.VIDConsumerKey,
		VCUBE_VID_OAUTH_VID_REST_PWD:           m.DeployConfig.VcubeOauth.VIDRestPwd,
		VCUBE_VID_OAUTH_VID_REST_URL:           m.DeployConfig.VcubeOauth.VIDRestUrl,
		VCUBE_VID_OAUTH_VID_SECRET_AUTH_CODE:   m.DeployConfig.VcubeOauth.VIDSecretAuthCode,
		VCUBE_VID_OAUTH_CONTAINER_VERSION:      m.DeployConfig.Containers.VcubeOauth.Version,
		VCUBE_VID_OAUTH_CONTAINER_REPO:         m.DeployConfig.Containers.VcubeOauth.Repo,
		VCUBE_VID_OAUTH_DB_NAME:                m.DeployConfig.VcubeOauth.DBName,
		VCUBE_VID_OAUTH_DB_USERNAME:            m.DeployConfig.VcubeOauth.DBUser,
		VCUBE_VID_OAUTH_DB_PASSWORD:            m.DeployConfig.VcubeOauth.DBPass,
		MATTERMOST_PORT:                        m.DeployConfig.MattermostInstance.ListenPort,
		AWS_PROD_S3_ACCESS_KEY_ID:              cred.AccessKeyID,
		AWS_PROD_S3_SECRET_ACCESS_KEY:          cred.SecretAccessKey,
		IMPORT_EXTERNAL_BUCKET_REGION:          m.DeployConfig.ImportBucketRegion,
		IMPORT_EXTERNAL_BUCKET:                 m.DeployConfig.ImportBucket,
		AWS_ACM_CERTIFICATE_ARN:                m.DeployConfig.AWSLoadBalancerControllerIAMPolicyARN,
	}

	b, err := stage2.DeploymentEnvironmentToJsonString(envConfig)

	fmt.Println("b:", b)

	if err != nil {
		return err
	}

	ioutil.WriteFile("env.json", []byte(b), 0666)

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
			ClusterName: &m.DeployConfig.ClusterName,
		})

		if err != nil {
			return err
		}

		fmt.Println("fg_profiles:", fg_profiles)

		deleting_count := 0

		for _, _fg_profile := range fg_profiles.FargateProfileNames {
			fg_profile, err := m.EKS.DescribeFargateProfile(&eks.DescribeFargateProfileInput{
				ClusterName:        &m.DeployConfig.ClusterName,
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
					ClusterName:        &m.DeployConfig.ClusterName,
					FargateProfileName: _fg_profile,
				})

				if err != nil {
					return err
				}

				// fmt.Println("fg_profile:", fg_profile)

				if *fg_profile.FargateProfile.Status == "ACTIVE" {
					_, err := m.EKS.DeleteFargateProfile(&eks.DeleteFargateProfileInput{
						ClusterName:        &m.DeployConfig.ClusterName,
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
		mm_eks_env := &MMDeployContext{ConfigFile: config_file}

		mm_eks_env.LoadDeployConfig(config_file)

		fmt.Println("domains:", mm_eks_env, "operation:", operation)

		if mm_eks_env.DeployConfig.ClusterName == "" {
			fmt.Println("Please supply 'ClusterName' in the config file.")
			os.Exit(1)
		}

		if mm_eks_env.DeployConfig.Region == "" {
			fmt.Println("Please supply 'Region' in the config file.")
			os.Exit(1)
		}

		mm_eks_env.ProbeResources()
		mm_eks_env.PatchDeployConfig()
		mm_eks_env.SaveDeployConfig()

		fmt.Println("mm_eks_env.DeployConfig.VpcId:", mm_eks_env.DeployConfig.VpcId)

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
			err := mm_eks_env.GenerateSaveEnvConfig()

			if err != nil {
				aws_util.ExitErrorf("Unable to create cluster, %v", err)
			}
		} else if operation == "generate_deployment" {
			dat, err := ioutil.ReadFile("env.json")

			if err != nil {
				fmt.Println("err:", err)
				os.Exit(1)
			}

			props := &map[string]string{}

			err = json.Unmarshal(dat, props)

			// fmt.Println("props:", props)

			for key, val := range *props {
				// fmt.Println("key:", key, "val:", val)

				os.Setenv("__"+key+"__", val)
			}

			baseDir := mm_eks_env.DeployConfig.OutputDir

			if baseDir != "" {
				err := os.Mkdir(baseDir, 0777)

				if err != nil && !strings.Contains(err.Error(), "file exists") {
					aws_util.ExitErrorf("Unable to create folder, %v", err)
				}
			}

			stage2.GenerateDeploymentFiles(baseDir)
		}
	}
}
