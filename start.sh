#!/bin/bash

# nohup ./start.sh >std.out 2>&1 &
echo 'Start to launch a wonton agent'

./wonton proxy --config ./examples/wonton/ubuntu/init.yml
if [ "$?" = "0" ]; then
    echo 'Start successfully.'
else
    echo "Cannot start wonton!"
    exit 1
fi
