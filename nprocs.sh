#!/bin/bash

if [[ "$OSTYPE" == "linux-gnu"* ]]; then
  nproc
elif [[ "$OSTYPE" == "darwin"* ]]; then
  sysctl -n hw.ncpu
else
  echo "Unsupported OS" >& 2
  echo 1
  exit 1
fi
