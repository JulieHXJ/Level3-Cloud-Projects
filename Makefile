

deploy:
	terraform init
	terraform fmt -check
	terraform validate
	terraform plan -out=tfplan
	terraform apply tfplan
	./scripts/wait-for-ssh.sh
	./scripts/install-kubernetes.sh