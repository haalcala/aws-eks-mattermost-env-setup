package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/route53"
)

func (c *MMDeployEnvironment) MMDeployConfigToJsonString() (string, error) {
	b, err := json.MarshalIndent(c, "", "\t")

	return string(b), err
}

func (c *MMDeployEnvironment) ApplyDefaults() {
	if c.KubernetesEnvironment == "" {
		c.KubernetesEnvironment = "eks"
	}

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

	if c.AWSLoadBalancerControllerName == "" {
		c.AWSLoadBalancerControllerName = "alb-ingress-controller"
	}

	if c.AWSLoadBalancerControllerIAMPolicyName == "" {
		c.AWSLoadBalancerControllerIAMPolicyName = "ALBIngressControllerIAMPolicy"
	}

	if c.RDS.DBSecurityGroupName == "" {
		c.RDS.DBSecurityGroupName = c.ClusterName + "-dbaccess"
	}

	if c.RDS.DBInstanceName == "" {
		c.RDS.DBInstanceName = c.ClusterName
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
		RDS:                                    MMDeployEnvironment_RDS{},
		AWSLoadBalancerControllerIAMPolicyName: "",
		MattermostInstance: MMDeployEnvironment_MattermostInstance{
			PushServerUrl: "http://mattermost-push-proxy-svc.default.svc.cluster.local:8066",
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
				return err
			}

			fmt.Println("Please try again.")
		}

		return err
	}

	config, err := MMDeployConfigFromJson(string(dat))

	m.DeployConfig = *config

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(m.DeployConfig.Region),
		Credentials: credentials.NewSharedCredentials("", ""),
	})

	if err != nil {
		return err
	}

	fmt.Println("sess:", sess)

	m.Session = sess
	m.EC2 = ec2.New(sess)
	m.EKS = eks.New(sess)
	m.CF = cloudformation.New(sess)
	m.RDS = rds.New(sess)
	m.IAM = iam.New(sess)
	m.ELB = elbv2.New(sess)
	m.R53 = route53.New(sess)

	err = m.LoadDomains()
	if err != nil {
		return err
	}

	config.ApplyDefaults()

	err = m.SaveDeployConfig()
	if err != nil {
		return err
	}
	// d := json.NewDecoder(strings.NewReader(string(dat)))

	// d.UseNumber()

	// d.Decode(&config)

	err, AWS_ACCESS_KEY_ID, _ := Execute("aws configure get aws_access_key_id", true, false)

	if err != nil {
		return err
	}

	err, AWS_ACCESS_SECRET, _ := Execute("aws configure get aws_secret_access_key", true, false)

	if err != nil {
		return err
	}

	// err, stdout, stderr = aws.Execute("aws rds help", true, false)
	// log.Printf("%v %s %s", err, out, _err)

	fmt.Println("AWS_ACCESS_KEY_ID:", AWS_ACCESS_KEY_ID)
	fmt.Println("AWS_ACCESS_SECRET:", AWS_ACCESS_SECRET)

	return nil
}

// this is just a comment
func (m *MMDeployContext) ValidateManualParameters() error {
	fmt.Println("------ func (m *MMDeployContext) ValidateManualParameters() error")

	if m.DeployConfig.AWSCertificateARN == "" {
		return errors.New("Please supply 'AWSCertificateARN' in the config file.")
	}

	if m.DeployConfig.ClusterName == "" {
		return errors.New("Please supply 'ClusterName' in the config file.")
	}

	if m.DeployConfig.Region == "" {
		return errors.New("Please supply 'Region' in the config file.")
	}

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

	m.DeployConfig.KubernetesVersion = *m.EKSCluster.Version

	m.DeployConfig.ApplyDefaults()

	if m.DeployConfig.KubernetesContext == "" {
		err, stdout, stderr := Execute("kubectl config current-context", true, true)

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

	if m.EKSCluster != nil {
		m.DeployConfig.VpcId = *m.EKSCluster.ResourcesVpcConfig.VpcId

		m.DeployConfig.Subnets = aws.StringValueSlice(m.EKSCluster.ResourcesVpcConfig.SubnetIds)
	}

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
func (m *MMDeployContext) EKSCreateCluster() error {
	fmt.Println("------ func (m *MMDeployContext) CreateCluster(result chan string) error")

	stack, err := m.GetCloudFormationMainStack()

	if err != nil {
		ExitErrorf("Error getting stack, %v", err)
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

	Execute(strings.Join(cmd, " "), true, true)

	err, out1, out2 := Execute("kubectl config current-context", true, true)

	fmt.Println("err:", err)
	fmt.Println("out1:", out1)
	fmt.Println("out2:", out2)

	os.Setenv("THIS_ENV_SET_BY_GO", "The quick brown fox jumps over the lazy dog")

	err, out1, out2 = Execute("go run check_env.go", true, true)

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
func (m *MMDeployContext) GenerateSaveEnvConfig(baseDir string) error {
	fmt.Println("------ func (m *MMDeployContext) GenerateSaveEnvConfig() error")

	cred, err := credentials.NewSharedCredentials("", m.DeployConfig.AWSCredentialProfile).Get()

	if err != nil {
		return err
	}

	envConfig := &DeploymentEnvironment{
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
		AWS_ACM_CERTIFICATE_ARN:                m.DeployConfig.AWSCertificateARN,
		ALB_INGRESS_CONTROLLER_NAME:            m.DeployConfig.AWSLoadBalancerControllerName,
		ALB_INGRESS_CONTROLLER_IAM_POLICY:      m.DeployConfig.AWSLoadBalancerControllerIAMPolicyName,
		ALB_INGRESS_CONTROLLER_IAM_POLICY_ARN:  m.DeployConfig.AWSLoadBalancerControllerIAMPolicyARN,
	}

	b, err := DeploymentEnvironmentToJsonString(envConfig)

	fmt.Println("b:", b)

	if err != nil {
		return err
	}

	fmt.Println("Writing environment settings to", baseDir+"/env.json")

	err = ioutil.WriteFile(baseDir+"/env.json", []byte(b), 0666)

	if err != nil {
		return err
	}

	dat, err := ioutil.ReadFile(baseDir + "/env.json")

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	props := &map[string]string{}

	err = json.Unmarshal(dat, props)

	env := []string{}

	for key, val := range *props {
		env = append(env, "export __"+key+"__=\""+val+"\"")
	}

	fmt.Println("Writing environment variables to", baseDir+"/env.sh")

	err = ioutil.WriteFile(baseDir+"/env.sh", []byte(strings.Join(env, "\n")), 0666)

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
	var operation, cluster_name string

	if len(os.Args) > 1 {
		cluster_name = os.Args[1]
	}

	if len(os.Args) > 2 {
		operation = os.Args[2]
	}

	if cluster_name == "" {
		fmt.Println("Missing required parameter configuration file.")
		os.Exit(1)
	}

	config_file := cluster_name + "-conf.json"

	deploy_context := &MMDeployContext{ConfigFile: config_file, DomainsFile: cluster_name + "-conf-domains.json"}

	var baseDir string

	apis := map[string]func() error{
		"create_cluster": func() error {
			return deploy_context.EKSCreateCluster()
		},
		"delete_cluster": func() error {
			err := deploy_context.DeleteCluster()
			if err != nil {
				return err
			}

			err = deploy_context.DeleteClusterVPC()
			if err != nil {
				return err
			}

			err = deploy_context.DeleteOtherStacks()
			if err != nil {
				return err
			}

			return nil
		},
		"fix_missing": func() error {
			return deploy_context.FixMissing()
		},
		"generate_config_env": func() error {
			return deploy_context.GenerateSaveEnvConfig(baseDir)
		},
		"generate_deployment": func() error {
			dat, err := ioutil.ReadFile(baseDir + "/env.json")

			if err != nil {
				return err
			}

			props := &map[string]string{}

			err = json.Unmarshal(dat, props)

			// fmt.Println("props:", props)

			for key, val := range *props {
				// fmt.Println("key:", key, "val:", val)

				os.Setenv("__"+key+"__", val)
			}

			return deploy_context.GenerateDeploymentFiles(baseDir)
		},
	}

	handler := apis[operation]

	if handler == nil {
		fmt.Println("Unrecognised operation:", operation, "\n")

		fmt.Println("eksmmctl <cluster name> <operation>\n")
		fmt.Println("Available operations are:\n")
		fmt.Println("\tcreate_cluster")
		fmt.Println("\tdelete_cluster")
		fmt.Println("\tfix_missing")
		fmt.Println("\tgenerate_config_env")
		fmt.Println("\tgenerate_deployment")
	} else {
		err := deploy_context.LoadDeployConfig(config_file)
		if err != nil {
			ExitErrorf("Invalid configration, %v", err)
		}

		err = deploy_context.ValidateManualParameters()
		if err != nil {
			ExitErrorf("Invalid configration, %v", err)
		}

		baseDir = "generated_deployments/" + deploy_context.DeployConfig.ClusterName

		if baseDir != "" {
			err := os.MkdirAll(baseDir, 0777)

			if err != nil && !strings.Contains(err.Error(), "file exists") {
				ExitErrorf("Unable to create folder, %v", err)
			}
		}

		deploy_context.ProbeResources()
		deploy_context.PatchDeployConfig()
		deploy_context.SaveDeployConfig()

		fmt.Println("deploy_context.DeployConfig.VpcId:", deploy_context.DeployConfig.VpcId)

		err = handler()

		if err != nil {
			ExitErrorf("Unable to create cluster, %v", err)
		}
	}
}
