#!/bin/bash

function to_csv() {
  grep TPS "$1.log" | awk -F '[ :,]+' '{split($2,title,"m");split($6,a,"m");split($8,b,"m");print title[3]","substr($11,3)","substr($9,4)","$4","a[1]","b[1]}' >"$1.csv"
}

endpoint=("server" "client")

for ((j = 0; j < ${#endpoint[@]}; j++)); do
  ep=${endpoint[j]}
  for ((i = 1; i <= 3; i++)); do
    filename="${ep}${i}"
    # benchmark
    echo "./scripts/benchmark_${ep}.sh ${i} > ${filename}.log"
    ./scripts/benchmark_${ep}.sh ${i} >"${filename}.log"

    # update data & images
    to_csv "${filename}"
    echo "python3 scripts/reports/render_images.py ${filename}"
    python3 scripts/reports/render_images.py "${filename}"
    mv "${filename}_qps.png" "${filename}_tp99.png" "${filename}_tp999.png" docs/images/

    # clean
    # rm "${filename}.log" "${filename}.csv"
  done
done

# not used
function to_table() {
  grep TPS output.log | awk -F '[ :,]+' '{print "<tr><td> "$2" </td><td> 传输 </td><td> "$4" </td><td> "$6" </td><td> "$8" </td></tr>"}' >"$1.txt"
}
