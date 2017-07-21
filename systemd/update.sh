#!/usr/bin/env bash

curl -Lo /opt/aerokubot/aerokubot https://github.com/aerokube/aerokubot/releases/download/$1/aerokubot && chmod +x /opt/aerokubot/aerokubot

curl -Lo /etc/systemd/system/aerokubot.service https://raw.githubusercontent.com/aerokube/aerokubot/master/systemd/aerokubot.service

systemctl enable aerokubot
