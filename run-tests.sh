#!/bin/sh

for d in config deploy; do
  (cd $d; go test)
done
