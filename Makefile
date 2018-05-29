GOFILES:=$(shell find . -name '*.go' | grep -v -E '(./vendor)')

all: \
	bin/linux/reboot-agent \
	bin/linux/reboot-controller \
	#bin/darwin/reboot-agent \
	#bin/darwin/reboot-controller

images: GVERSION=$(shell $(CURDIR)/git-version.sh)
images: bin/linux/reboot-agent bin/linux/reboot-controller
	docker build -f Dockerfile-agent -t demo-reboot-agent:$(GVERSION) .
	docker build -f Dockerfile-controller -t demo-reboot-controller:$(GVERSION) .

check:
	@find . -name vendor -prune -o -name '*.go' -exec gofmt -s -d {} +
	@go vet $(shell go list ./... | grep -v '/vendor/')
	@go test -v $(shell go list ./... | grep -v '/vendor/')

.PHONY: vendor
vendor:
	glide update --strip-vendor
	glide-vc

clean:
	rm -rf bin

bin/%: LDFLAGS=-X github.com/aaronlevy/kube-controller-demo/common.Version=$(shell $(CURDIR)/git-version.sh)
bin/%: $(GOFILES)
	mkdir -p $(dir $@)
	GOOS=$(word 1, $(subst /, ,$*)) GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $@ github.com/aaronlevy/kube-controller-demo/$(notdir $@)

