# HosterRed-HyperVisor
![HosterRed Screenshot 1](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/main/screenshots/HosterRed_screenshot_1.png)
#### List of supported OSes
- Debian 11
- Ubuntu 20.04
- Ubuntu 20.04 LVM Hardened
- Fedora (latest)
- CentOS 7
- AlmaLinux 8 (RHEL 8)
- Windows 10 (you'll have to provide your own image)
## Backups
### Automatic sheduled snapshots
```
@hourly  hoster vm snapshot-all --stype hourly  --keep 3
@daily   hoster vm snapshot-all --stype daily   --keep 5
@weekly  hoster vm snapshot-all --stype weekly  --keep 3
@monthly hoster vm snapshot-all --stype monthly --keep 6
```
