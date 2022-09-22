#!/bin/bash
gobinsec -wait -cache -config utils/gobinsec_conf.yml build/bridge || FAILED=true
if [ $FAILED ]; then
  gobinsec -wait -cache -config utils/gobinsec_conf.yml build/bridge
fi