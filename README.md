# HosterRed-HyperVisor
## Backups
### Automatic sheduled snapshots
```
@hourly hoster vm snapshot-all --stype hourly --keep 3
@daily hoster vm snapshot-all --stype daily --keep 5
@weekly hoster vm snapshot-all --stype weekly --keep 3
@monthly hoster vm snapshot-all --stype monthly --keep 6
```
