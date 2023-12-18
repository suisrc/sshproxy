#!/bin/bash

if [ ! -f '/etc/ssh/_sshd_init' ]; then
    echo 'create ssh certs to /etc/ssh ...'
    echo '' > /etc/ssh/_sshd_init
    # sshd-keygen,初始化, dsa 已经被禁用
    echo y | ssh-keygen -q -t rsa -N '' -f /etc/ssh/ssh_host_rsa_key
    echo y | ssh-keygen -q -t rsa -N '' -f /etc/ssh/ssh_host_ecdsa_key
    echo y | ssh-keygen -q -t rsa -N '' -f /etc/ssh/ssh_host_ed25519_key
fi

exec ./app
