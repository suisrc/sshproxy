#!/bin/bash


## KEYS_FOLDER 环境变量不存在，默认为 /etc/ssh
if [ ! -n "$KEYS_FOLDER" ]; then
    KEYS_FOLDER='/etc/ssh'
fi

if [ ! -f "$KEYS_FOLDER/_sshd_init" ]; then
    echo "create ssh certs to $KEYS_FOLDER ..."
    echo `date` > $KEYS_FOLDER/_sshd_init
    # sshd-keygen,初始化, dsa 已经被禁用
    echo y | ssh-keygen -q -t rsa -N '' -f $KEYS_FOLDER/ssh_host_rsa_key
    echo y | ssh-keygen -q -t rsa -N '' -f $KEYS_FOLDER/ssh_host_ecdsa_key
    echo y | ssh-keygen -q -t rsa -N '' -f $KEYS_FOLDER/ssh_host_ed25519_key
fi

exec ./app
