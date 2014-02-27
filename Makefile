
default: gherkin.peg.go

.PHONY: default clean test get-deps

peg=$(shell which peg \
	|| ([ -x $(GOPATH)/bin/peg ] && echo $(GOPATH)/bin/peg) \
	|| ([ -x /home/travis/gopath/bin/peg ] && echo /home/travis/gopath/bin/peg) \
	|| echo peg)

get-deps:
	go get github.com/pointlander/peg
	go get github.com/stretchr/testify/assert

gherkin.peg.go: gherkin.peg
	$(peg) -switch -inline gherkin.peg
	# dirty way to not export PEG specific types, constants or variables
	cat $@ | sed -e 's/State/state/g;s/TokenTree/tokenTree/g;s/Rul3s/rul3s/g;s/Rule/rule/g;s/END_SYMBOL/end_symbol/;' > $@.tmp
	rm $@
	mv $@.tmp $@

test: gherkin.peg.go
	go test

clean:
	- rm gherkin.peg.go
