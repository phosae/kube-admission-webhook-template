TAG = kube-admission-webhook-template:local

install:
	make ensure-image
	helm install my-admission-webhook --namespace test --create-namespace ./charts

uninstall:
	helm uninstall my-admission-webhook --namespace test 

ensure-image: 
ifeq ([], $(shell docker inspect --type=image $(TAG)))
	make build-load
endif

build-load:
	docker buildx build --load -t $(TAG) .
	kind load docker-image $(TAG)