import os
import math

type
  KvmSwap = object
    ksw_devname: string
    ksw_total: uint32
    ksw_used: uint32
    ksw_flags: int32

proc swapmode(var retavail, var retfree: int): int =
  var swapary = newSeq[KvmSwap](1)
  var pagesize = os.getPageSize()
  var swap_maxpages = os.getSysctl("vm.swap_maxpages")

  # CONVERT macro in C:
  # #define CONVERT(v) ((quad_t)(v) * pagesize / 1024)
  template CONVERT(v: uint32): int =
    cast[int](v * uint64(pagesize) div 1024)

  var n = os.kvm_getswapinfo(swapary.ctypes.data, swapary.len, 0)
  if n < 0:
    echo stderr, "Sorry, kvm_getswapinfo returned ", n
    quit(1)

  if swapary[0].ksw_total == 0:
    echo stderr, "Sorry, kvm_getswapinfo said there is 0 swap available"
    quit(1)

  if swapary[0].ksw_total > swap_maxpages:
    swapary[0].ksw_total = swap_maxpages

  retavail = CONVERT(swapary[0].ksw_total)
  retfree = CONVERT(swapary[0].ksw_total - swapary[0].ksw_used)

  return cast[int](swapary[0].ksw_used * 100.0 / swapary[0].ksw_total)

# Example usage
var a, f: int
swapmode(a, f)
echo a, f
