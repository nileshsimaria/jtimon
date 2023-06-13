outfile="config-list.txt"
directories=("./up-configs" "./cp-configs")

truncate -s 0 $outfile
for ((i=0; i<${#directories[@]}; i++)); do
	for file_name in "${directories[$i]}"/*; do
		base_name=$(basename "${file_name}")
		echo "${directories[$i]}/${base_name}" >> $outfile
	done
done
