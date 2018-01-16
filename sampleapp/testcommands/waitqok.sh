#!/bin/bash

COUNTER=0
while [  $COUNTER -lt 1 ]; do
        echo The counter is $COUNTER
        let COUNTER=COUNTER+1 
	sleep 1
done
