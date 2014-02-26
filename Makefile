
.default: gherkin.peg.go

.PHONY: clean test get-deps

get-deps:
	go get github.com/pointlander/peg

gherkin.peg.go: gherkin.peg
	[ -n "$$(which peg)" ] || (echo "Error: missing peg executable.\n\
	    try: go get github.com/pointlander/peg"; exit 1)
	peg -switch -inline gherkin.peg
	# dirty way to not export PEG specific types, constants or variables
	cat gherkin.peg.go | sed -e 's/State/state/g;s/TokenTree/tokenTree/g;s/Rul3s/rul3s/g;s/Rule/rule/g;s/END_SYMBOL/end_symbol/;' > gherkin.peg.go.tmp
	rm gherkin.peg.go; mv gherkin.peg.go{.tmp,}

test: gherkin.peg.go
	go test

clean:
	- rm gherkin.peg.go
