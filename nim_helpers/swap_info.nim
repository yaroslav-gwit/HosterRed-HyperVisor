import os
import strutils
import times

type
  KvmSwap = object
    ksw_devname: string
    ksw_total: uint
    ksw_used: uint
    ksw_flags: int

proc getSwapMode(): (int, int) =
  var kd = kvm_open(nil, "/dev/null", nil, O_RDONLY, "kvm_open")
  if kd == nil:
    raiseException(OSError, "kvm_open failed")
  defer:
    if kd != nil:
      kvm_close(kd)
  
  let pagesize = getpagesize().int32
  var swap_maxpages: uint64 = 0
  sysctlbyname("vm.swap_maxpages", swap_maxpages.addr, swap_maxpages.len.addr, nil, 0)
  
  var swapary = kvm_swap(ksw_devname: init, ksw_total: 0, ksw_used: 0, ksw_flags: 0)
  let n = kvm_getswapinfo(kd, &swapary, 1, 0)
  if n < 0:
    raiseException(OSError, "kvm_getswapinfo returned $(n)")
  elif swapary.ksw_total == 0:
    raiseException(OSError, "kvm_getswapinfo said there is 0 swap available")
  
  # ksw_total contains the total size of swap all devices which may
  # exceed the maximum swap size allocatable in the system
  if swapary.ksw_total > swap_maxpages:
    swapary.ksw_total = swap_maxpages
  
  let totalSwapKB = swapary.ksw_total * pagesize / 1024
  let freeSwapKB = (swapary.ksw_total - swapary.ksw_used) * pagesize / 1024
  
  return (totalSwapKB.int, freeSwapKB.int)

when isMainModule:
  let (totalSwap, freeSwap) = getSwapMode()
  echo "Total Swap: ", totalSwap, " KB"
  echo "Free Swap: ", freeSwap, " KB"
