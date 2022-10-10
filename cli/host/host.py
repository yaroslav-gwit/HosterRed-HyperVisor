import re
import typer


def json_output():
    import json

    json_d = {}
    zfs_pool_stats = get_zfs_pool_stats()
    ram_info = get_ram_info(table_output = False)
    swap_info = get_swap_info(table_output = False)
    zfs_arc_size = get_arc_size(table_output = False)
    
    json_d["hostname"] = get_hostname()
    json_d["uptime"] = get_system_uptime()
    json_d["number_of_running_vms"] = get_number_of_live_vms()
    json_d["zfs_arc_size"] = zfs_arc_size["zfs_arc_size"]
    json_d["zfs_arc_size_h"] = zfs_arc_size["zfs_arc_size_h"]
    json_d["zroot_space_overall"] = zfs_pool_stats["pool_full_size"]
    json_d["zroot_space_overall_h"] = zfs_pool_stats["pool_full_size_h"]
    json_d["zroot_space_used"] = zfs_pool_stats["pool_used"]
    json_d["zroot_space_used_h"] = zfs_pool_stats["pool_used_h"]
    json_d["zroot_space_free"] = zfs_pool_stats["pool_free"]
    json_d["zroot_space_free_h"] = zfs_pool_stats["pool_free_h"]
    json_d["zroot_status"] = zfs_pool_stats["pool_status"]
    json_d["real_memory"] = ram_info["real_memory"]
    json_d["real_memory_h"] = ram_info["real_memory_h"]
    json_d["used_memory"] = ram_info["used_memory"]
    json_d["used_memory_h"] = ram_info["used_memory_h"]
    json_d["free_memory"] = ram_info["free_memory"]
    json_d["free_memory_h"] = ram_info["free_memory_h"]
    json_d["swap_used"] = swap_info["swap_used"]
    json_d["swap_used_h"] = swap_info["swap_used_h"]
    json_d["swap_overall"] = swap_info["swap_overall"]
    json_d["swap_overall_h"] = swap_info["swap_overall_h"]

    return json.dumps(json_d, indent=3)


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


# Returns Dict or Str depending if "table_output" is True
def get_ram_info(table_output:bool = True):
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

    memory_dict = {}
    memory_dict["free_memory"] = free_memory
    memory_dict["free_memory_h"] = free_memory_h
    memory_dict["real_memory"] = real_memory
    memory_dict["real_memory_h"] = real_memory_h
    memory_dict["used_memory"] = used_memory
    memory_dict["used_memory_h"] = used_memory_h

    if table_output:
        return used_memory_h + "/" + real_memory_h
    else:
        return memory_dict


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


# Returns Dict or Str depending if "table_output" is True
def get_arc_size(table_output:bool = True):
    import subprocess
    
    command = "sysctl -nq kstat.zfs.misc.arcstats.size"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()[0]

    zfs_arc_size = int(shell_output)
    zfs_arc_size_h = byte_converter(bytes = int(shell_output))

    if table_output:
        return zfs_arc_size_h
    else:
        arc_size_d = {}
        arc_size_d["zfs_arc_size"] = zfs_arc_size
        arc_size_d["zfs_arc_size_h"] = zfs_arc_size_h
        return arc_size_d


def get_zfs_pool_stats(pool:str = "zroot") -> dict:
    import subprocess

    command = "zpool list -pH " + pool
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()

    pool_full_size = int(shell_output[1])
    pool_full_size_h = byte_converter(bytes = int(shell_output[1]))
    pool_used = int(shell_output[2])
    pool_used_h = byte_converter(bytes = int(shell_output[2]))
    pool_free = int(shell_output[3])
    pool_free_h = byte_converter(bytes = int(shell_output[3]))

    pool_status = shell_output[9]
    if pool_status == "ONLINE":
        pool_status = "Online"
    else:
        pool_status = "Problem!"

    result = {}
    result["pool_full_size"] = pool_full_size
    result["pool_full_size_h"] = pool_full_size_h
    result["pool_used"] = pool_used
    result["pool_used_h"] = pool_used_h
    result["pool_free"] = pool_free
    result["pool_free_h"] = pool_free_h
    result["pool_status"] = pool_status

    return result


def get_hostname() -> str:
    import subprocess
    command = "sysctl -nq kern.hostname"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()[0]

    return shell_output


def human_readable_uptime(seconds_since_boot:int) -> str:
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


# Returns Dict or Str depending if "table_output" is True
def get_swap_info(table_output:bool = True):
    import subprocess
    command = "swapinfo"
    shell_command = subprocess.check_output(command, shell=True)
    shell_output = shell_command.decode("utf-8").split()

    swap_used = shell_output[7]
    swap_overall = shell_output[8]
    if int(swap_used) > 0:
        swap_used = int(swap_used) * 1024
        swap_used_h = byte_converter(bytes = swap_used)
    else:
        swap_used = 0
        swap_used_h = "0B"

    if int(swap_overall) > 0:
        swap_overall = int(swap_overall) * 1024
        swap_overall_h = byte_converter(bytes = swap_overall)
    else:
        swap_overall = 0
        swap_overall_h = "0B"

    if table_output:
        return swap_used_h + "/" + swap_overall_h
    else:
        swap_dict = {}
        swap_dict["swap_used"] = swap_used
        swap_dict["swap_used_h"] = swap_used_h
        swap_dict["swap_overall"] = swap_overall
        swap_dict["swap_overall_h"] = swap_overall_h
        return swap_dict


def table_output(table_title:bool = False) -> None:
    from rich.console import Console
    from rich.table import Table
    from rich import box

    if not table_title:
        table = Table(box=box.ROUNDED)
    else:
        table = Table(title = " Host Information", box=box.ROUNDED, title_justify = "left")

    table.add_column("Host", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Live VMs", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Uptime", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("RAM (Used/Total)", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Swap (Used/Total)", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("Zroot (Used/Total)", justify="center", style="bright_cyan", no_wrap=True)
    table.add_column("ZFS Arc Size", justify="center", style="bright_cyan", no_wrap=True)
    zfs_pool_stats = get_zfs_pool_stats()
    zroot_status = zfs_pool_stats["pool_status"]
    if zroot_status != "Online":
        table.add_column("Zroot Status", justify="center", style="red", no_wrap=True)
    else:
        table.add_column("Zroot Status", justify="center", style="bright_cyan", no_wrap=True)

    zroot_free = zfs_pool_stats["pool_used_h"] + "/" + zfs_pool_stats["pool_full_size_h"]
    table.add_row(get_hostname(), get_number_of_live_vms() , get_system_uptime(), get_ram_info(), get_swap_info(), zroot_free, get_arc_size(), zroot_status)

    console = Console()
    console.print(table)


""" Section below is responsible for the CLI input/output """
app = typer.Typer(context_settings=dict(max_content_width=800))


@app.command()
def info(
        json:bool = typer.Option(False, help="Output json instead of a table"),
        table_title:bool = typer.Option(False, help="Show table title (useful when showing multiple tables)")
    ):

    """ Print out the host related info """

    if json:
        print(json_output())
    else:
        table_output(table_title = table_title)


@app.command()
def init():
    """ Initialise Kernel modules and required services """

    #_ LIST OF MODULES TO LOAD _#
    """
    kldstat -m $MODULE
    kldstat -mq $MODULE

    kldload vmm
    kldload nmdm
    kldload if_bridge
    kldload if_tuntap
    kldload if_tap

    sysctl net.link.tap.up_on_open=1

    13.0-RELEASE-p11
    """

    import invoke

    result = invoke.run("kldstat -v", hide=True)
    kld_stat_lines = result.stdout.splitlines()

    module_list = ["vmm", "nmdm", "if_tap", "if_bridge", "if_tuntap"]
    modules_loaded = []

    for module in module_list:
        module_str = ".*" + module + ".*"
        module_str_ko = ".*" + module + ".ko.*"
        re_match_1 = re.compile(module_str_ko)
        re_match_2 = re.compile(module_str)

        for line in kld_stat_lines:
            if re_match_1.match(line) or re_match_2.match(line):
                modules_loaded.append(module)
                continue

    modules_loaded = set(modules_loaded); modules_loaded = list(modules_loaded)
    for module in modules_loaded:
        if module in module_list:
            module_list.remove(module)

    if len(module_list) > 0:
        for module in module_list:
            kldload = "kldload " + module
            result = invoke.run(kldload, hide=True)
            print("Module loaded: " + module)

    result = invoke.run("sysctl net.link.tap.up_on_open=1", hide=True)


""" If this file is executed from the command line, activate Typer """
if __name__ == "__main__":
    app()
