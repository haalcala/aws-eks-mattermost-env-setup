package main

import (
	"path"
)

func (m *MMDeployContext) GenerateDeploymentFiles(baseDir string) error {
	if baseDir == "" {
		baseDir = "./"
	}

	tokens, err := LoadTokenFromJson(path.Join(baseDir, "env.json"))
	if err != nil {
		return err
	}

	domain_conf, alb_domain_conf, err := m.ProcessDomains(tokens, baseDir)
	if err != nil {
		return err
	}
	// fmt.Println("domain_conf:", domain_conf)

	_tokens := append(tokens, &Token{Key: "__NGINX_MM_DOMAINS__", Value: domain_conf})
	_tokens = append(_tokens, &Token{Key: "__ALB_DOMAIN_RULES__", Value: alb_domain_conf})

	_, err = ProcessTemplate("templates/deploy-common-components.sh.template", baseDir+"/deploy-common-components.sh", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/alb-ingress-iam-policy.json.template", baseDir+"/alb-ingress-iam-policy.json", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/rbac-role.yaml.template", baseDir+"/rbac-role.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/deploy-nginx-router.yaml.template", baseDir+"/deploy-nginx-router.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/deploy-aws-alb.yaml.template", baseDir+"/deploy-aws-alb.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/deploy-smtp.yaml.template", baseDir+"/deploy-smtp.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/deploy-push-proxy.yaml.template", baseDir+"/deploy-push-proxy.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/deploy-redis.yaml.template", baseDir+"/deploy-redis.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/deploy-vid-oauth-wrapper.yaml.template", baseDir+"/deploy-vid-oauth-wrapper.yaml", _tokens, 0666)
	if err != nil {
		return err
	}
	_, err = ProcessTemplate("templates/configmap-metricbeat.yaml.template", baseDir+"/configmap-metricbeat.yaml", _tokens, 0666)
	if err != nil {
		return err
	}

	return nil
}
