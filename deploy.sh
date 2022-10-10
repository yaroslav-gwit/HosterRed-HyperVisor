#!/usr/bin/env bash

#_ SET DEFAULT VARS _#
NETWORK_NAME="${DEF_NETWORK_NAME:=internal}"
NETWORK_BR_ADDR="${DEF_NETWORK_BR_ADDR:=10.0.101.254}"
PUBLIC_INTERFACE="${DEF_PUBLIC_INTERFACE:=igb0}"


#_ CHECK IF USER IS ROOT _#
if [ "$EUID" -ne 0 ]; then echo " 🚦 ERROR: Please run this script as root user!" && exit 1; fi


#_ INSTALL THE REQUIRED PACKAGES _#
pkg update -y
pkg upgrade -y
pkg install -y nano micro bash bmon iftop mc fusefs-sshfs pftop fish nginx git
pkg install -y rsync gnu-watch tmux fping qemu-utils python39 py39-devtools
pkg install -y htop curl wget gtar unzip pv sanoid cdrkit-genisoimage openssl
pkg install -y bhyve-firmware bhyve-rc grub2-bhyve uefi-edk2-bhyve uefi-edk2-bhyve-csm

if [[ -f /bin/bash ]]; then rm /bin/bash; fi
ln $(which bash) /bin/bash


#_ SET ENCRYPTED ZFS PASSWORD _#
if [ -z "${DEF_ZFS_ENCRYPTION_PASSWORD}" ]; then 
    ZFS_RANDOM_PASSWORD=`openssl rand -base64 32 | sed "s/=//g" | sed "s/\///g" | sed "s/\+//g"`
else 
    ZFS_RANDOM_PASSWORD=${DEF_ZFS_ENCRYPTION_PASSWORD}
fi


#_ SET WORKING DIRECTORY _#
HOSTER_WD="/opt/hoster-red/"


#_ PULL THE GITHUB REPO _#
if [[ ! -d ${HOSTER_WD} ]]; then
    mkdir -p ${HOSTER_WD}
    git clone https://github.com/yaroslav-gwit/PyVM-Bhyve.git ${HOSTER_WD}
else
    cd ${HOSTER_WD}
    git pull
fi
if [[ ! -f ${HOSTER_WD}.gitignore ]]; then echo "vm_images" > .gitignore; fi


#_ GENERATE SSH KEYS _#
if [[ ! -f /root/.ssh/id_rsa ]]; then ssh-keygen -b 4096 -t rsa -f /root/.ssh/id_rsa -q -N ""; else echo " 🔷 INFO: SSH key was found, no need to generate a new one"; fi
if [[ ! -f /root/.ssh/config ]]; then touch /root/.ssh/config && chmod 600 /root/.ssh/config; fi
HOST_SSH_KEY=`cat /root/.ssh/id_rsa.pub`


#_ REGISTER IF REQUIRED DATASETS EXIST _#
ENCRYPTED_DS=`zfs list | grep -c "zroot/vm-encrypted"`
UNENCRYPTED_DS=`zfs list | grep -c "zroot/vm-unencrypted"`


#_ CREATE ZFS DATASETS IF THEY DON'T EXIST _#
if [[ ${ENCRYPTED_DS} < 1 ]]
then
    zpool set autoexpand=on zroot
    zpool set autoreplace=on zroot
    zfs set primarycache=metadata zroot
    echo -e "${ZFS_RANDOM_PASSWORD}" | zfs create -o encryption=on -o keyformat=passphrase zroot/vm-encrypted
fi

if [[ ${UNENCRYPTED_DS} < 1 ]]
then
    zpool set autoexpand=on zroot
    zpool set autoreplace=on zroot
    zfs set primarycache=metadata zroot
    zfs create zroot/vm-unencrypted
fi


#_ BOOTLOADER OPTIMISATIONS _#
BOOTLOADER_FILE="/boot/loader.conf"
CMD_LINE='fusefs_load="YES"' && if [[ `grep -c ${CMD_LINE} ${BOOTLOADER_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${BOOTLOADER_FILE}; fi
CMD_LINE='vm.kmem_size="330M"' && if [[ `grep -c ${CMD_LINE} ${BOOTLOADER_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${BOOTLOADER_FILE}; fi
CMD_LINE='vm.kmem_size_max="330M"' && if [[ `grep -c ${CMD_LINE} ${BOOTLOADER_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${BOOTLOADER_FILE}; fi
CMD_LINE='vfs.zfs.arc_max="40M"' && if [[ `grep -c ${CMD_LINE} ${BOOTLOADER_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${BOOTLOADER_FILE}; fi
CMD_LINE='vfs.zfs.vdev.cache.size="5M"' && if [[ `grep -c ${CMD_LINE} ${BOOTLOADER_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${BOOTLOADER_FILE}; fi


#_ PF CONFIG BLOCK IN rc.conf _#
RC_CONF_FILE="/etc/rc.conf"
CMD_LINE='pf_enable="yes"' && if [[ `grep -c ${CMD_LINE} ${RC_CONF_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${RC_CONF_FILE}; fi
CMD_LINE='pf_rules="/etc/pf.conf"' && if [[ `grep -c ${CMD_LINE} ${RC_CONF_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${RC_CONF_FILE}; fi
CMD_LINE='pflog_enable="yes"' && if [[ `grep -c ${CMD_LINE} ${RC_CONF_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${RC_CONF_FILE}; fi
CMD_LINE='pflog_logfile="/var/log/pflog"' && if [[ `grep -c ${CMD_LINE} ${RC_CONF_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${RC_CONF_FILE}; fi
CMD_LINE='pflog_flags=""' && if [[ `grep -c ${CMD_LINE} ${RC_CONF_FILE}` < 1 ]]; then echo ${CMD_LINE} >> ${RC_CONF_FILE}; fi


#_ SET CORRECT PROFILE FILE _#
cat << 'EOF' | cat > /root/.profile
# $FreeBSD$
# This is a .profile template for FreeBSD Bhyve Hosters
PATH=/sbin:/bin:/usr/sbin:/usr/bin:/usr/local/sbin:/usr/local/bin:~/bin
export PATH
HOME=/root
export HOME
TERM=${TERM:-xterm}
export TERM
PAGER=less
export PAGER

# set ENV to a file invoked each time sh is started for interactive use.
ENV=$HOME/.shrc; export ENV

# Query terminal size; useful for serial lines.
if [ -x /usr/bin/resizewin ] ; then /usr/bin/resizewin -z ; fi

# Uncomment to display a random cookie on each login.
# if [ -x /usr/bin/fortune ] ; then /usr/bin/fortune -s ; fi

export EDITOR=micro
EOF

cat << EOF | cat >> /root/.profile
export EDITOR=micro
EOF


#_ GENERATE MINIMAL REQUIRED CONFIG FILES _#
### NETWORK CONFIG ###
cat << EOF | cat > ${HOSTER_WD}configs/networks.json
{
    "networks": [
        {
            "bridge_name": "${NETWORK_NAME}",
            "bridge_address": "${NETWORK_BR_ADDR}",
            "bridge_interface": "None",
            "apply_bridge_address": true,
            "range_start": 10,
            "range_end": 200,
            "bridge_subnet": 24,
            "comment": "Internal Network"
        }
}
EOF

### HOST CONFIG ###
cat << EOF | cat > ${HOSTER_WD}configs/host.json
{
    "backup_servers": [
    ],
    "host_dns_acls": [
      "${NETWORK_BR_ADDR}/24"
    ],
    "host_ssh_keys": [
        {
            "key_value": "${HOST_SSH_KEY}",
            "comment": "Host Key"
        }
    ]
}
EOF

### ZFS DS CONFIG ###
cat << EOF | cat > ${HOSTER_WD}configs/datasets.json
{
    "datasets": [
        {"name": "zfs-encrypted",
            "id": "0",
            "type": "zfs",
            "mount_path": "/zroot/vm-encrypted/",
            "zfs_path": "zroot/vm-encrypted",
            "encrypted": true,
            "comment": "ZFS Encrypted"
        },
        {"name": "zfs-unencrypted",
            "id": "1",
            "type": "zfs",
            "mount_path": "/zroot/vm-unencrypted/",
            "zfs_path": "zroot/vm-unencrypted",
            "encrypted": false,
            "comment": "ZFS Unecrypted"
        }
    ]
}
EOF


#_ INIT PYTHON ENV _#
cd ${HOSTER_WD}
if [[ ! -d venv ]]; then python3 -m venv venv; fi
${HOSTER_WD}venv/bin/python3 -m ensurepip
${HOSTER_WD}venv/bin/python3 -m pip install --upgrade pip
${HOSTER_WD}venv/bin/python3 -m pip install -r requirements.txt --upgrade


#_ COPY OVER UNBOUND CONFIG _#
cat << 'EOF' | cat > /var/unbound/unbound.conf
# This file is automatically generated by HosterRed: HyperVisor.
# Modifications will be overwritten.
server:
        username: unbound
        directory: /var/unbound
        chroot: /var/unbound
        pidfile: /var/run/local_unbound.pid
        auto-trust-anchor-file: /var/unbound/root.key
        
        interface: 0.0.0.0
        access-control: 127.0.0.0/8 allow
        
include: /var/unbound/forward.conf
include: /var/unbound/lan-zones.conf
include: /var/unbound/control.conf
include: /var/unbound/conf.d/*.conf
EOF


#_ COPY OVER NGINX CONFIG _#
cat << 'EOF' | cat > /usr/local/etc/nginx/nginx.conf
load_module /usr/local/libexec/nginx/ngx_stream_module.so;
user  nobody;
worker_processes  5;

error_log  /var/log/nginx/error.log;

events {
    worker_connections  1024;
}

stream {
# EXAMPLE_RECORDS
#    server {
#        listen 80;
#        proxy_pass 10.0.0.10:80;
#        proxy_buffer_size 16k;
#    }
#    server {
#        listen 443;
#        proxy_pass 10.0.0.10:443;
#        proxy_buffer_size 16k;
#   }
}
EOF


#_ COPY OVER PF CONFIG _#
cat << EOF | cat > /etc/pf.conf
table <private-ranges> { 0.0.0.0/8 10.0.0.0/8 100.64.0.0/10 127.0.0.0/8 169.254.0.0/16 172.16.0.0/12 192.0.0.0/24 192.0.0.0/29 192.0.2.0/24 \
                         192.88.99.0/24 192.168.0.0/16 198.18.0.0/15 198.51.100.0/24 203.0.113.0/24 240.0.0.0/4 255.255.255.255/32 }

set skip on lo0
scrub in all fragment reassemble max-mss 1440


### OUTBOUND NAT ###
nat on { ${PUBLIC_INTERFACE} } from { ${NETWORK_BR_ADDR}/24 } to any -> { ${PUBLIC_INTERFACE} }


### INBOUND NAT EXAMPLES ###
#rdr pass on { ${PUBLIC_INTERFACE} } proto { tcp udp } from any to EXTERNAL_INTERFACE_IP_HERE port 28967 -> 10.0.0.3 port 28967
#rdr pass on { ${PUBLIC_INTERFACE} } proto tcp from any to EXTERNAL_INTERFACE_IP_HERE port 14000 -> 10.0.0.3 port 14002
#rdr pass on { ${PUBLIC_INTERFACE} } proto tcp from any to 1.12.13.14 port { 80 443 } -> 10.0.0.10 # Inline comments go here


### ANTISPOOF RULE ###
antispoof quick for { ${PUBLIC_INTERFACE} } # DISABLE IF USING ANY ADDITIONAL ROUTERS IN THE VM, LIKE OPNSENSE


### FIREWALL RULES ###
#block in quick log on egress from <private-ranges>
#block return out quick on egress to <private-ranges>
block in all
pass out all keep state

# Allow internal NAT networks to go out + examples #
#pass in proto tcp to port 5900:5950 keep state
#pass in quick inet proto { tcp udp icmp } from { ${NETWORK_BR_ADDR}/24 } to any # Uncomment this rule to allow any traffic out
pass in quick inet proto { udp } from { ${NETWORK_BR_ADDR}/24 } to { ${NETWORK_BR_ADDR} } port 53
block in quick inet from { ${NETWORK_BR_ADDR}/24 } to <private-ranges>
pass in quick inet proto { tcp udp icmp } from { ${NETWORK_BR_ADDR}/24 } to any


### INCOMING HOST RULES ###
pass in quick on { ${NETWORK_BR_ADDR} } inet proto icmp all # allow PING in
pass in quick on { ${NETWORK_BR_ADDR} } proto tcp to port 22 keep state #ALLOW_SSH_ACCESS_TO_HOST
#pass in proto tcp to port 80 keep state #HTTP_NGINX_PROXY
#pass in proto tcp to port 443 keep state #HTTPS_NGINX_PROXY
EOF


#_ CREATE AN EXECUTABLE HOSTER FILE _#
cd ${HOSTER_WD}
if [[ -f /bin/hoster ]]; then rm -f /bin/hoster; fi
ln hoster /bin/hoster
chmod +x /bin/hoster


#_ LET USER KNOW THE STATE OF DEPLOYMENT _#
cat << EOF | cat


##### START #####

The installation is now finished.
Your ZFS encryption password: ${ZFS_RANDOM_PASSWORD}
Please save your password! If you lose it, your VMs on the encrypted dataset will be lost!

Reboot the system now to apply changes.

After the reboot mount the encrypted ZFS dataset and initialize HosterRed (these 2 steps are required after each reboot):
zfs mount -a -l
hoster init

#####  END  #####

EOF
