#!/bin/bash

if [ "$#" -ne 5 ]; then
    echo "Illegal number of parameters"
    echo $@> ./tmp
    exit 1
fi
echo $@> ./tmp
