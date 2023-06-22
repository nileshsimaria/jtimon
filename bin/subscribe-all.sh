#!/bin/bash

./bin/update-config-list.sh
rm -rf cp-logs/* up-logs/*
if [ $# -lt 1 ]; then
	echo "Usage: ./exec <config-list-file>"
	exit 1	
fi
echo "Parsing config list..."
filename=$1
mapfile -t config_files < "$filename"
cmd=(./jtimon-linux-amd64 --json)
for config_file in "${config_files[@]}"; do
	cmd+=("--config" "$config_file")
done

"${cmd[@]}"

echo "outputs in cp-logs/ and up-logs/"

sudo chmod -R +r ./cp-logs
sudo chmod -R +r ./up-logs
