#!/bin/bash

screen -dmAS v2ray -t inbound -c screenrc bash -ic './v2ray -config vmess_inbound.json'
screen -S v2ray -X screen -t outbound bash -ic './v2ray -config vmess_outbound.json'

