#!/bin/bash
#Install script to install go programs

#Build Program
go build hpcgpufanctrl.go

#making new directory for fan control
mkdir $HOME/.diygpufanctrl

mv hpcgpufanctrl $HOME/.diygpufanctrl
mv config.json $HOME/.diygpufanctrl

echo -e "PATH=$PATH:$HOME/.diygpufanctrl" >> $HOME/.profile
