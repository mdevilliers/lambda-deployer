HANDLER ?= handler
PACKAGE ?= deployer

GOPATH ?= $(HOME)/go
RM := rm -f

MAKEFILE = $(word $(words $(MAKEFILE_LIST)),$(MAKEFILE_LIST))

docker: clean
	docker run --rm\
		-e HANDLER=$(HANDLER)\
		-e PACKAGE=$(PACKAGE)\
		-e GOPATH=$(GOPATH)\
		-e LDFLAGS='$(LDFLAGS)'\
		-v $(CURDIR):$(CURDIR)\
		$(foreach GP,$(subst :, ,$(GOPATH)),-v $(GP):$(GP))\
		-w $(CURDIR)\
		eawsy/aws-lambda-go-shim:latest make -f $(MAKEFILE) all

.PHONY: docker

all: build pack perm

.PHONY: all

build:
	go build -buildmode=plugin -ldflags='-w -s $(LDFLAGS)' -o $(HANDLER).so ./cmd/deployer/

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
