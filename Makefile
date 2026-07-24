.PHONY: deploy plan validate destroy clean

deploy:
	terraform init -input=false
	terraform fmt -recursive
	terraform validate
	terraform plan -input=false -out=tfplan
	terraform apply -input=false tfplan

clean:
	rm -f tfplan