package main

import (
	"fmt"
	"os"
)

func main() {
	for _, token := range tokens {
		val := os.Getenv(token.Key)

		fmt.Println("Key:", token.Key, "val:", val)

		if val == "" && token.Default == "" {
			fmt.Println("Missing required environment variable:", token)
			os.Exit(1)
		}
	}

	domain_conf := LoadDomains()

	// fmt.Println("domain_conf:", domain_conf)

	ProcessTemplate("./deploy-nginx-router.yaml.template", "./deploy-nginx-router.yaml", append(tokens, Token{Key: "__NGINX_MM_DOMAINS__", Value: domain_conf}))
}
