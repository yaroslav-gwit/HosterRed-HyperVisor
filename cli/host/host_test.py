def byte_converter(bytes:int = 1) -> str:
    # B
    iteration = bytes
    value_type = "B"
    # KB
    if iteration > 1024:
        iteration = iteration / 1024
        value_type = "KB"
    # MB
    if iteration > 1024:
        iteration = iteration / 1024
        value_type = "MB"
    # GB
    if iteration > 1024:
        iteration = iteration / 1024
        value_type = "GB"

    # Will return a value similar to 2GB
    return str(round(float(iteration), 2)) + value_type


def get_ram_info() -> str:
    import subprocess

    def get_page_size() -> str:
        command = "sysctl -nq hw.pagesize"
        shell_command = subprocess.check_output(command, shell=True)
        shell_output = shell_command.decode("utf-8").split()[0]
        return shell_output

    def get_free_mem_pages() -> str:
        command = "sysctl -nq vm.stats.vm.v_free_count"
        shell_command = subprocess.check_output(command, shell=True)
        shell_output = shell_command.decode("utf-8").split()[0]
        return shell_output

    def get_real_memory_pages() -> str:
        command = "sysctl -nq hw.realmem"
        shell_command = subprocess.check_output(command, shell=True)
        shell_output = shell_command.decode("utf-8").split()[0]
        return shell_output

    # Get system memory
    page_size = int(get_page_size())
    free_memory = int(get_free_mem_pages()) * page_size
    real_memory = int(get_real_memory_pages())
    used_memory = real_memory - free_memory

    # Transform values into human readable format
    free_memory_h = byte_converter(bytes=free_memory)
    real_memory_h = byte_converter(bytes=real_memory)
    used_memory_h = byte_converter(bytes=used_memory)

    final_result = used_memory_h + "/" + real_memory_h
    return final_result


def get_number_of_live_vms() -> str:
    from os.path import exists
    import subprocess

    if exists("/dev/vmm/"):
        command = "ls /dev/vmm/"
        shell_command = subprocess.check_output(command, shell=True)
        shell_output = len(shell_command.decode("utf-8").split())
        result = str(shell_output)
    else:
        result = "0"

    return result


def get_arc_size() -> str:
    import subprocess
    
    command = "sysctl -nq kstat.zfs.misc.arcstats.size"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()[0]

    result = byte_converter(bytes = int(shell_output))

    return result


def get_zfs_pool_stats(pool:str = "zroot") -> dict:
    import subprocess

    command = "zpool list -pH " + pool
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()

    pool_full_size = byte_converter(bytes = int(shell_output[1]))
    pool_used = byte_converter(bytes = int(shell_output[2]))
    pool_free = byte_converter(bytes = int(shell_output[3]))

    pool_status = shell_output[9]
    if pool_status == "ONLINE":
        pool_status = "Online"
    else:
        pool_status = "Problem!"

    result = {}
    result["pool_full_size"] = pool_full_size
    result["pool_used"] = pool_used
    result["pool_free"] = pool_free
    result["pool_status"] = pool_status

    return result


def get_hostname() -> str:
    import subprocess
    command = "sysctl -nq kern.hostname"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()[0]

    return shell_output


def human_readable_uptime(seconds_since_boot:float) -> str:
    seconds_mod = seconds_since_boot % 60
    result = str(int(seconds_mod)) + "s"
    
    if seconds_since_boot >= 60:
        minutes = (seconds_since_boot - seconds_mod) / 60
        minutes_mod = minutes % 60
        result = str(int(minutes_mod)) + "m " + result

    if minutes >= 60:
        hours = (minutes - minutes_mod) / 60
        hours_mod = hours % 24
        result = str(int(hours_mod)) + "h " + result

    if hours >= 24:
        days = (hours - hours_mod) / 24
        result = str(int(days)) + "d " + result

    return result


def get_system_uptime() -> str:
    from datetime import datetime
    import subprocess

    command = "sysctl -nq kern.boottime"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()[3].replace(",", "")

    time_now = datetime.now()
    system_boot_time = datetime.fromtimestamp(int(shell_output))
    seconds_since_boot = (time_now - system_boot_time).total_seconds()

    return human_readable_uptime(seconds_since_boot)


def get_swap_info() -> str:
    import subprocess
    command = "swapinfo"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()

    swap_used = shell_output[7]
    swap_overall = shell_output[8]
    if int(swap_used) > 0:
        swap_used = int(swap_used) * 1024
        swap_used = byte_converter(bytes = swap_used)
    else:
        swap_used = "0B"

    if int(swap_overall) > 0:
        swap_overall = int(swap_overall) * 1024
        swap_overall = byte_converter(bytes = swap_overall)
    else:
        swap_overall = "0B"
        
    return swap_used + "/" + swap_overall


def table_output(table_title:bool = False) -> None:
    from rich.console import Console
    from rich.table import Table
    from rich import box

    if not show_title:
        table = Table(box=box.ROUNDED)
    else:
        table = Table(title = "Host Information", box=box.ROUNDED)

    table.add_column("Host", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Live VMs", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Uptime", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("RAM (Used/Overall)", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Swap (Used/Overall)", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("ZFS Arc Size", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Zroot Free", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Zroot Status", justify="center", style="bright_cyan", no_wrap=True)

    zfs_pool_stats = get_zfs_pool_stats()
    zroot_free = zfs_pool_stats["pool_used"] + "/" + zfs_pool_stats["pool_full_size"]
    zroot_status = zfs_pool_stats["pool_status"]
    table.add_row(get_hostname(), get_number_of_live_vms() , get_system_uptime(), get_ram_info(), get_swap_info(), get_arc_size(), zroot_free, zroot_status)

    console = Console()
    console.print(table)


if __name__ == "__main__":
    table_output()
