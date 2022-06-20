#!/bin/bash

mkdir -p ./schema/testdata/$1

cd ./schema/testdata/$1

touch errors.json
touch schema.keel

cat <<EOF > errors.json
  {
    "Errors": [
      {
        "code": "",
        "hint": "",
        "message": "",
        "end_pos": {
          "column": 0,
          "filename": "testdata/$1/schema.keel",
          "line": 0,
          "offset": 0
        },
        "pos": {
          "column": 0,
          "filename": "testdata/$1/schema.keel",
          "line": 0,
          "offset": 0
        },
        "short_message": ""
      }
    ]
  }
EOF

echo "Test case created at ./schema/testdata/$1"
