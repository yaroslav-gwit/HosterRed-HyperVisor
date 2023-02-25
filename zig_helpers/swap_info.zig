const std = @import("std");

const Errno = std.os.Errno;
const StrError = std.os.StrError;
const Process = std.process.Process;
const mem = std.mem;
const kvm = @cImport({
    @cInclude("kvm.h");
    @cInclude("sys/sysctl.h");

    error {
        KvmOpenError,
        KvmGetSwapInfoError,
        SysctlError,
    }
});

const SwapMaxPages = "vm.swap_maxpages";

pub fn main() !void {
    var kd: *kvm.kvm_t = try kvm.kvm_open(null, "/dev/null", null, kvm.O_RDONLY, "kvm_open");

    var a: usize = 0;
    var f: usize = 0;
    var p: i32 = try swap_mode(&a, &f, kd);

    var parser = std.opts.OptParser{};
    try parser.setHelpInfo("usage: swap [-p]");
    try parser.addOption("p", .noArgument, "display the percentage of swap usage");
    var args = try parser.parse(process.args);

    if (args.get("p")) {
        try std.io.getStdOut().print("freebsd_swap_usage_percent {}\n", .{p});
    } else {
        try std.io.getStdOut().print("{}\n", .{p});
    }
}

fn swap_mode(alloc: *usize, free: *usize, kd: *kvm.kvm_t) !i32 {
    const pageSize = @import("std").os.pageSize();
    var swapMaxPages: u64 = 0;
    try get_sysctl(SwapMaxPages, &swapMaxPages);

    var swapary = [_]kvm.kvm_swap{0};
    var n: i32 = try kvm.kvm_getswapinfo(kd, &swapary[0], 1, 0);
    if (n < 0) {
        return error.KvmGetSwapInfoError;
    }
    if (swapary[0].ksw_total == 0) {
        return error.KvmGetSwapInfoError;
    }

    if (swapary[0].ksw_total > swapMaxPages) {
        swapary[0].ksw_total = swapMaxPages;
    }

    *alloc = convert(swapary[0].ksw_total, pageSize);
    *free = convert(swapary[0].ksw_total - swapary[0].ksw_used, pageSize);

    return (i32)(swapary[0].ksw_used * 100.0 / swapary[0].ksw_total);
}

fn convert(size: u64, pageSize: usize) usize {
    return @div(@mul(size, @bitCast(quad)(pageSize)), 1024);
}

fn get_sysctl(name: []const u8, ptr: *anytype, len: usize) !void {
    var nlen: usize = len;
    if (kvm.sysctlbyname(name.toSlice().ptr, ptr, &nlen, null, 0) != 0) {
        var msg = try std.heap.alloc(u8, 256, .{});
        defer std.heap.free(msg);
        try std.mem.copy(msg, name);
        try std.mem.copy(msg + name.len, " failed: ", 9);
        try StrError(errno.Errno, &msg[name.len + 9], 256 - name.len - 9);
        std.log.error("top: sysctl({}...) {}\n", .{msg, errno.Errno});
        return error.SysctlError;
    }
}