/* copied from /usr/src/usr.bin/top/machine.c */

#include <sys/errno.h>
#include <sys/file.h>
#include <sys/sysctl.h>
#include <fcntl.h>
#include <kvm.h>
#include <err.h>
#include <paths.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#define GETSYSCTL(name, var) getsysctl(name, &(var), sizeof(var))

static void
getsysctl(const char *name, void *ptr, size_t len)
{
		size_t nlen = len;

		if (sysctlbyname(name, ptr, &nlen, NULL, 0) == -1) {
				fprintf(stderr, "top: sysctl(%s...) failed: %s\n", name,
					strerror(errno));
				exit(1);
		}
		if (nlen != len) {
				fprintf(stderr, "top: sysctl(%s...) expected %lu, got %lu\n",
					name, (unsigned long)len, (unsigned long)nlen);
				exit(1);
		}
}

#define GETSYSCTL(name, var) getsysctl(name, &(var), sizeof(var))

static kvm_t *kd;

static int
swapmode(int *retavail, int *retfree)
{
		int n;
		struct kvm_swap swapary[1];
	/* The getpagesize() function returns the number of bytes in a page. */
		int pagesize = getpagesize();
		u_long swap_maxpages;

	GETSYSCTL("vm.swap_maxpages", swap_maxpages);

#define CONVERT(v)	  ((quad_t)(v) * pagesize / 1024)

	/*
	kvm_swap structure:
		char ksw_devname[];
		u_int ksw_total;
		u_int ksw_used;
		int ksw_flags;
	Values are in PAGE_SIZE'd chunks (see getpagesize(3)).  ksw_flags
	contains a copy of the swap device flags.
	*/
		n = kvm_getswapinfo(kd, swapary, 1, 0);
		if (n < 0) {
		fprintf(stderr, "Sorry, kvm_getswapinfo returned %d\n",n);
		exit(1);
	}
		if (swapary[0].ksw_total == 0) {
		fprintf(stderr, "Sorry, kvm_getswapinfo said there is 0 swap available\n");
		exit(1);
	}

		/* ksw_total contains the total size of swap all devices which may
		   exceed the maximum swap size allocatable in the system */
		if ( swapary[0].ksw_total > swap_maxpages )
				swapary[0].ksw_total = swap_maxpages;

		*retavail = CONVERT(swapary[0].ksw_total);
		*retfree = CONVERT(swapary[0].ksw_total - swapary[0].ksw_used);

		n = (int)(swapary[0].ksw_used * 100.0 / swapary[0].ksw_total);
		return (n);
}

void usage() {
	(void)fprintf(stderr,
		"usage: swap [-p]\n");
	exit(1);
}

int main(int argc,char **argv) {
	int a,f;
		kd = kvm_open(NULL, _PATH_DEVNULL, NULL, O_RDONLY, "kvm_open");
	if (kd == NULL)
		return (-1);
	int p = swapmode(&a,&f);	

	int ch,pflag=0;
	while ((ch = getopt(argc, argv, "p")) != -1) {
		switch (ch) {
			case 'p':
				pflag = 1;
				break;
			case '?':
			default:
				usage();
		}
	 	}

	if ( pflag )
		printf("freebsd_swap_usage_percent %d\n",p);
	else
		printf("%d\n",p);
}
