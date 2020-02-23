pkgprefix = github.com/pbaettig/gorem-ipsum

gorem: cmd/gorem/main.go go.mod go.sum
	go build -o gorem cmd/gorem/main.go

.PHONY: test
test: internal/fifo internal/handlers
	for d in $^ ; do go test  $(pkgprefix)/$$d ; done

.PHONY: clean
clean:
	rm ./gorem

.PHONY: all
all: test gorem