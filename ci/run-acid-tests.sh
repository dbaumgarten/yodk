#!/bin/bash
set -e

for FILE in acid-tests/*.yolol; do

echo Testing $FILE
cat << EOF > acid_test.yaml
scripts: 
  - $FILE
ignoreerrs: true
cases:
  - name: TestOutput
    outputs:
      OUTPUT: "ok"
EOF
./yodk test acid_test.yaml
rm acid_test.yaml

done

