#!/bin/bash
set -e
TAGS=`git tag | sort -V -r | head -n 10`
TAG1=""
TAG2=""
echo -e "# Auto-generated changelog\n\n" > CHANGELOG.md
while read line; do 
    TAG1=$TAG2
    TAG2=$line
    if [ "$TAG1" != "" ] ; then
        echo "## $TAG1" >> CHANGELOG.md
        git log '--format=format: - %s' $TAG2..$TAG1 >> CHANGELOG.md
        echo -e "\n" >> CHANGELOG.md
    fi
done <<< "$TAGS"

cat CHANGELOG.md