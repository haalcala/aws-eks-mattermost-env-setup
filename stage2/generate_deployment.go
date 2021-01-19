package stage2

import (
	"fmt"
	"os"
)

func GenerateDeploymentFiles(baseDir string) {
	envConfig := ConfigWithDefaults()

	tokens := envConfig.LoadTokenEnvironment()

	if baseDir == "" {
		baseDir = "."
	}

	for _, token := range tokens {
		val := os.Getenv(token.Key)

		if val == "" && token.Default == "" && token.Required {
			fmt.Println("Missing required environment variable:", token)
			os.Exit(1)
		}
	}

	domain_conf, alb_domain_conf := LoadDomains(tokens, baseDir)

	// fmt.Println("domain_conf:", domain_conf)

	_tokens := append(tokens, &Token{Key: "__NGINX_MM_DOMAINS__", Value: domain_conf})
	_tokens = append(_tokens, &Token{Key: "__ALB_DOMAIN_RULES__", Value: alb_domain_conf})

	ProcessTemplate("deploy-nginx-router.yaml.template", baseDir+"/deploy-nginx-router.yaml", _tokens, 0666)
	ProcessTemplate("deploy-aws-alb.yaml.template", baseDir+"/deploy-aws-alb.yaml", _tokens, 0666)
	ProcessTemplate("deploy-smtp.yaml.template", baseDir+"/deploy-smtp.yaml", _tokens, 0666)
	ProcessTemplate("deploy-push-proxy.yaml.template", baseDir+"/deploy-push-proxy.yaml", _tokens, 0666)
	ProcessTemplate("deploy-redis.yaml.template", baseDir+"/deploy-redis.yaml", _tokens, 0666)
	ProcessTemplate("deploy-vid-oauth-wrapper.yaml.template", baseDir+"/deploy-vid-oauth-wrapper.yaml", _tokens, 0666)
	ProcessTemplate("configmap-metricbeat.yaml.template", baseDir+"/configmap-metricbeat.yaml", _tokens, 0666)
}

func main() {
	GenerateDeploymentFiles(".")
}
