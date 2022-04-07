#!/bin/bash
#Install script to install go programs

#Build Program
go build diygpufanctrl.go

#moving file and config to locations
sudo cp diygpufanctrl /usr/local/bin
sudo cp diygpufanctrlconfigs.json /etc

sudo cp diygpufanctrl.service /etc/systemd/system
sudo chmod 640 /etc/systemd/system/diygpufanctrl.service
sudo systemctl daemon-reload
sudo systemctl enable diygpufanctrl
sudo systemctl start diygpufanctrl