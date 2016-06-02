#!/bin/bash
ps -ef | grep cropimage | grep -v grep | awk '{print $2}' | xargs kill -9
