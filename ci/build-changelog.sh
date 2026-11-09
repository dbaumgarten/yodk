#!/bin/bash
set -e
NUM_TAGS=${1:-10}
OUTFILE=${2:-CHANGELOG.md}
TAGS=`git tag | sort -V -r | head -n $((NUM_TAGS+1))`
TAG1=""
TAG2=""
echo -e "# Auto-generated changelog\n\n" > $OUTFILE
while read line; do 
    TAG1=$TAG2
    TAG2=$line
    if [ "$TAG1" != "" ] ; then
        echo "## $TAG1" >> $OUTFILE
        git log '--format=format: - %s' $TAG2..$TAG1 >> $OUTFILE
        echo -e "\n" >> $OUTFILE
    fi
done <<< "$TAGS"
