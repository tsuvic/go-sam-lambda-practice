.PHONY: build lcoal deploy

build:
	sam build

api: build
	sam local start-api --env-vars env.json

FUNCNAME = InfectionStatusRegisterFunction
invoke: build
	sam local invoke $(FUNCNAME) --env-vars env.json 

deploy: build
	sam deploy --s3-bucket XXX --profile XXX --parameter-overrides ENV=PROD DBHOST=XXX DBNAME=XXX DBUSER=XXX DBPASS=XXX WEBHOOK=XXX TOKEN=XXX --force-upload --no-fail-on-empty-changeset