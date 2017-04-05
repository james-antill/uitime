all:
		@echo Building uitime
		@go build uitime.go
fmt:
		@gofmt -s uitime.go > uitiem.go.tmp
		@diff -uw uitime.go uitiem.go.tmp
		@rm -f uitiem.go.tmp
clean:
		@rm -f uitiem.go.tmp uitime *~
