# Default installation directory.
#	make install dir=$HOME/bin
dir ?= /bin/
mandir ?= /usr/share/man/man1/
root ?= root
group ?= root

.PHONY: all
all: dp update-doc

dp: dp.go
	@echo Building dp...
	@go build dp.go

.PHONY: update-doc
update-doc: dp.1
	@echo Updating README.md...
	@(echo '# dp(1) - Directory Pipe';echo; COLUMNS=80 man ./dp.1 | sed 's/^/    /') > README.md

.PHONY: install
install: dp
	@echo Installing dp to ${dir}/dp...
	@install -o ${root} -g ${group} -m 755 dp ${dir}/dp
	@echo Installing dp.1 to ${mandir}/dp.1...
	@install -o ${root} -g ${group} -m 644 dp.1 ${mandir}/dp.1

.PHONY: uninstall
uninstall:
	@echo Removing ${dir}/dp...
	@rm -f ${dir}/dp
	@echo Removing ${mandir}/dp.1...
	@rm -f ${mandir}/dp.1
