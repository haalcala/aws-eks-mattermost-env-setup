package stage2

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
)

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

func ConfigWithDefaults() *DeploymentEnvironment {
	return &DeploymentEnvironment{
		AWS_ACCESS_KEY_ID:                      strings.Trim(os.Getenv("__AWS_ACCESS_KEY_ID__"), "\r"),
		AWS_SECRET_ACCESS_KEY:                  strings.Trim(os.Getenv("__AWS_SECRET_ACCESS_KEY__"), "\r"),
		AWS_PROD_S3_ACCESS_KEY_ID:              strings.Trim(os.Getenv("__AWS_PROD_S3_ACCESS_KEY_ID__"), "\r"),
		AWS_PROD_S3_SECRET_ACCESS_KEY:          strings.Trim(os.Getenv("__AWS_PROD_S3_SECRET_ACCESS_KEY__"), "\r"),
		AWS_EKS_CLUSTER_NAME:                   strings.Trim(os.Getenv("__AWS_EKS_CLUSTER_NAME__"), "\r"),
		AWS_VPC_ID:                             strings.Trim(os.Getenv("__AWS_VPC_ID__"), "\r"),
		EKS_PUBLIC_SUBNETS:                     strings.Trim(os.Getenv("__EKS_PUBLIC_SUBNETS__"), "\r"),
		AWS_REGION:                             strings.Trim(os.Getenv("__AWS_REGION__"), "\r"),
		DEPLOY_BUCKET:                          strings.Trim(os.Getenv("__DEPLOY_BUCKET__"), "\r"),
		IMPORT_EXTERNAL_BUCKET:                 strings.Trim(os.Getenv("__IMPORT_EXTERNAL_BUCKET__"), "\r"),
		IMPORT_EXTERNAL_BUCKET_REGION:          strings.Trim(os.Getenv("__IMPORT_EXTERNAL_BUCKET_REGION__"), "\r"),
		AWS_ACM_CERTIFICATE_ARN:                strings.Trim(os.Getenv("__AWS_ACM_CERTIFICATE_ARN__"), "\r"),
		MATTERMOST_PORT:                        strings.Trim(os.Getenv("__MATTERMOST_PORT__"), "\r"),
		MM_DB_HOST:                             strings.Trim(os.Getenv("__MM_DB_HOST__"), "\r"),
		MM_DB_PORT:                             strings.Trim(os.Getenv("__MM_DB_PORT__"), "\r"),
		MM_DB_MASTER_USER:                      strings.Trim(os.Getenv("__MM_DB_MASTER_USER__"), "\r"),
		MM_DB_MASTER_PASS:                      strings.Trim(os.Getenv("__MM_DB_MASTER_PASS__"), "\r"),
		NGINX_CONFIG_VERSION:                   strings.Trim(os.Getenv("__NGINX_CONFIG_VERSION__"), "\r"),
		MM_DEPLOY_VERSION:                      strings.Trim(os.Getenv("__MM_DEPLOY_VERSION__"), "\r"),
		MM_CONF_PLUGIN_ENABLE_UPLOAD:           strings.Trim(os.Getenv("__MM_CONF_PLUGIN_ENABLE_UPLOAD__"), "\r"),
		SMTP_USER:                              strings.Trim(os.Getenv("__SMTP_USER__"), "\r"),
		SMTP_PASS:                              strings.Trim(os.Getenv("__SMTP_PASS__"), "\r"),
		SMTP_HOST:                              strings.Trim(os.Getenv("__SMTP_HOST__"), "\r"),
		SMTP_PORT:                              strings.Trim(os.Getenv("__SMTP_PORT__"), "\r"),
		SMTP_FROM:                              strings.Trim(os.Getenv("__SMTP_FROM__"), "\r"),
		MM_PROXY_PROXY_CONFIG_VERSION:          strings.Trim(os.Getenv("__MM_PROXY_PROXY_CONFIG_VERSION__"), "\r"),
		MATTERMOST_PUSH_NOTIFICATION_URL:       strings.Trim(os.Getenv("__MATTERMOST_PUSH_NOTIFICATION_URL__"), "\r"),
		MATTERMOST_PUSH_PROXY_DOCKER_REPO:      strings.Trim(os.Getenv("__MATTERMOST_PUSH_PROXY_DOCKER_REPO__"), "\r"),
		MM_DOCKER_REPO:                         strings.Trim(os.Getenv("__MM_DOCKER_REPO__"), "\r"),
		MM_CLUSTER_DRIVER:                      strings.Trim(os.Getenv("__MM_CLUSTER_DRIVER__"), "\r"),
		MM_CLUSTER_REDIS_HOST:                  strings.Trim(os.Getenv("__MM_CLUSTER_REDIS_HOST__"), "\r"),
		MM_CLUSTER_REDIS_PORT:                  strings.Trim(os.Getenv("__MM_CLUSTER_REDIS_PORT__"), "\r"),
		MM_CLUSTER_REDIS_PASS:                  strings.Trim(os.Getenv("__MM_CLUSTER_REDIS_PASS__"), "\r"),
		VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME: strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME__"), "\r"),
		VCUBE_VID_OAUTH_INITIAL_ADMIN_PASSWORD: strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_INITIAL_ADMIN_PASSWORD__"), "\r"),
		VCUBE_VID_OAUTH_EXPRESS_SESSION_SECRET: strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_EXPRESS_SESSION_SECRET__"), "\r"),
		VCUBE_VID_OAUTH_VMEETING_URL:           strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_VMEETING_URL__"), "\r"),
		VCUBE_VID_OAUTH_VID_CONSUMER_KEY:       strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_VID_CONSUMER_KEY__"), "\r"),
		VCUBE_VID_OAUTH_VID_REST_PWD:           strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_VID_REST_PWD__"), "\r"),
		VCUBE_VID_OAUTH_VID_REST_URL:           strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_VID_REST_URL__"), "\r"),
		VCUBE_VID_OAUTH_VID_SECRET_AUTH_CODE:   strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_VID_SECRET_AUTH_CODE__"), "\r"),
		VCUBE_VID_OAUTH_CONTAINER_VERSION:      strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_CONTAINER_VERSION__"), "\r"),
		VCUBE_VID_OAUTH_CONTAINER_REPO:         strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_CONTAINER_REPO__"), "\r"),
		VCUBE_VID_OAUTH_DB_NAME:                strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_DB_NAME__"), "\r"),
		VCUBE_VID_OAUTH_DB_USERNAME:            strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_DB_USERNAME__"), "\r"),
		VCUBE_VID_OAUTH_DB_PASSWORD:            strings.Trim(os.Getenv("__VCUBE_VID_OAUTH_DB_PASSWORD__"), "\r"),
	}
}

func (d *DeploymentEnvironment) getToken(key, def string, req bool) *Token {
	r := reflect.ValueOf(d)
	f := reflect.Indirect(r).FieldByName(key)
	return &Token{Key: "__" + key + "__", Value: f.String(), Default: def, Required: req}
}

func (d *DeploymentEnvironment) LoadTokenEnvironment() []*Token {
	return []*Token{
		d.getToken("AWS_ACCESS_KEY_ID", "", true),
		d.getToken("AWS_SECRET_ACCESS_KEY", "", true),
		d.getToken("AWS_PROD_S3_ACCESS_KEY_ID", "", true),
		d.getToken("AWS_PROD_S3_SECRET_ACCESS_KEY", "", true),
		d.getToken("AWS_EKS_CLUSTER_NAME", "", true),
		d.getToken("AWS_VPC_ID", "", true),
		d.getToken("EKS_PUBLIC_SUBNETS", "", true),
		d.getToken("AWS_REGION", "", true),
		d.getToken("DEPLOY_BUCKET", "", true),
		d.getToken("IMPORT_EXTERNAL_BUCKET", "", true),
		d.getToken("IMPORT_EXTERNAL_BUCKET_REGION", "", true),
		d.getToken("AWS_ACM_CERTIFICATE_ARN", "", true),
		d.getToken("MATTERMOST_PORT", "8065", true),
		d.getToken("MM_DB_HOST", "", true),
		d.getToken("MM_DB_PORT", "", true),
		d.getToken("MM_DB_MASTER_USER", "", true),
		d.getToken("MM_DB_MASTER_PASS", "", true),
		d.getToken("NGINX_CONFIG_VERSION", "", true),
		d.getToken("MM_DEPLOY_VERSION", "", true),
		d.getToken("MM_CONF_PLUGIN_ENABLE_UPLOAD", "false", true),
		d.getToken("SMTP_USER", "", true),
		d.getToken("SMTP_PASS", "", true),
		d.getToken("SMTP_HOST", "", true),
		d.getToken("SMTP_PORT", "", true),
		d.getToken("SMTP_FROM", "", true),
		d.getToken("MM_PROXY_PROXY_CONFIG_VERSION", "v1", true),
		d.getToken("MATTERMOST_PUSH_NOTIFICATION_URL", "https://push-test.mattermost.com", true),
		d.getToken("MATTERMOST_PUSH_PROXY_DOCKER_REPO", "haalcala/mattermost-push-proxy", true),
		d.getToken("MM_DOCKER_REPO", "haalcala/mattermost-prod", true),
		d.getToken("MM_CLUSTER_DRIVER", "", true),
		d.getToken("MM_CLUSTER_REDIS_HOST", "localhost", true),
		d.getToken("MM_CLUSTER_REDIS_PORT", "6379", true),
		d.getToken("MM_CLUSTER_REDIS_PASS", "", false),
		d.getToken("VCUBE_VID_OAUTH_INITIAL_ADMIN_USERNAME", "", true),
		d.getToken("VCUBE_VID_OAUTH_INITIAL_ADMIN_PASSWORD", "", true),
		d.getToken("VCUBE_VID_OAUTH_EXPRESS_SESSION_SECRET", "", true),
		d.getToken("VCUBE_VID_OAUTH_VMEETING_URL", "", true),
		d.getToken("VCUBE_VID_OAUTH_VID_CONSUMER_KEY", "", true),
		d.getToken("VCUBE_VID_OAUTH_VID_REST_PWD", "", true),
		d.getToken("VCUBE_VID_OAUTH_VID_REST_URL", "", true),
		d.getToken("VCUBE_VID_OAUTH_VID_SECRET_AUTH_CODE", "", true),
		d.getToken("VCUBE_VID_OAUTH_CONTAINER_VERSION", "", true),
		d.getToken("VCUBE_VID_OAUTH_CONTAINER_REPO", "", true),
		d.getToken("VCUBE_VID_OAUTH_DB_NAME", "", true),
		d.getToken("VCUBE_VID_OAUTH_DB_USERNAME", "", true),
		d.getToken("VCUBE_VID_OAUTH_DB_PASSWORD", "", true),
	}
}

func ProcessTemplate(templateFile, destinationFile string, tokens []*Token, mode os.FileMode) string {
	fmt.Println("------ func ProcessTemplate(templateFile, destinationFile string, tokens []*Token, mode os.FileMode) string")

	fmt.Println("Processing template:", templateFile)

	dat, err := ioutil.ReadFile(templateFile)

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
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
		ioutil.WriteFile(destinationFile, []byte(template), mode)
	}

	return template
}

func LoadDomains(tokens []*Token, baseDir string) (string, string) {
	dat, err := ioutil.ReadFile("domains.json")

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	var domains []*MattermostDomainDeployment

	// d := json.NewDecoder(strings.NewReader(string(dat)))

	// d.UseNumber()

	// d.Decode(&domains)

	err = json.Unmarshal(dat, &domains)

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	for _, domain := range domains {
		if domain.DockerRepoTag == "" {
			domain.DockerRepoTag = "change-in-domains.json"
		}
		if domain.DeployEnv == "" {
			domain.DeployEnv = "env"
		}
	}

	fmt.Println("domains:", domains)

	b, err := json.MarshalIndent(domains, "", "\t")

	ioutil.WriteFile("domains.json", b, 0666)

	nginx_domains := []string{}
	alb_domains := []string{}

	err = os.Mkdir(baseDir+"/domains", 0777)

	for _, domain := range domains {
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

		nginx_domains = append(nginx_domains, ProcessTemplate("configmap_domain.yaml.template", "", append(tokens, domain_tokens...), 0666))
		alb_domains = append(alb_domains, ProcessTemplate("alb-domain-host.yaml.template", "", append(tokens, domain_tokens...), 0666))

		domainBaseDir := baseDir + "/domains/" + domain.Key

		os.Mkdir(domainBaseDir, 0777)

		ProcessTemplate("mm_domain_deploy_service.yaml.template", fmt.Sprintf(domainBaseDir+"/mm_domain_deploy_service-%s.yaml", domain.Key), append(tokens, domain_tokens...), 0666)

		ProcessTemplate("mm_domain_docker_starter.template", fmt.Sprintf(domainBaseDir+"/mm_domain_docker_starter-%s.sh", domain.Key), append(tokens, domain_tokens...), 0755)
	}

	return strings.Join(nginx_domains, "\n"), strings.Join(alb_domains, "\n")
}
