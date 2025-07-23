#!/usr/bin/env bash


trap 'env | grep INJECTION' SIGHUP

# shellcheck disable=SC2034
for i in {1..10} ; do
    env | grep INJECTION
    echo "Going to sleep"
    cat $INJECTION
    echo
    sleep 10
done
