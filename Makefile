COMMIT_HASH = $(shell git rev-parse --short HEAD)
COMMIT = $(shell git rev-parse HEAD)
RET = $(shell git describe --contains $(COMMIT_HASH) 1>&2 2> /dev/null; echo $$?)
PWD = $(shell pwd)
USER = $(shell whoami)
buildTime = $(shell date +%Y-%m-%dT%H:%M:%S%z)
PROJ_NAME = twcc-self-checker
DOCKER_REPO = ogre0403
RELEASE_TAG = v0.2.1

ifeq ($(RET),0)
    TAG = $(shell git describe --contains $(COMMIT_HASH))
else
    TAG = $(USER)-$(COMMIT_HASH)
endif


run:
	rm -rf bin/${PROJ_NAME}
	go build  -mod vendor -ldflags '-X "main.buildTime='"${buildTime}"'" -X "main.commitID='"${COMMIT}"'"'  -o bin/${PROJ_NAME} cmd/main.go
	./bin/${PROJ_NAME}  -v=1


run-in-docker:
	docker run -ti --rm  ${DOCKER_REPO}/${PROJ_NAME}:$(TAG)


build-img:
	docker build -t ${DOCKER_REPO}/${PROJ_NAME}:$(RELEASE_TAG) .
#	docker tag ${DOCKER_REPO}/${PROJ_NAME}:$(TAG) ${DOCKER_REPO}/${PROJ_NAME}:$(RELEASE_TAG)


build-in-docker:
	rm -rf bin/*
	CGO_ENABLED=0 GOOS=linux go build  -mod vendor -ldflags '-X "main.buildTime='"${buildTime}"'" -X "main.commitID='"${COMMIT}"'"' -a -installsuffix cgo -o bin/${PROJ_NAME} cmd/main.go


clean:
	rm -rf bin/*