
BUILD_NUMBER ?= 0

SOURCEDIR=.
SOURCES := $(shell find $(SOURCEDIR) -name '*.go')

bin/datadog-podsizer: $(SOURCES)
	go build -ldflags "-X main.build=${BUILD_NUMBER}" -o bin/datadog-podsizer

run: bin/datadog-podsizer
	bin/datadog-podsizer -r test-release -c test-container
