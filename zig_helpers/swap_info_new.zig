const std = @import("std");
const os = std.os;

const GETSYSCTL = fn(name: []const u8, var: anytype) !void {
    var nlen: usize = @sizeOf(var);

    var ret = os.sysctl.byName(name, var, &nlen, null, 0);
    if (ret != 0) {
        std.debug.print("top: sysctl({}) failed: {}\n", .{name, os.errorName(ret)});
        os.exit(1);
    }

    if (nlen != @sizeOf(var)) {
        std.debug.print("top: sysctl({}) expected {}, got {}\n", .{name, @sizeOf(var), nlen});
        os.exit(1);
    }
}

const CONVERT = fn(v: u64, pagesize: u64) u64 {
    return @div(@mul(@bitCast(u128, v), pagesize), 1024);
};

const swapmode = fn(kd: *os.Kvm, outRetavail: *u32, outRetfree: *u32) !u32 {
    const pagesize = os.getPageSizes().standard;

    var swap_maxpages: u64 = 0;
    try GETSYSCTL("vm.swap_maxpages", swap_maxpages);

    var swapary = [_]os.KvmSwap{0};
    const n = try kd.getSwapInfo(&swapary[0], swapary.len, 0);
    if (n < 0) {
        std.debug.print("Sorry, kvm_getswapinfo returned {}\n", .{n});
        os.exit(1);
    }

    if (swapary[0].total == 0) {
        std.debug.print("Sorry, kvm_getswapinfo said there is 0 swap available\n", .{});
        os.exit(1);
    }

    if (swapary[0].total > swap_maxpages) {
        swapary[0].total = swap_maxpages;
    }

    *outRetavail = CONVERT(swapary[0].total, pagesize);
    *outRetfree = CONVERT(swapary[0].total - swapary[0].used, pagesize);

    return u32(swapary[0].used * 100.0 / swapary[0].total);
}

pub fn main() !void {
    var args = try os.Args.parseOptions({
        .p = std.Options.ArgumentKind.Present,
    });

    var kd = try os.Kvm.open(os.Kvm.OpenFlag.None, os.OpenFlag.O_RDONLY, "kvm_open");
    defer kd.close();

    var a: u32 = 0;
    var f: u32 = 0;
    var p = try swapmode(&kd, &a, &f);

    switch (args.get("p", std.Options.ArgumentKind.Present)) {
        .Present => {
            std.debug.print("freebsd_swap_usage_percent {}\n", .{p});
        },
        _ => {
            std.debug.print("{}\n", .{p});
        },
    }
}
