package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type MattermostDeployment struct {
	Key            string `json:"key"`
	Domain         string `json:"domain"`
	Replicas       string `json:"replicas"`
	CompanyId      string `json:"companyId"`
	SiteName       string `json:"site-name"`
	AdminEmail     string `json:"admin-email"`
	AdminEmailName string `json:"admin-email-name"`
	CompanyName    string `json:"company-name"`
	DockerRepoTag  string `json:"docker_repo_tag"`
}

type Token struct {
	Key,
	Value,
	Default string
}

var tokens = []Token{{Key: "__AWS_ACCESS_KEY_ID__", Value: strings.Trim(os.Getenv("__AWS_ACCESS_KEY_ID__"), "\r")},
	{Key: "__AWS_SECRET_ACCESS_KEY__", Value: strings.Trim(os.Getenv("__AWS_SECRET_ACCESS_KEY__"), "\r")},
	{Key: "__AWS_PROD_S3_ACCESS_KEY_ID__", Value: strings.Trim(os.Getenv("__AWS_PROD_S3_ACCESS_KEY_ID__"), "\r")},
	{Key: "__AWS_PROD_S3_SECRET_ACCESS_KEY__", Value: strings.Trim(os.Getenv("__AWS_PROD_S3_SECRET_ACCESS_KEY__"), "\r")},
	{Key: "__AWS_EKS_CLUSTER_NAME__", Value: strings.Trim(os.Getenv("__AWS_EKS_CLUSTER_NAME__"), "\r")},
	{Key: "__AWS_VPC_ID__", Value: strings.Trim(os.Getenv("__AWS_VPC_ID__"), "\r")},
	{Key: "__EKS_PUBLIC_SUBNETS__", Value: strings.Trim(os.Getenv("__EKS_PUBLIC_SUBNETS__"), "\r")},
	{Key: "__AWS_REGION__", Value: strings.Trim(os.Getenv("__AWS_REGION__"), "\r")},
	{Key: "__DEPLOY_BUCKET__", Value: strings.Trim(os.Getenv("__DEPLOY_BUCKET__"), "\r")},
	{Key: "__IMPORT_EXTERNAL_BUCKET__", Value: strings.Trim(os.Getenv("__IMPORT_EXTERNAL_BUCKET__"), "\r")},
	{Key: "__IMPORT_EXTERNAL_BUCKET_REGION__", Value: strings.Trim(os.Getenv("__IMPORT_EXTERNAL_BUCKET_REGION__"), "\r")},
	{Key: "__AWS_ACM_CERTIFICATE_ARN__", Value: strings.Trim(os.Getenv("__AWS_ACM_CERTIFICATE_ARN__"), "\r")},
	{Key: "__MATTERMOST_PORT__", Default: "8065", Value: strings.Trim(os.Getenv("__MATTERMOST_PORT__"), "\r")},
	{Key: "__MM_DB_HOST__", Value: strings.Trim(os.Getenv("__MM_DB_HOST__"), "\r")},
	{Key: "__MM_DB_PORT__", Value: strings.Trim(os.Getenv("__MM_DB_PORT__"), "\r")},
	{Key: "__MM_DB_MASTER_USER__", Value: strings.Trim(os.Getenv("__MM_DB_MASTER_USER__"), "\r")},
	{Key: "__MM_DB_MASTER_PASS__", Value: strings.Trim(os.Getenv("__MM_DB_MASTER_PASS__"), "\r")},
	{Key: "__NGINX_CONFIG_VERSION__", Value: strings.Trim(os.Getenv("__NGINX_CONFIG_VERSION__"), "\r")},
	{Key: "__MM_DEPLOY_VERSION__", Value: strings.Trim(os.Getenv("__MM_DEPLOY_VERSION__"), "\r")},
	{Key: "__MM_CONF_PLUGIN_ENABLE_UPLOAD__", Value: strings.Trim(os.Getenv("__MM_CONF_PLUGIN_ENABLE_UPLOAD__"), "\r"), Default: "false"},
	{Key: "__SMTP_USER__", Value: strings.Trim(os.Getenv("__SMTP_USER__"), "\r"), Default: ""},
	{Key: "__SMTP_PASS__", Value: strings.Trim(os.Getenv("__SMTP_PASS__"), "\r"), Default: ""},
	{Key: "__SMTP_HOST__", Value: strings.Trim(os.Getenv("__SMTP_HOST__"), "\r"), Default: ""},
	{Key: "__SMTP_PORT__", Value: strings.Trim(os.Getenv("__SMTP_PORT__"), "\r"), Default: ""},
	{Key: "__SMTP_FROM__", Value: strings.Trim(os.Getenv("__SMTP_FROM__"), "\r"), Default: ""},
	{Key: "__MM_PROXY_PROXY_CONFIG_VERSION__", Value: strings.Trim(os.Getenv("__MM_PROXY_PROXY_CONFIG_VERSION__"), "\r"), Default: "v1"},
	{Key: "__MATTERMOST_PUSH_NOTIFICATION_URL__", Value: strings.Trim(os.Getenv("__MATTERMOST_PUSH_NOTIFICATION_URL__"), "\r"), Default: "https://push-test.mattermost.com"},
	{Key: "__MATTERMOST_PUSH_PROXY_DOCKER_REPO__", Value: strings.Trim(os.Getenv("__MATTERMOST_PUSH_PROXY_DOCKER_REPO__"), "\r"), Default: "haalcala/mattermost-push-proxy"},
	{Key: "__MM_DOCKER_REPO__", Value: strings.Trim(os.Getenv("__MM_DOCKER_REPO__"), "\r"), Default: "haalcala/mattermost-prod"}}

func ProcessTemplate(templateFile, destinationFile string, tokens []Token, mode os.FileMode) string {
	dat, err := ioutil.ReadFile(templateFile)

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	template := string(dat)

	fmt.Println("------------------ tokens:", tokens)

	for _, token := range tokens {
		val := token.Value
		key := token.Key

		if val == "" && token.Default != "" {
			val = token.Default
		}

		template = strings.ReplaceAll(template, key, val)
	}

	fmt.Println("template:", template)

	if destinationFile != "" {
		ioutil.WriteFile(destinationFile, []byte(template), mode)
	}

	return template
}

func LoadDomains() (string, string) {
	dat, err := ioutil.ReadFile("./domains.json")

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	var domains []MattermostDeployment

	d := json.NewDecoder(strings.NewReader(string(dat)))

	d.UseNumber()

	d.Decode(&domains)

	fmt.Println("domains:", domains)

	nginx_domains := []string{}
	alb_domains := []string{}

	err = os.Mkdir("./mm_domain_deploy_service", 0777)
	err = os.Mkdir("./mm_docker_starter", 0777)

	for _, domain := range domains {
		fmt.Println("domain:", domain)

		domain_tokens := []Token{
			{Key: "__MM_INSTANCE_COMPANY_NAME__", Value: domain.CompanyName},
			{Key: "__MM_INSTANCE_ADMIN_EMAIL_NAME__", Value: domain.AdminEmailName},
			{Key: "__MM_INSTANCE_ADMIN_EMAIL__", Value: domain.AdminEmail},
			{Key: "__MM_INSTANCE_KEY__", Value: domain.Key},
			{Key: "__MM_INSTANCE_DOMAIN__", Value: domain.Domain},
			{Key: "__MM_INSTANCE_REPLICAS__", Value: domain.Replicas},
			{Key: "__MM_COMPANY_ID__", Value: domain.CompanyId},
			{Key: "__MM_DB_NAME__", Value: "mm_" + strings.ReplaceAll(domain.Key, "-", "_")},
			{Key: "__MM_DB_USER__", Value: "mm_" + domain.Key + "-mmuser"},
			{Key: "__MM_DB_PASS__", Value: "mm_" + domain.Key + "-mostest"},
			{Key: "__MM_DOCKER_REPO_TAG__", Value: domain.DockerRepoTag, Default: "latest"}}

		fmt.Println("domain_tokens:", domain_tokens)

		nginx_domains = append(nginx_domains, ProcessTemplate("./configmap_domain.yaml.template", "", append(tokens, domain_tokens...), 0666))
		alb_domains = append(alb_domains, ProcessTemplate("./alb-domain-host.yaml.template", "", append(tokens, domain_tokens...), 0666))

		_ = ProcessTemplate("./mm_domain_deploy_service.yaml.template", fmt.Sprintf("./mm_domain_deploy_service/mm_domain_deploy_service-%s.yaml", domain.Key), append(tokens, domain_tokens...), 0666)

		_ = ProcessTemplate("./mm_domain_docker_starter.template", fmt.Sprintf("./mm_docker_starter/mm_domain_docker_starter-%s.sh", domain.Key), append(tokens, domain_tokens...), 0755)
	}

	return strings.Join(nginx_domains, "\n"), strings.Join(alb_domains, "\n")
}
