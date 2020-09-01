grokkingiot1: dev_iot_private_key_check axis axispwd
	docker build -t local/grokkingiot1 .
	docker stop grokkingiot1 || true
	docker rm grokkingiot1 || true
	docker run -d --name grokkingiot1 --restart unless-stopped -e DEV_IOT_PRIVATE_KEY -e AXIS -e AXISPWD -v /opt/grokkingiot1:/config -t local/grokkingiot1 -conf config/provision-dev.json -env dev

dev_iot_private_key_check:
ifndef DEV_IOT_PRIVATE_KEY
	$(error 'DEV_IOT_PRIVATE_KEY envvar is undefined -- run "source integration_tests_go/iot/.envrc" to set)
endif

axis:
ifndef AXIS
	$(error 'AXIS envvar is undefined -- run "source integration_tests_go/iot/.envrc" to set)
endif

axispwd:
ifndef AXISPWD
	$(error 'AXISPWD envvar is undefined -- run "source integration_tests_go/iot/.envrc" to set)
endif