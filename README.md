# HosterRed-HyperVisor
![HosterRed Screenshot 1](https://github.com/yaroslav-gwit/HosterRed-HyperVisor/blob/main/screenshots/HosterRed_screenshot_1.png)
## Backups
### Automatic sheduled snapshots
```
@hourly  hoster vm snapshot-all --stype hourly  --keep 3
@daily   hoster vm snapshot-all --stype daily   --keep 5
@weekly  hoster vm snapshot-all --stype weekly  --keep 3
@monthly hoster vm snapshot-all --stype monthly --keep 6
```
