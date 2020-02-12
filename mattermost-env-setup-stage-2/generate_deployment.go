package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var tokens = []Token{{Key: "__AWS_ACCESS_KEY_ID__", Value: os.Getenv("__AWS_ACCESS_KEY_ID__")},
	{Key: "__AWS_SECRET_ACCESS_KEY__", Value: os.Getenv("__AWS_SECRET_ACCESS_KEY__")},
	{Key: "__AWS_EKS_CLUSTER_NAME__", Value: os.Getenv("__AWS_EKS_CLUSTER_NAME__")},
	{Key: "__AWS_VPC_ID__", Value: os.Getenv("__AWS_VPC_ID__")},
	{Key: "__AWS_REGION__", Value: os.Getenv("__AWS_REGION__")},
	{Key: "__AWS_ACM_CERTIFICATE_ARN__", Value: os.Getenv("__AWS_ACM_CERTIFICATE_ARN__")},
	{Key: "__MATTERMOST_PORT__", Default: "8065", Value: os.Getenv("__MATTERMOST_PORT__")},
	{Key: "__DB_NAME__", Value: os.Getenv("__DB_NAME__")},
	{Key: "__DB_USER__", Value: os.Getenv("__DB_USER__")},
	{Key: "__DB_PASS__", Value: os.Getenv("__DB_PASS__")},
	{Key: "__DB_HOST__", Value: os.Getenv("__DB_HOST__")},
	{Key: "__DB_PORT__", Value: os.Getenv("__DB_PORT__")}}

type Token struct {
	Key,
	Value,
	Default string
}

type MattermostDeployment struct {
	Key     string `json:"key"`
	Domain  string `json:"domain"`
	Replica int    `json:"replica"`
}

func processTemplate(templateFile, destinationFile string, tokens []Token) string {
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
		ioutil.WriteFile(destinationFile, []byte(template), 0666)
	}

	return template
}

func loadDomains() string {
	dat, err := ioutil.ReadFile("./domains.json")

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	var domains []MattermostDeployment

	json.NewDecoder(strings.NewReader(string(dat))).Decode(&domains)

	fmt.Println("domains:", domains)

	ret := []string{}

	err = os.Mkdir("./mm_domain_deploy_service", 0777)

	for _, domain := range domains {
		fmt.Println("domain:", domain)

		domain_tokens := []Token{
			{Key: "__MM_INSTANCE_KEY__", Value: domain.Key},
			{Key: "__MM_INSTANCE_DOMAIN__", Value: domain.Domain},
			{Key: "__MM_INSTANCE_REPLICA__", Value: string(domain.Replica)}}

		fmt.Println("domain_tokens:", domain_tokens)

		ret = append(ret, processTemplate("./configmap_domain.yaml.template", "", append(tokens, domain_tokens...)))

		_ = processTemplate("./mm_domain_deploy_service.yaml.template", fmt.Sprintf("./mm_domain_deploy_service/mm_domain_deploy_service-%s.yaml", domain.Key), append(tokens, domain_tokens...))
	}

	return strings.Join(ret, "\n\n")
}

func main() {
	for _, token := range tokens {
		val := os.Getenv(token.Key)

		fmt.Println("Key:", token.Key, "val:", val)

		if val == "" && token.Default == "" {
			fmt.Println("Missing required environment variable:", token)
			os.Exit(1)
		}
	}

	domain_conf := loadDomains()

	// fmt.Println("domain_conf:", domain_conf)

	processTemplate("./deploy-nginx-router.yaml.template", "./deploy-nginx-router.yaml", append(tokens, Token{Key: "__NGINX_MM_DOMAINS__", Value: domain_conf}))
}