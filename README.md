# HosterRed-HyperVisor
![HosterRed Screenshot 1](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/main/screenshots/HosterRed_screenshot_1.png)
#### State icons
🟢 - VM is running
<br>🔴 - VM is stopped
<br>🔒 - VM is located on the encrypted Datased
<br>🔁 - Production VM icon: VM will be included in the autostart, automatic snapshots/replication, etc
#### List of supported OSes
- [x] Debian 11
- [x] Ubuntu 20.04
- [x] Ubuntu 20.04 LVM Hardened
- [x] Fedora (latest)
- [x] CentOS 7
- [x] AlmaLinux 8 (RHEL 8)
- [ ] Windows 10 (you'll have to provide your own image)
- [ ] FreeBSD 13 (image is not ready yet)
## Backups
### Automatic sheduled snapshots
```
@hourly  hoster vm snapshot-all --stype hourly  --keep 3
@daily   hoster vm snapshot-all --stype daily   --keep 5
@weekly  hoster vm snapshot-all --stype weekly  --keep 3
@monthly hoster vm snapshot-all --stype monthly --keep 6
```
