#!/bin/bash
set -e

# This script automatically creates the content of the docs-folder (at least some of it) from other files in this repo
# It is automatically run by the ci-tool
# This way, duplicatin between repo and documentation is reduced massively and the documentation is automatically kept up-to-date.


rm -rf docs/generated || true
mkdir -p docs/generated/code/yolol
mkdir -p docs/generated/code/nolol
mkdir -p docs/generated/cli/
mkdir -p docs/generated/tests

cp examples/yolol/*.yolol docs/generated/code/yolol
cp examples/nolol/*.nolol docs/generated/code/nolol
cp examples/yolol/fizzbuzz_test.yaml docs/generated/tests

./yodk compile docs/generated/code/nolol/*.nolol
./yodk format docs/generated/code/nolol/*.nolol
./yodk format docs/generated/code/yolol/*.yolol
./yodk optimize docs/generated/code/yolol/unoptimized.yolol

cp vscode-yolol/README.md docs/vscode-yolol.md
cp README.md docs/README.md
sed -i 's/https:\/\/dbaumgarten.github.io\/yodk\/#//g' docs/README.md
echo "help" | ./yodk debug examples/yolol/fizzbuzz.yolol | grep -v EOF > docs/generated/cli/debug-help.txt

cat docs/nolol-stdlib-header.md > docs/nolol-stdlib.md
./yodk docs stdlib/src/*.nolol -n 'std/$1' -r 'stdlib/src/(.*).nolol' >> docs/nolol-stdlib.md

chmod -R a+r docs
