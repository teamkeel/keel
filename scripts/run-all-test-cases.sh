#!/bin/bash

for dir in schema/testdata/*
do
  find="schema/testdata/"
  replace=""
  test_case_name=${dir//$find/$replace}
  echo "Running $test_case_name"

  (set -x; go run . validate -f "$dir/schema.keel")
done