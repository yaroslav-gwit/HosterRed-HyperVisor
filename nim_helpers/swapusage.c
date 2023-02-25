#include <sys/types.h>
#include <sys/sysctl.h>
#include <stdio.h>

int main() {
    int mib[2];
    size_t size;
    struct xsw_usage swap_usage;

    mib[0] = CTL_VM;
    mib[1] = VM_SWAPUSAGE;
    size = sizeof(swap_usage);

    if (sysctl(mib, 2, &swap_usage, &size, NULL, 0) == 0) {
        printf("Total swap space: %lu\n", swap_usage.xsu_total);
        printf("Used swap space: %lu\n", swap_usage.xsu_used);
        printf("Free swap space: %lu\n", swap_usage.xsu_avail);
        return 0;
    } else {
        perror("sysctl");
        return 1;
    }
}
