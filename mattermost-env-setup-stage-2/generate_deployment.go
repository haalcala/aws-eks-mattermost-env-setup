package main

import (
	"fmt"
	"os"
)

func main() {
	for _, token := range tokens {
		val := os.Getenv(token.Key)

		fmt.Println("Key:", token.Key, "val:", val)

		if val == "" && token.Default == "" && token.Required {
			fmt.Println("Missing required environment variable:", token)
			os.Exit(1)
		}
	}

	domain_conf, alb_domain_conf := LoadDomains()

	// fmt.Println("domain_conf:", domain_conf)

	_tokens := append(tokens, Token{Key: "__NGINX_MM_DOMAINS__", Value: domain_conf})
	_tokens = append(_tokens, Token{Key: "__ALB_DOMAIN_RULES__", Value: alb_domain_conf})

	ProcessTemplate("./deploy-nginx-router.yaml.template", "./deploy-nginx-router.yaml", _tokens, 0666)
	ProcessTemplate("./deploy-smtp.yaml.template", "./deploy-smtp.yaml", _tokens, 0666)
	ProcessTemplate("./deploy-push-proxy.yaml.template", "./deploy-push-proxy.yaml", _tokens, 0666)
	ProcessTemplate("./deploy-redis.yaml.template", "./deploy-redis.yaml", _tokens, 0666)
}
