VERSION=unversioned
TOKEN=none

ifeq (, $(shell which go.exe))
go=go
else
go=go.exe
endif

ifeq ($(filter v%,${VERSION}),${VERSION}) 
NPMVERSION=$(subst v,,${VERSION})
else
NPMVERSION=0.0.0
endif


all: build test package docs CHANGELOG.md

.PHONY: setup
setup:
	${go} mod download
	cd vscode-yolol && \
	npm install && \
	npm install -g vsce && \
	git submodule init
	git submodule update



.PHONY: test
test: go-test acid-tests vsc-test

.PHONY: go-test
go-test:
	${go} test ./...

.PHONY: acid-tests
acid-tests: yodk
	./ci/run-acid-tests.sh

.PHONY: vsc-test
vsc-test: vscode-yolol.vsix
ifdef WSLENV
	echo Skipping extension-tests on wsl
else
	cd vscode-yolol && npm test --silent
endif



.PHONY: build
build: binaries vscode-yolol.vsix

.PHONY: binaries
binaries: yodk yodk.exe yodk-darwin 

yodk: yodk-${VERSION}
	cp yodk-${VERSION} yodk
yodk-${VERSION}: $(shell find pkg) $(shell find cmd) $(shell find stdlib)
	GOOS=linux ${go} build -o yodk-${VERSION} -ldflags "-X github.com/dbaumgarten/yodk/cmd.YodkVersion=${VERSION}"

yodk.exe: yodk-${VERSION}.exe
	-cp yodk-${VERSION}.exe yodk.exe
yodk-${VERSION}.exe: $(shell find pkg) $(shell find cmd) $(shell find stdlib)
	GOOS=windows ${go} build -o yodk-${VERSION}.exe -ldflags "-X github.com/dbaumgarten/yodk/cmd.YodkVersion=${VERSION}"
	
yodk-darwin: yodk-darwin-${VERSION}
	cp yodk-darwin-${VERSION} yodk-darwin
yodk-darwin-${VERSION}: $(shell find pkg) $(shell find cmd) $(shell find stdlib)
	GOOS=darwin ${go} build -o yodk-darwin-${VERSION} -ldflags "-X github.com/dbaumgarten/yodk/cmd.YodkVersion=${VERSION}"



.PHONY: package
package: zips vscode-yolol.vsix

.PHONY: zips
zips: yodk-win.zip yodk-linux.zip yodk-darwin.zip

yodk-win.zip: yodk.exe
	zip yodk-win.zip yodk.exe

yodk-linux.zip: yodk
	zip yodk-linux.zip yodk

yodk-darwin.zip: yodk-darwin
	zip yodk-darwin.zip yodk-darwin

vscode-yolol.vsix: vscode-yolol/vscode-yolol-${NPMVERSION}.vsix
	cp vscode-yolol/vscode-yolol-${NPMVERSION}.vsix vscode-yolol.vsix

vscode-yolol/vscode-yolol-${NPMVERSION}.vsix: yodk yodk.exe yodk-darwin CHANGELOG.md $(shell find vscode-yolol/src) $(shell find vscode-yolol/syntaxes/) vscode-yolol/package.json
	cd vscode-yolol && \
	origtime=`stat -c %Y package.json` && \
	npm version --no-git-tag-version ${NPMVERSION} --allow-same-version && \
	vsce package && \
	npm version 0.0.0 --allow-same-version && \
	touch -m -d @$${origtime} package.json



CHANGELOG.md: .git/
	./ci/build-changelog.sh
	cp CHANGELOG.md vscode-yolol/



.PHONY: docs
docs: yodk yodk.exe
	./ci/build-docs.sh



publish-vsix: vscode-yolol.vsix
	vsce publish --packagePath vscode-yolol.vsix -p ${TOKEN}



.PHONY: clean
clean:
	-rm -rf yodk* *.zip *.vsix CHANGELOG.md vscode-yolol/*.vsix vscode-yolol/CHANGELOG.md vscode-yolol/bin/yo* acid_test.yaml
	-rm -rf docs/sitemap.xml docs/generated/* docs/vscode-yolol.md docs/README.md docs/nolol-stdlib.md