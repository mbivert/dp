.PHONY: all
all: dp update-doc

dp: dp.go
	@echo Building dp...
	@go build dp.go

.PHONY: update-doc
update-doc: dp.1
	@echo Updating README.md...
	@(echo '# dp(1) - Directory Pipe';echo; man ./dp.1 | sed 's/^/    /') > README.md
