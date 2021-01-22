package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/eks"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/route53"
)

const (
	GENERATED_DEPLOYMENT_BASE = "generated_deployments"
)

type MMDeployEnvironment struct {
	// The EKS cluster name
	ClusterName string `json:"ClusterName"`

	// The availability zone to which the cluster is going to be available when it is created.
	AvailabilityZones []string `json:"AvailabilityZones"`

	// The subnets to which the cluster is going to be available when it is created.  This is probed from existing cluster if existing.
	Subnets []string `json:"Subnets"`

	// This is for convenience access only.  This is probed from existing cluster if existing.
	PrivateSubnets []string `json:"PrivateSubnets"`
	// This is for convenience access only.  This is probed from existing cluster if existing.
	PublicSubnets []string `json:"PublicSubnets"`

	// The VPC to which the cluster is going to be available when it is created.  This is probed from existing cluster if existing.
	VpcId string `json:"VpcId"`

	// The AWS region to which the cluster is going to be available when it is created.  This is probed from existing cluster if existing.
	Region string `json:"Region"`

	// The Kubernetes version to be used when it is created.  This is probed from existing cluster if existing.
	KubernetesVersion string `json:"KubernetesVersion"`

	AWSLoadBalancerControllerName          string `json:"AWSLoadBalancerControllerName"`
	AWSLoadBalancerControllerIAMPolicyName string `json:"AWSLoadBalancerControllerIAMPolicyName"`
	AWSLoadBalancerControllerIAMPolicyARN  string `json:"AWSLoadBalancerControllerIAMPolicyARN"`

	Route53ZoneId string `json:"Route53ZoneId"`

	// the certificate that is going to be used for SSL (https)
	AWSCertificateARN     string                                 `json:"AWSCertificateARN"`
	RDS                   MMDeployEnvironment_RDS                `json:"RDS"`
	MattermostInstance    MMDeployEnvironment_MattermostInstance `json:"MattermostInstance"`
	InfraComponents       MMDeployEnvironment_InfraComponents    `json:"InfraComponents"`
	OutputDir             string                                 `json:"OutputDir"`
	AWSCredentialProfile  string                                 `json:"AWSCredentialProfile"`
	DeployBucket          string                                 `json:"DeployBucket"`
	Containers            MMDeployEnvironment_Containers         `json:"Containers"`
	VcubeOauth            MMDeployEnvironment_VcubeOauth         `json:"VcubeOauth"`
	ImportBucketRegion    string                                 `json:"ImportBucketRegion"`
	ImportBucket          string                                 `json:"ImportBucket"`
	KubernetesContext     string                                 `json:"KubernetesContext"`
	KubernetesEnvironment string                                 `json:"KubernetesEnvironment"`
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
	ELB            *elbv2.ELBV2
	CF             *cloudformation.CloudFormation
	IAM            *iam.IAM
	RDS            *rds.RDS
	R53            *route53.Route53
	ConfigFile     string
	DomainsFile    string
	Subnets        []*ec2.Subnet
	PrivateSubnets []*ec2.Subnet
	PublicSubnets  []*ec2.Subnet
	Domains        []*MattermostDomainDeployment
}

type MattermostDomainDeployment struct {
	Key            string `json:"key"`
	Domain         string `json:"domain"`
	Replicas       string `json:"replicas"`
	CompanyId      string `json:"companyId"`
	SiteName       string `json:"site-name"`
	AdminEmail     string `json:"admin-email"`
	AdminEmailName string `json:"admin-email-name"`
	CompanyName    string `json:"company-name"`
	DockerRepoTag  string `json:"docker-repo-tag"`
	DeployEnv      string `json:"deploy-env"`
	ClientLocale   string `json:"client-locale"`
	OverrideDBUser string `json:"OverrideDBUser"`
}

type Token struct {
	Key,
	Value,
	Default string
	Required bool
}

type DeploymentEnvironment struct {
	// the aws access to be used by the containers when communicating to AWS infra such as creating the
	// instance-specific S3 bucket
	AWS_ACCESS_KEY_ID string `json:"AWS_ACCESS_KEY_ID"`
	// the aws access to be used by the containers when communicating to AWS infra such as creating the
	// instance-specific S3 bucket
	AWS_SECRET_ACCESS_KEY string `json:"AWS_SECRET_ACCESS_KEY"`

	AWS_PROD_S3_ACCESS_KEY_ID     string `json:"AWS_PROD_S3_ACCESS_KEY_ID"`
	AWS_PROD_S3_SECRET_ACCESS_KEY string `json:"AWS_PROD_S3_SECRET_ACCESS_KEY"`

	// the eks cluster name
	AWS_EKS_CLUSTER_NAME string `json:"AWS_EKS_CLUSTER_NAME"`

	// the VPC ID for the cluster
	AWS_VPC_ID string `json:"AWS_VPC_ID"`

	// the public subnets to being used by EKS
	EKS_PUBLIC_SUBNETS string `json:"EKS_PUBLIC_SUBNETS"`

	// the AWS region where the cluster is
	AWS_REGION string `json:"AWS_REGION"`

	// deployment S3 bucket. This is the bucket where plugin is stored
	DEPLOY_BUCKET string `json:"DEPLOY_BUCKET"`

	// TODO
	IMPORT_EXTERNAL_BUCKET        string `json:"IMPORT_EXTERNAL_BUCKET"`
	IMPORT_EXTERNAL_BUCKET_REGION string `json:"IMPORT_EXTERNAL_BUCKET_REGION"`

	// TODO
	AWS_ACM_CERTIFICATE_ARN string `json:"AWS_ACM_CERTIFICATE_ARN"`

	// The mattermost app should listen to if not the default 8065.
	MATTERMOST_PORT string `json:"MATTERMOST_PORT"`

	// the database details for mattermost instance
	MM_DB_HOST string `json:"MM_DB_HOST"`
	// the database details for mattermost instance
	MM_DB_PORT string `json:"MM_DB_PORT"`
	// the database details for mattermost instance
	MM_DB_MASTER_USER string `json:"MM_DB_MASTER_USER"`
	// the database details for mattermost instance
	MM_DB_MASTER_PASS string `json:"MM_DB_MASTER_PASS"`

	// the initial version mentioned in the Deployment/StatefulSet label
	NGINX_CONFIG_VERSION string `json:"NGINX_CONFIG_VERSION"`

	// the initial version mentioned in the Deployment/StatefulSet label
	MM_DEPLOY_VERSION string `json:"MM_DEPLOY_VERSION"`

	// Initial plugin upload settings in mattermost
	MM_CONF_PLUGIN_ENABLE_UPLOAD string `json:"MM_CONF_PLUGIN_ENABLE_UPLOAD"`

	// smtp details for mattermost to use
	SMTP_USER string `json:"SMTP_USER"`
	// smtp details for mattermost to use
	SMTP_PASS string `json:"SMTP_PASS"`
	// smtp details for mattermost to use
	SMTP_HOST string `json:"SMTP_HOST"`
	// smtp details for mattermost to use
	SMTP_PORT string `json:"SMTP_PORT"`
	// the sender smtp details for mattermost to use. Some SMTP providers may reject the sender if not whitelisted
	SMTP_FROM string `json:"SMTP_FROM"`

	// the initial version mentioned in the Deployment/StatefulSet label
	MM_PROXY_PROXY_CONFIG_VERSION string `json:"MM_PROXY_PROXY_CONFIG_VERSION"`

	// the push proxy notification service details
	MATTERMOST_PUSH_NOTIFICATION_URL string `json:"MATTERMOST_PUSH_NOTIFICATION_URL"`
	// the push proxy container repo
	MATTERMOST_PUSH_PROXY_DOCKER_REPO string `json:"MATTERMOST_PUSH_PROXY_DOCKER_REPO"`

	// the mattermost container (not docker) repo
	MM_DOCKER_REPO string `json:"MM_DOCKER_REPO"`

	// the mattermost custom cluster driver to be used if not the default, 'redis'
	MM_CLUSTER_DRIVER string `json:"MM_CLUSTER_DRIVER"`

	// the mattermost custom cluster redis details
	MM_CLUSTER_REDIS_HOST string `json:"MM_CLUSTER_REDIS_HOST"`
	// the mattermost custom cluster redis details
	MM_CLUSTER_REDIS_PORT string `json:"MM_CLUSTER_REDIS_PORT"`
	// the mattermost custom cluster redis details
	MM_CLUSTER_REDIS_PASS string `json:"MM_CLUSTER_REDIS_PASS"`

	// ALB Ingress Controller Name
	ALB_INGRESS_CONTROLLER_NAME           string `json:"ALB_INGRESS_CONTROLLER_NAME"`
	ALB_INGRESS_CONTROLLER_IAM_POLICY     string `json:"ALB_INGRESS_CONTROLLER_IAM_POLICY"`
	ALB_INGRESS_CONTROLLER_IAM_POLICY_ARN string `json:"ALB_INGRESS_CONTROLLER_IAM_POLICY_ARN"`

	// VID OAuth Provider initial admin user
	VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME string `json:"VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME"`
	// VID OAuth Provider initial admin password
	VCUBE_VID_OAUTH_INITIAL_ADMIN_PASSWORD string `json:"VCUBE_VID_OAUTH_INITIAL_ADMIN_PASSWORD"`

	// VID OAuth Provider session secret
	VCUBE_VID_OAUTH_EXPRESS_SESSION_SECRET string `json:"VCUBE_VID_OAUTH_EXPRESS_SESSION_SECRET"`

	// VID OAuth Provider VMeeting service details
	VCUBE_VID_OAUTH_VMEETING_URL string `json:"VCUBE_VID_OAUTH_VMEETING_URL"`
	// VID OAuth Provider VMeeting service details
	VCUBE_VID_OAUTH_VID_CONSUMER_KEY string `json:"VCUBE_VID_OAUTH_VID_CONSUMER_KEY"`
	// VID OAuth Provider VMeeting service details
	VCUBE_VID_OAUTH_VID_REST_PWD string `json:"VCUBE_VID_OAUTH_VID_REST_PWD"`
	// VID OAuth Provider VMeeting service details
	VCUBE_VID_OAUTH_VID_REST_URL string `json:"VCUBE_VID_OAUTH_VID_REST_URL"`
	// VID OAuth Provider VMeeting service details
	VCUBE_VID_OAUTH_VID_SECRET_AUTH_CODE string `json:"VCUBE_VID_OAUTH_VID_SECRET_AUTH_CODE"`

	// the initial version mentioned in the Deployment/StatefulSet label
	VCUBE_VID_OAUTH_CONTAINER_VERSION string `json:"VCUBE_VID_OAUTH_CONTAINER_VERSION"`
	// the mattermost container (not docker) repo
	VCUBE_VID_OAUTH_CONTAINER_REPO string `json:"VCUBE_VID_OAUTH_CONTAINER_REPO"`

	// VID OAuth Provider VMeeting service database details
	VCUBE_VID_OAUTH_DB_NAME string `json:"VCUBE_VID_OAUTH_DB_NAME"`
	// VID OAuth Provider VMeeting service database details
	VCUBE_VID_OAUTH_DB_USERNAME string `json:"VCUBE_VID_OAUTH_DB_USERNAME"`
	// VID OAuth Provider VMeeting service database details
	VCUBE_VID_OAUTH_DB_PASSWORD string `json:"VCUBE_VID_OAUTH_DB_PASSWORD"`
}

func DeploymentEnvironmentFromJson(_json string) (*DeploymentEnvironment, error) {
	c := &DeploymentEnvironment{}

	err := json.Unmarshal([]byte(_json), c)

	return c, err
}

func DeploymentEnvironmentToJsonString(c *DeploymentEnvironment) (string, error) {
	b, err := json.MarshalIndent(c, "", "\t")

	return string(b), err
}

func (d *DeploymentEnvironment) getToken(key, def string, req bool) *Token {
	r := reflect.ValueOf(d)
	f := reflect.Indirect(r).FieldByName(key)
	return &Token{Key: "__" + key + "__", Value: f.String(), Default: def, Required: req}
}

func LoadTokenFromJson(json_file string) ([]*Token, error) {
	b, err := ioutil.ReadFile(json_file)

	if err != nil {
		return nil, err
	}

	_env := &map[string]string{}

	err = json.Unmarshal(b, _env)
	if err != nil {
		return nil, err
	}

	tokens := []*Token{}

	for key, val := range *_env {
		tokens = append(tokens, &Token{Key: "__" + key + "__", Value: val})
	}

	return tokens, nil
}

func ProcessTemplate(templateFile, destinationFile string, tokens []*Token, mode os.FileMode) (string, error) {
	fmt.Println("------ func ProcessTemplate(templateFile, destinationFile string, tokens []*Token, mode os.FileMode) (string,error)")

	fmt.Println("Processing template:", templateFile)

	dat, err := ioutil.ReadFile(templateFile)

	if err != nil {
		return "", err
	}

	template := string(dat)

	// fmt.Println("------------------ tokens:", tokens)

	for _, token := range tokens {
		val := token.Value
		key := token.Key

		if val == "" && token.Default != "" {
			val = token.Default
		}

		fmt.Println("key:", key, "val:", val)

		template = strings.ReplaceAll(template, key, val)
	}

	fmt.Println("template:", template)

	if destinationFile != "" {
		err := ioutil.WriteFile(destinationFile, []byte(template), mode)

		if err != nil {
			return "", err
		}
	}

	return template, nil
}

func (m *MMDeployContext) LoadDomains() error {
	dat, err := ioutil.ReadFile(m.DomainsFile)

	if err != nil {
		return err
	}

	// d := json.NewDecoder(strings.NewReader(string(dat)))

	// d.UseNumber()

	// d.Decode(&domains)

	domains := []*MattermostDomainDeployment{}

	err = json.Unmarshal(dat, &domains)

	if err != nil {
		return err
	}

	for _, domain := range domains {
		if domain.DockerRepoTag == "" {
			domain.DockerRepoTag = "change-in-" + m.DomainsFile
		}
		if domain.DeployEnv == "" {
			domain.DeployEnv = "env"
		}
	}

	fmt.Println("domains:", domains)

	b, err := json.MarshalIndent(domains, "", "\t")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(m.DomainsFile, b, 0666)
	if err != nil {
		return err
	}

	m.Domains = domains

	return nil
}

func (m *MMDeployContext) ProcessDomains(tokens []*Token, baseDir string) (string, string, error) {
	nginx_domains := []string{}
	alb_domains := []string{}

	err := os.Mkdir(baseDir+"/domains", 0777)

	if err != nil && !strings.Contains(err.Error(), "file exists") {
		return "", "", err
	}

	for _, domain := range m.Domains {
		fmt.Println("domain:", domain)

		domain_tokens := []*Token{
			&Token{Key: "__MM_INSTANCE_COMPANY_NAME__", Value: domain.CompanyName},
			&Token{Key: "__MM_INSTANCE_CLIENT_LOCALE__", Value: domain.ClientLocale, Default: "en"},
			&Token{Key: "__MM_INSTANCE_ADMIN_EMAIL_NAME__", Value: domain.AdminEmailName},
			&Token{Key: "__MM_INSTANCE_ADMIN_EMAIL__", Value: domain.AdminEmail},
			&Token{Key: "__MM_INSTANCE_KEY__", Value: domain.Key},
			&Token{Key: "__MM_INSTANCE_DOMAIN__", Value: domain.Domain},
			&Token{Key: "__MM_INSTANCE_REPLICAS__", Value: domain.Replicas},
			&Token{Key: "__MM_COMPANY_ID__", Value: domain.CompanyId},
			&Token{Key: "__MM_DB_NAME__", Value: "mm_" + strings.ReplaceAll(domain.Key, "-", "_")},
			&Token{Key: "__MM_DB_PASS__", Value: "mm_" + domain.Key + "-mostest"},
			&Token{Key: "__MM_DOCKER_REPO_TAG__", Value: domain.DockerRepoTag},
			&Token{Key: "__MM_DEPLOY_ENV__", Value: domain.DeployEnv, Default: "dev"},
		}

		dbUserToken := &Token{Key: "__MM_DB_USER__", Value: "mm_" + domain.Key}
		if domain.OverrideDBUser != "" {
			dbUserToken.Value = domain.OverrideDBUser
		}

		domain_tokens = append(domain_tokens, dbUserToken)

		for _, domain_token := range domain_tokens {
			fmt.Println("domain_token:", *domain_token)
		}

		nginx_domains_dat, err := ProcessTemplate("templates/configmap_domain.yaml.template", "", append(tokens, domain_tokens...), 0666)
		if err != nil {
			return "", "", err
		}
		nginx_domains = append(nginx_domains, nginx_domains_dat)

		alb_domains_dat, err := ProcessTemplate("templates/alb-domain-host.yaml.template", "", append(tokens, domain_tokens...), 0666)
		if err != nil {
			return "", "", err
		}
		alb_domains = append(alb_domains, alb_domains_dat)

		domainBaseDir := baseDir + "/domains/" + domain.Key

		err = os.Mkdir(domainBaseDir, 0777)
		if err != nil {
			return "", "", err
		}

		_, err = ProcessTemplate("templates/mm_domain_deploy_service.yaml.template", fmt.Sprintf(domainBaseDir+"/mm_domain_deploy_service-%s.yaml", domain.Key), append(tokens, domain_tokens...), 0666)
		if err != nil {
			return "", "", err
		}

		_, err = ProcessTemplate("templates/mm_domain_docker_starter.template", fmt.Sprintf(domainBaseDir+"/mm_domain_docker_starter-%s.sh", domain.Key), append(tokens, domain_tokens...), 0755)
		if err != nil {
			return "", "", err
		}
	}

	return strings.Join(nginx_domains, "\n"), strings.Join(alb_domains, "\n"), nil
}
