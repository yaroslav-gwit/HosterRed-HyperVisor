#!/usr/local/bin/bash

COMMAND=$1
VM_NAME=$2

echo ""
echo "🟢 INFO: NEW VM START $(date)"
echo "🔶 INFO: This bhyve command was executed:"
echo "$COMMAND"

echo ""
$COMMAND

while [[ $? == 0 ]]
do
    echo ""
    echo "🔶 INFO: The VM has been restarted: $(date)"
    $COMMAND
    sleep 1
    echo ""
done

sleep 1
echo "🔴 INFO: The VM exited on $(date)" && hoster vm kill "$VM_NAME" > /dev/null
