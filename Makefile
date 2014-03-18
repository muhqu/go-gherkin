
default: gherkin.peg.go

.PHONY: default clean test integration get-deps

peg=$(shell which peg \
	|| ([ -x $(GOPATH)/bin/peg ] && echo $(GOPATH)/bin/peg) \
	|| ([ -x /home/travis/gopath/bin/peg ] && echo /home/travis/gopath/bin/peg) \
	|| echo peg)

get-deps:
	go get github.com/pointlander/peg
	go get github.com/stretchr/testify/assert

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

test: gherkin.peg.go
	go test

integration:
	@$(MAKE) test
	@$(MAKE) clean test

clean:
	- rm gherkin.peg.go
