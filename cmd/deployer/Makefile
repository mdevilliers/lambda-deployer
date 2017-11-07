SOURCE_VERSION = $(shell git rev-parse --short=6 HEAD)
HANDLER ?= handler
PACKAGE ?= lambda-deployer

GOPATH ?= $(HOME)/go
RM := rm -f

MAKEFILE = $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))

docker: clean
	docker run --rm\
		-e HANDLER=$(HANDLER)\
		-e PACKAGE=$(PACKAGE)\
		-e GOPATH=$(GOPATH)\
		-e LDFLAGS='-X github.com/mdevilliers/lambda-deployer.SourceVersion=$(SOURCE_VERSION)"'\
		-v $(CURDIR):$(CURDIR)\
		$(foreach GP,$(subst :, ,$(GOPATH)),-v $(GP):$(GP))\
		-w $(CURDIR)\
		eawsy/aws-lambda-go-shim:latest make -f $(MAKEFILE) all

.PHONY: docker

ci: build pack

.PHONY: ci

all: build pack perm

.PHONY: all

build:
	go build -buildmode=plugin -ldflags='-w -s $(LDFLAGS)' -o $(HANDLER).so .

.PHONY: build

pack:
	pack $(HANDLER) $(HANDLER).so $(PACKAGE).zip

.PHONY: pack

perm:
	chown $(shell stat -c '%u:%g' .) $(HANDLER).so $(PACKAGE).zip

.PHONY: perm

clean:
	$(RM) $(HANDLER).so $(PACKAGE).zip

.PHONY: clean
