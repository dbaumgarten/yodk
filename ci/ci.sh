#!/bin/bash

PHASE=$1
VERSION="${2:-unversioned}"

go="go"
if ! which go; then
    go="go.exe"
fi

if [ "$PHASE" == "" ]; then
    echo "Usage: ./ci/ci.sh <phase> ( [version] | [vsce_token] )+"
    exit 1
fi

set -ex

if [ "$PHASE" == "clean" ]; then
    rm -rf yodk* *.zip *.vsix CHANGELOG.md vscode-yolol/*.vsix vscode-yolol/CHANGELOG.md vscode-yolol/bin/yo* acid_test.yaml|| true
    rm -rf docs/sitemap.xml docs/generated/* docs/vscode-yolol.md docs/README.md docs/nolol-stdlib.md || true

elif [ "$PHASE" == "install" ]; then
    $go mod download
    cd vscode-yolol
    npm install
    npm install -g vsce
    cd ..
    git submodule init
    git submodule update

elif [ "$PHASE" == "build" ]; then
    GOOS=linux $go build -o yodk -ldflags "-X github.com/dbaumgarten/yodk/cmd.YodkVersion=${VERSION}"
    GOOS=windows $go build -o yodk.exe -ldflags "-X github.com/dbaumgarten/yodk/cmd.YodkVersion=${VERSION}"
    GOOS=darwin $go build -o yodk-darwin -ldflags "-X github.com/dbaumgarten/yodk/cmd.YodkVersion=${VERSION}"
    cd vscode-yolol
    npm run vscode:prepublish
    cd ..

elif [ "$PHASE" == "test" ]; then
    $go test ./...
    ./ci/run-acid-tests.sh
    cd vscode-yolol
    npm test --silent
    cd ..

elif [ "$PHASE" == "prepublish" ]; then
  ./ci/build-changelog.sh
  cp CHANGELOG.md vscode-yolol/
  cd vscode-yolol
  VERSION=$(echo ${VERSION} | tr -d v)
  if ! npm version --no-git-tag-version ${VERSION} --allow-same-version; then
    echo No valid version. Using v0.0.0 instead
    export VERSION=0.0.0
    npm version --no-git-tag-version ${VERSION} --allow-same-version
  fi
  vsce package
  npm version 0.0.0 --allow-same-version
  cp vscode-yolol-${VERSION}.vsix ../vscode-yolol.vsix
  cd ..
  zip yodk-win.zip yodk.exe
  zip yodk-linux.zip yodk
  zip yodk-darwin.zip yodk-darwin
  ./ci/build-docs.sh

elif [ "$PHASE" == "publish" ]; then
  if [ "$2" == "" ]; then
    echo "Usage: ./ci/ci.sh publish [vsce_token]"
    exit 1
  fi
  vsce publish --packagePath vscode-yolol.vsix -p $2
fi