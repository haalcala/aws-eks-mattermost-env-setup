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
	{Key: "__MATTERMOST_PORT__", Default: "8065"}}

type Token struct {
	Key     string
	Default string
}

func main() {
	dat, err := ioutil.ReadFile("./deploy-nginx-router.yaml.template")

	if err != nil {
		fmt.Println("err:", err)
		os.Exit(1)
	}

	template := string(dat)

	fmt.Println("tokens:", tokens)

	for _, token := range tokens {
		val := os.Getenv(token.Key)

		fmt.Println("val:", val)

		if val == "" && token.Default == "" {
			fmt.Println("Missing required environment variable:", token)
			os.Exit(1)
		}

		if val == "" && token.Default != "" {
			val = token.Default
		}

		template = strings.ReplaceAll(template, token.Key, val)
	}

	fmt.Println("dat:", template)

	ioutil.WriteFile("./deploy-nginx-router.yaml", []byte(template), 0666)
}
