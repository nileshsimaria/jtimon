#!/bin/bash


config="sample-config/cups.json"
if [ $# -ge 1 ]; then
	case "$1" in
		"ctl")
			;;
		"sgrp")
			;;
		"up")
			;;
		*)
			echo "sensor identifiers:\nctl, sgrp, up"
			exit 1
			;;
	esac
	config="sample-config/cups-$1.json"
	echo "Subscribing to $1 sensors..."
else
	echo "Subscribing to all CUPS sensors..."
fi

./jtimon --config $config --json
echo "jtimon.log output"
