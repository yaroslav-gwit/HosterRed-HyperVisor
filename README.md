# New version available!
This version is now archived due to new version available, which was written in Go to improve performance and developer experience.
<br>
<br>
You can a new version here:
<br>
https://github.com/yaroslav-gwit/HosterCore

# General Information
![HosterRed Logo](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/python-branch-main/screenshots/HosterRed%20Logo%20Dark.png)
Hoster is a highly opinionated VM management framework, which includes: network isolation (at the VM level), dataset encryption (at the ZFS level), instant VM deployments, storage replication between 2 or more hosts and more. It uses Python3, FreeBSD, bhyve, ZFS, and PF to achieve all of it's goals ✅🚀.</br></br>
![HosterRed Screenshot](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/python-branch-main/screenshots/HosterRed-Main-Screen-Latest.png)

# The why?
For a long time I've been a Proxmox user, swearing by it and promoting it any way I could. That stopped when I've started to rent Hetzner hardware servers, as many problems risen up: unjustifiably high RAM usage on smaller servers, no integrated NAT configuration, multiple public IP management is a nightmare, and some other minor issues like root ZFS dataset encryption on BIOS systems. That's when I discovered FreeBSD and bhyve: I realized that I could just use PF to control the traffic between the VMs and internal/external bridges, use native ZFS encryption, and do so much more but there was only 1 small problem -- vmbhyve and CBSD just weren't what I needed. Meet HosterRed -- my own automation framework (or even a small ecosystem, if you will) around FreeBSD's bhyve, PF and ZFS. As a result I can deploy the VMs in a matter of seconds on any hardware old/new/powerfull or otherwise with minimal RAM overhead.</br></br>
To network all of the nodes together as one happy family of "hosters" I use Nebula (https://www.defined.net/nebula/) and a FastAPI wrapper around it to automate the process.</br></br>
Now HosterRed is used by a couple of individuals (including myself) as their hosting platform of choise. Install it, play around with it, and you will be pleasantly surprised by the experience.</br></br>

### VM Status (state) icons
🟢 - VM is running
<br>🔴 - VM is stopped
<br>💾 - VM is a backup from another node
<br>🔒 - VM is located on the encrypted Datased
<br>🔁 - Production VM icon: VM will be included in the autostart, automatic snapshots/replication, etc

## OS Support
### List of supported OSes
- [x] Debian 11
- [x] AlmaLinux 8
- [x] RockyLinux 8
- [x] Ubuntu 20.04
- [x] Windows 10 (You'll have to provide your own image, instructions on how to build one will be released in the Wiki section soon)

### OSes on the roadmap
- [ ] FreeBSD 13 UFS
- [ ] FreeBSD 13 ZFS
- [ ] Ubuntu 20.04 LVM Hardened
- [ ] Fedora (latest)
- [ ] CentOS 7

### OSes not on the roadmap
- [x] ~~MacOS (any release)~~

# Quickstart Section
## Installation
Login as root and install bash
```
sudo su -
pkg update && pkg install -y bash curl tmux
```

The first step is optional but highly recommended. Esentially, if you ignore to set any of these values they will be generated automatically. Specifically look at the network port and ZFS encryption password:
```
export DEF_NETWORK_NAME=internal
export DEF_NETWORK_BR_ADDR=10.0.0.254
export DEF_PUBLIC_INTERFACE=igb0

export DEF_ZFS_ENCRYPTION_PASSWORD="SuperSecretRandom_password"
```

Run the installation script:
```
curl -S https://raw.githubusercontent.com/yaroslav-gwit/HosterRed-HyperVisor/python-branch-main/deploy.sh | bash
```

At the end of the installation you will receive a message like the below:
```
##### START #####

The installation is now finished.
Your ZFS encryption password: SuperSecretRandom_password
Please save your password! If you lose it, your VMs on the encrypted dataset will be lost!

Reboot the system now to apply changes.

After the reboot mount the encrypted ZFS dataset and initialize HosterRed (these 2 steps are required after each reboot):
zfs mount -a -l
hoster init

#####  END  #####
```
Save the ZFS encryption password, otherwise you'll lose your data!


## Backups
### Sheduled automatic snapshots and replication for all production VMs
```
#== AUTOMATIC SNAPSHOTS ==#
@hourly     root  hoster vm snapshot-all  --stype hourly  --keep 3
@daily      root  hoster vm snapshot-all  --stype daily   --keep 5
@weekly     root  hoster vm snapshot-all  --stype weekly  --keep 3
@monthly    root  hoster vm snapshot-all  --stype monthly --keep 6
#== AUTOMATIC REPLICATION TO OTHER NODES ==#
#20 * * * *  root  hoster vm replicate-all --ep-address 192.168.1.11
```
