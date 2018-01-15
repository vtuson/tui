#!/bin/bash

if [ "$#" -ne 5 ]; then
    echo "Illegal number of parameters"
    echo $@> ./testcommands/tmp
    exit 1
fi
echo $@> ./testcommands/tmp
