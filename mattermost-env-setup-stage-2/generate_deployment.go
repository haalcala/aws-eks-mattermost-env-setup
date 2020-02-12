package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var tokens = []Token{{Key: "__AWS_ACCESS_KEY_ID__"},
	{Key: "__AWS_SECRET_ACCESS_KEY__"},
	{Key: "__AWS_EKS_CLUSTER_NAME__"},
	{Key: "__AWS_VPC_ID__"},
	{Key: "__AWS_REGION__"},
	{Key: "__AWS_ACM_CERTIFICATE_ARN__"},
	{Key: "__MATTERMOST_PORT__", Default: "8065"},
	{Key: "__DB_NAME__"},
	{Key: "__DB_USER__"},
	{Key: "__DB_PASS__"},
	{Key: "__DB_HOST__"},
	{Key: "__DB_PORT__"}}

type Token struct {
	Key     string
	Default string
}

func processTemplate(templateFile, destinationFile string) {
	dat, err := ioutil.ReadFile(templateFile)

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	template := string(dat)

	fmt.Println("tokens:", tokens)

	for _, token := range tokens {
		val := os.Getenv(token.Key)

		if val == "" && token.Default != "" {
			val = token.Default
		}

		template = strings.ReplaceAll(template, token.Key, val)
	}

	fmt.Println("dat:", template)

	ioutil.WriteFile(destinationFile, []byte(template), 0666)
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

	processTemplate("./deploy-nginx-router.yaml.template", "./deploy-nginx-router.yaml")
}
