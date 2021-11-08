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
./yodk docs stdlib/src/*.nolol -n 'std/$1' -r 'stdlib/src/([^_]*)(_.*)?.nolol' -c professional >> docs/nolol-stdlib.md

echo Generating sitemap
ROOTURL="https://dbaumgarten.github.io/yodk/#/"
IGNORE_IN_SITEMAP="README.md,nolol-stdlib-header.md,_sidebar.md"

cd docs
cat << EOF > sitemap_new.xml
<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
<url>
      <loc>${ROOTURL}</loc>
      <lastmod>$(date -r README.md "+%Y-%m-%dT%H:%M:%S%:z")</lastmod>
</url>
EOF

for FILE in *.md; do
if ! echo ${IGNORE_IN_SITEMAP} | grep -q ${FILE}; then
CHANGEDATE=$(date -r ${FILE} "+%Y-%m-%dT%H:%M:%S%:z")
cat << EOF >> sitemap_new.xml
<url>
      <loc>${ROOTURL}${FILE}</loc>
      <lastmod>${CHANGEDATE}</lastmod>
</url>
EOF
fi
done

cat << EOF >> sitemap_new.xml
</urlset> 
EOF

cd ..


chmod -R a+r docs
