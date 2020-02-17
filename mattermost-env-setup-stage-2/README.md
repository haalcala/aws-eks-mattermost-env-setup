Copy .env_sample to .env_file and modify

## execute

    source ./.env_file; go run  generate_deployment.go common.go

    for file in `ls mm_domain_deploy_service/*.yaml`; do kubectl apply -f $file; done; kubectl apply -f deploy-nginx-router.yaml
