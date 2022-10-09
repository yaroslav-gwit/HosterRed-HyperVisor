# General Information
![HosterRed Logo](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/python-branch-main/screenshots/HosterRed%20Logo%20Dark.png)
HosterRed: HyperVisor is a highly opinionated VM management framework, which includes: network isolation (at the VM level), dataset encryption (at the ZFS level), instant VM deployments, storage replication between 2 or more hosts and more. It uses Python3, FreeBSD, bhyve, ZFS, and PF to achieve all of it's goals ✅🚀.</br></br>
![HosterRed Screenshot](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/python-branch-main/screenshots/HosterRed-Main-Screen.png)

# The why?
For a long time I've been a Proxmox user, swearing by it and promoting it any way I could. That stopped when I've started to rent Hetzner hardware servers, as many problems risen up: unjustifiably high RAM usage on smaller servers, no integrated NAT configuration, multiple public IP management is a nightmare, and some other minor issues like root ZFS dataset encryption on BIOS systems. That's when I discovered FreeBSD and bhyve: I realized that I could just use PF to control the traffic between the VMs and internal/external bridges, use native ZFS encryption, and do so much more but there was only 1 small problem -- vmbhyve and CBSD just weren't what I needed. Meet HosterRed -- my own automation framework (or even a small ecosystem, if you will) around FreeBSD's bhyve, PF and ZFS. As a result I can deploy the VMs in a matter of seconds on any hardware old/new/powerfull or otherwise with minimal RAM overhead.</br></br>
To network all of the nodes together as one happy family of "hosters" I use Nebula (https://www.defined.net/nebula/).</br></br>
Now HosterRed is used by a couple of individuals (including myself) as their hosting platform of choise. Install it, play around with it, and you will be pleasantly surprised by the experience.</br></br>
P.S. WebUI is coming too, stay tuned for that 😉

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
- [x] Ubuntu 20.04
- [x] FreeBSD 13 UFS
- [x] FreeBSD 13 ZFS
- [x] Windows 10 (You'll have to provide your own image, instructions on how to build one will be released in the Wiki section soon)

### OSes on the roadmap
- [ ] Ubuntu 20.04 LVM Hardened
- [ ] Fedora (latest)
- [ ] CentOS 7
- [ ] OpenBSD
- [ ] OpenSUSE Leap
- [ ] OpenSUSE Tumbleweed
- [ ] Windows 11
- [ ] Windows Server 2019

### OSes not on the roadmap
- [x] ~~MacOS (any release)~~

# Quickstart Section
## Backups
### Sheduled automatic snapshots and replication for all production VMs
```
@hourly     root  hoster vm snapshot-all  --stype hourly  --keep 3
@daily      root  hoster vm snapshot-all  --stype daily   --keep 5
@weekly     root  hoster vm snapshot-all  --stype weekly  --keep 3
@monthly    root  hoster vm snapshot-all  --stype monthly --keep 6
20 * * * *  root  hoster vm replicate-all --ep-address 192.168.1.11
```
