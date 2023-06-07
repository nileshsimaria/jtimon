#!/bin/bash


config="sample-config/cups.json"
if [ $# -lt 1 ]; then
	echo "Usage: ./exec <config>"
	exit 1	
fi
config=$1
./jtimon --json --config $config
echo "jtimon.log output"
