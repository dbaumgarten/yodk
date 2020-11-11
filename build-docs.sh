#!/bin/bash
set -e

# This script automatically creates the content of the docs-folder (at least some of it) from other files in this repo
# It is automatically run by the ci-tool
# This way, duplicatin between repo and documentation is reduced massively and the documentation is automatically kept up-to-date.


if [ -f "./yodk" ]; then
    YODK_BINARY="./yodk"
elif [ -f "./yodk.exe" ]; then
    YODK_BINARY="./yodk.exe"
else
    echo No compiled yodk binary found
    exit 1
fi

rm -rf docs/generated || true
mkdir -p docs/generated/code/yolol
mkdir -p docs/generated/code/nolol
mkdir -p docs/generated/cli/
mkdir -p docs/generated/tests

cp examples/yolol/*.yolol docs/generated/code/yolol
cp examples/nolol/*.nolol docs/generated/code/nolol
cp examples/yolol/fizzbuzz_test.yaml docs/generated/tests

${YODK_BINARY} compile docs/generated/code/nolol/*.nolol
${YODK_BINARY} format docs/generated/code/nolol/*.nolol
${YODK_BINARY} format docs/generated/code/yolol/*.yolol
${YODK_BINARY} optimize docs/generated/code/yolol/unoptimized.yolol

cp vscode-yolol/README.md docs/vscode-yolol.md
cp README.md docs/README.md
sed -i 's/https:\/\/dbaumgarten.github.io\/yodk\/#//g' docs/README.md
echo "help" | ${YODK_BINARY} debug examples/yolol/fizzbuzz.yolol | grep -v EOF > docs/generated/cli/debug-help.txt

chmod -R a+r docs
