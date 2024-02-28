DOCKER_USER = phosae
TAG = $(DOCKER_USER)/kube-admission-webhook-template:local

install:
	make ensure-image
	helm install my-admission-webhook --namespace test --create-namespace ./charts --set-string image=$(TAG)

uninstall:
	helm uninstall my-admission-webhook --namespace test 

ensure-image: 
ifeq ([], $(shell docker inspect --type=image $(TAG)))
	make build-load
endif

build-load:
	docker buildx build --load -t $(TAG) .
	kind load docker-image $(TAG)	

ko-build-load: ko
	KO_DOCKER_REPO=$(DOCKER_USER) KO_DEFAULTBASEIMAGE=ubuntu:jammy KOCACHE=/tmp/ko \
		ko build . --platform linux/amd64 -B -t local --local --push=false
	kind load docker-image $(TAG)

ko:
ifeq (, $(shell which ko))
	GOBIN=/usr/local/bin/ go install github.com/google/ko@v0.15.2
endif