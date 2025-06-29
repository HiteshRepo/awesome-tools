GODOC_PORT ?= 6060

.PHONY: help install-godoc serve-doc gen-doc

help:
	@echo "Usage:"
	@echo "  make install-godoc     Install godoc tool"
	@echo "  make serve-doc         Serve documentation at http://localhost:$(GODOC_PORT)"

install-godoc:
	go install golang.org/x/tools/cmd/godoc@latest

gen-doc:
	go doc ./...

serve-doc:
	@echo "Serving documentation at http://localhost:$(GODOC_PORT)"
	@echo "Press Ctrl+C to stop."
	godoc -http=localhost:$(GODOC_PORT)
