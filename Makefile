
default: build

GIT_VERSION=$(shell git describe HEAD 2>/dev/null || git describe --tags HEAD)

.PHONY: default version clean build test install integration get-deps

peg=$(shell which peg \
	|| ([ -x $(GOPATH)/bin/peg ] && echo $(GOPATH)/bin/peg) \
	|| ([ -x /home/travis/gopath/bin/peg ] && echo /home/travis/gopath/bin/peg) \
	|| echo peg)

get-deps:
	go get github.com/pointlander/peg
	go get github.com/stretchr/testify/assert
	go get github.com/pebbe/util

gherkin.peg.go: gherkin.peg
	
	# pre-process peg file
	cat gherkin.peg | sed \
	  -e 's/{{++\([^}]*\)}}/{ p.\1 = p.\1 + buffer[begin:end] }/g' \
	  -e 's/{{\[\]\([^}]*\)}}/{ p.\1 = append(p.\1, buffer[begin:end]) }/g' \
	  -e 's/{{\([^}]*\)}}/{ p.\1 = buffer[begin:end] }/g' \
	  > gherkin.peg.pp

	$(peg) -switch -inline gherkin.peg.pp

	# dirty way to not export PEG specific types, constants or variables
	cat gherkin.peg.pp.go | sed \
	  -e 's/State/state/g' \
	  -e 's/TokenTree/tokenTree/g' \
	  -e 's/Rul3s/rul3s/g' \
	  -e 's/Rule/rule/g' \
	  -e 's/END_SYMBOL/end_symbol/' \
	  > $@
	rm gherkin.peg.pp gherkin.peg.pp.go

version: 
	@echo "version: $(GIT_VERSION)" >&2
	@cat version.go | sed -e 's/\(VERSION = "\)[^\"]*\("\)/\1'$(GIT_VERSION)'\2/' > version.go.tmp
	@diff version.go.tmp version.go || (cat version.go.tmp > version.go)
	@rm version.go.tmp

build: version gherkin.peg.go
	go build ./ ./formater ./cmd/gherkinfmt

install: version gherkin.peg.go
	go install ./cmd/gherkinfmt

test: version gherkin.peg.go
	go test ./ ./formater ./cmd/gherkinfmt

integration: get-deps clean test
	@echo "done: $(GIT_VERSION)" >&2

clean:
	- rm gherkin.peg.go
