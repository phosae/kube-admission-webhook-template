TAG = kube-admission-webhook-template:local

ensure-image: 
ifeq ([], $(shell docker inspect --type=image $(TAG)))
	make build-load
endif

build-load:
	docker buildx build --load -t $(TAG) .
	kind load docker-image $(TAG)