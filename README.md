# General Information
Hoster Red: HyperVisor is based on FreeBSD, Bhyve and ZFS. It's a new approach to VM deployment and management.
![HosterRed Screenshot 3](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/main/screenshots/HosterRed_screenshot_3.png)

### State icons
🟢 - VM is running
<br>🔴 - VM is stopped
<br>🔶 - VM is a backup from another node
<br>🔒 - VM is located on the encrypted Datased
<br>🔁 - Production VM icon: VM will be included in the autostart, automatic snapshots/replication, etc

## OS Support
### List of supported OSes
- [x] Debian 11
- [x] Ubuntu 20.04
- [x] AlmaLinux 8
- [x] Windows 10 (You'll have to provide your own image, instructions how to build one will be released in the Wiki section soon)

### OSes on the roadmap
- [ ] Ubuntu 20.04 LVM Hardened
- [ ] Fedora (latest)
- [ ] CentOS 7
- [ ] FreeBSD 13 UFS
- [ ] FreeBSD 13 ZFS
- [ ] OpenBSD
- [ ] OpenSUSE Leap
- [ ] OpenSUSE Tumbleweed
- [ ] Windows 11
- [ ] Windows Server 2019

### OSes not on the roadmap
- [x] ~~MacOS (any release)~~

# Quickstart Section
## Backups
### Automatic sheduled snapshots
```
@hourly   root  hoster vm snapshot-all --stype hourly  --keep 3
@daily    root  hoster vm snapshot-all --stype daily   --keep 5
@weekly   root  hoster vm snapshot-all --stype weekly  --keep 3
@monthly  root  hoster vm snapshot-all --stype monthly --keep 6
```
