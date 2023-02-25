import system
import strutils

const CTL_VM = 2
const VM_SWAPUSAGE = 5

type
  SwapUsage = object
    su_total: cint
    su_avail: cint
    su_used: cint
    su_pagesize: cint

proc getSwapUsage(): SwapUsage =
  var mib: seq[int] = @[CTL_VM, VM_SWAPUSAGE]
  var swapusage: SwapUsage
  var size: int = sizeof(SwapUsage)

  if sysctl(mib.addr, mib.len, &swapusage, &size, nil, 0) != 0:
    raise newException(OSError, "sysctl(VM_SWAPUSAGE) failed")

  return swapusage

proc formatBytes(bytes: cint): string =
  var units: seq[string] = @["B", "KB", "MB", "GB", "TB"]
  var unitIndex: int = 0
  var remaining: float = float(bytes)

  while remaining >= 1024.0 and unitIndex < units.len - 1:
    remaining = remaining / 1024.0
    unitIndex += 1

  return $remaining & " " & units[unitIndex]

proc main() =
  let swapusage = getSwapUsage()
  echo "Total swap space: " & formatBytes(swapusage.su_total)
  echo "Available swap space: " & formatBytes(swapusage.su_avail)
  echo "Used swap space: " & formatBytes(swapusage.su_used)

when isMainModule:
  main()
