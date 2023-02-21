package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	vmDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the VM, using a pre-defined template",
		Long:  `Deploy the VM, using a pre-defined template`,
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args[0])
		},
	}
)

const ciUserData = `
#cloud-config

users:
  - default
  - name: root
    lock_passwd: false
    ssh_pwauth: true
    disable_root: false
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDs7hczETEkQ7k1f4xxQCHHWjqOaiVVKpJegMXqiOkHmmJyarnrxGb2YOKx9Vn4jHEJyzO5vcUCgSDhbDQ3AWoMyUnKbEn/beOy31Fft0Pt54McIb0G6M2gM7Ywgwek6JL2ltJMj6Q1PvZkBoBGNVc+0q7AYq1J80s9baO7l9pAJ73BJm18lqwir0kaFHHxB7IdBVoKTaNFSEu8Lbt8axwOjiPiNKv5jFKdAXkU7IEO5Ts+UOEMQf8tCFkMmWH5h71WtcMy9BglqtvSjxxn1bWcU9MEvunOaXyNTVy+FUvpaVvCcKm5EsLNMXtVAQK0K5lfzHgcXiHw4f2bgUr2oubm5KuLyMmneq/5NPf8B4yR6rXD6D+d7ZzUVwW8LhKyd/MfCNjudwShrV8kkp/cc0JoWhelDCxp+YOqPKeIWZBYHZkDP5cQCM6TjYyZ0JfTlZaATk6PV7LM3xHSlBnbXKYDwp3UlvVDARFiCQMKIQDqKHC37SzL0vX4BEvhf7m1oXhv+P7dbBIGrZThDD4sjaHgegTfouOcG+ggQSto1Y9uApXepeU/5I0+TtPuoKr2u9xzX8VYnlNceOrx2+52sYa1AlFG/OhL2tEMV91QpZox5T35mDv1nKhflcLc4YLIMvO/f2w3FOfnrjbcF2U3y4bYr8ul9OJZzX++uC7Q8cZNvw== root@hoster-test-0101
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC4ecr6eYKbz7zUnsUsZ2Lw+mkSEYiIIyjhsrHBPoCoNN0V6O8CAog6Xc52KM4r+ozC07J4iz0G2UKjO7INSoDGVa6gMwzTrPeg2hLY7aNJytRqSVwpo5nujKcNCISoY9wCcXkTG1yVvty7pJ2Dk8ysHl6nCn6pIfQ87hRJ+ywSiE4y5AqDC5xMeGIw/bPajzofOXCQSMNgtflynbeBC6McA0kClmUp0kCX6kwGdPU2FNmz3Vba6ixMhP7Ng0vTJSmOStEqpd2in/nhz2JNtUcWIickPZf1II0s4LXZ74H308QJtYw1nVavKF4SKiVPkP7gKGrnCYDsq2p0P+EhZlHvK/nJPdS8qaWyVmeFr07o+F5mhYbw4t2BJUKgfKnStqatXZtAzwKvPyYzVpjD0nGVhxjF2BH0XdtWfbdTARaeZxJEVPoPtLm4cTP0CA7RljplJuGwp9DGgzwO3zbfx8CcjtmkJaELG9hrTeMhRmY6FMccci4LeJLhs9LHHblNXYEu+5xGzPPEj2AP3x73XmVdcRpUExtwScN4PrcV0JPSqkTwfwB25mCKfEv76vTl9iPe9D8Veerfyuhq2LsboGExTqw4HtazO2tmwfVIptF2Ak6eeEKX09V1lPQbIWi/vRIYVLgTHMeaU66rvT6LWqxyWUB++11ToNBNrigTa2tOWw== root@hoster-test-0102
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDsOnhH5R/Incbxjtb/c4Z3t3hlDo3aHt02miE5R1i09s78zljYFHCAmJ6/fKDTPmt+irNPDTaCTShZTen/U1B5E/aljFzl4ZJxL9Z4MV0YXfbq7IMGs2ln/DWV+pVdgmuhCP4cZsFOeuX7T6JPk4IgAEGQrGl+J/+xzlU9PeU0DMgUFl1sljOEtvKnnjH/cs4s/gvQGOBFiggB4Gl2UPBYaT88dRrv5VJJ5G/WKq+Ngk3qwA0ilx/9L3Z9Cio1ROHrHw32Hxa6BgBz4fwZWRM7oeqtc2jIYP0upbAsthcU6nG8WAq1i/fKmJe7p89b6afIoqPA8ZxmHjVOQJ92K9a+1wrR7z2elZ/F+aosOQ0kda66cmYvSWyc2CylySobxJsHNrx7Jn+kEHzRd7zFOgOitS2QbwvleHpfd+EeDIHc09xmcJfS6AtPg3cuku5EMS+EHOD9TgshqgzNxVRARMAo8LPepsoCnUw7+l2rt/ff07Zs3Ka8jJ+3DRTkqARHQS8= yaroslav@lenovo-yoga-052155

  - name: gwitsuper
    sudo: ALL=(ALL) NOPASSWD:ALL
    groups: wheel
    ssh_pwauth: true
    lock_passwd: false
    ssh_authorized_keys:
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDs7hczETEkQ7k1f4xxQCHHWjqOaiVVKpJegMXqiOkHmmJyarnrxGb2YOKx9Vn4jHEJyzO5vcUCgSDhbDQ3AWoMyUnKbEn/beOy31Fft0Pt54McIb0G6M2gM7Ywgwek6JL2ltJMj6Q1PvZkBoBGNVc+0q7AYq1J80s9baO7l9pAJ73BJm18lqwir0kaFHHxB7IdBVoKTaNFSEu8Lbt8axwOjiPiNKv5jFKdAXkU7IEO5Ts+UOEMQf8tCFkMmWH5h71WtcMy9BglqtvSjxxn1bWcU9MEvunOaXyNTVy+FUvpaVvCcKm5EsLNMXtVAQK0K5lfzHgcXiHw4f2bgUr2oubm5KuLyMmneq/5NPf8B4yR6rXD6D+d7ZzUVwW8LhKyd/MfCNjudwShrV8kkp/cc0JoWhelDCxp+YOqPKeIWZBYHZkDP5cQCM6TjYyZ0JfTlZaATk6PV7LM3xHSlBnbXKYDwp3UlvVDARFiCQMKIQDqKHC37SzL0vX4BEvhf7m1oXhv+P7dbBIGrZThDD4sjaHgegTfouOcG+ggQSto1Y9uApXepeU/5I0+TtPuoKr2u9xzX8VYnlNceOrx2+52sYa1AlFG/OhL2tEMV91QpZox5T35mDv1nKhflcLc4YLIMvO/f2w3FOfnrjbcF2U3y4bYr8ul9OJZzX++uC7Q8cZNvw== root@hoster-test-0101
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC4ecr6eYKbz7zUnsUsZ2Lw+mkSEYiIIyjhsrHBPoCoNN0V6O8CAog6Xc52KM4r+ozC07J4iz0G2UKjO7INSoDGVa6gMwzTrPeg2hLY7aNJytRqSVwpo5nujKcNCISoY9wCcXkTG1yVvty7pJ2Dk8ysHl6nCn6pIfQ87hRJ+ywSiE4y5AqDC5xMeGIw/bPajzofOXCQSMNgtflynbeBC6McA0kClmUp0kCX6kwGdPU2FNmz3Vba6ixMhP7Ng0vTJSmOStEqpd2in/nhz2JNtUcWIickPZf1II0s4LXZ74H308QJtYw1nVavKF4SKiVPkP7gKGrnCYDsq2p0P+EhZlHvK/nJPdS8qaWyVmeFr07o+F5mhYbw4t2BJUKgfKnStqatXZtAzwKvPyYzVpjD0nGVhxjF2BH0XdtWfbdTARaeZxJEVPoPtLm4cTP0CA7RljplJuGwp9DGgzwO3zbfx8CcjtmkJaELG9hrTeMhRmY6FMccci4LeJLhs9LHHblNXYEu+5xGzPPEj2AP3x73XmVdcRpUExtwScN4PrcV0JPSqkTwfwB25mCKfEv76vTl9iPe9D8Veerfyuhq2LsboGExTqw4HtazO2tmwfVIptF2Ak6eeEKX09V1lPQbIWi/vRIYVLgTHMeaU66rvT6LWqxyWUB++11ToNBNrigTa2tOWw== root@hoster-test-0102
      - ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDsOnhH5R/Incbxjtb/c4Z3t3hlDo3aHt02miE5R1i09s78zljYFHCAmJ6/fKDTPmt+irNPDTaCTShZTen/U1B5E/aljFzl4ZJxL9Z4MV0YXfbq7IMGs2ln/DWV+pVdgmuhCP4cZsFOeuX7T6JPk4IgAEGQrGl+J/+xzlU9PeU0DMgUFl1sljOEtvKnnjH/cs4s/gvQGOBFiggB4Gl2UPBYaT88dRrv5VJJ5G/WKq+Ngk3qwA0ilx/9L3Z9Cio1ROHrHw32Hxa6BgBz4fwZWRM7oeqtc2jIYP0upbAsthcU6nG8WAq1i/fKmJe7p89b6afIoqPA8ZxmHjVOQJ92K9a+1wrR7z2elZ/F+aosOQ0kda66cmYvSWyc2CylySobxJsHNrx7Jn+kEHzRd7zFOgOitS2QbwvleHpfd+EeDIHc09xmcJfS6AtPg3cuku5EMS+EHOD9TgshqgzNxVRARMAo8LPepsoCnUw7+l2rt/ff07Zs3Ka8jJ+3DRTkqARHQS8= yaroslav@lenovo-yoga-052155

chpasswd:
  list: |
    root:Fx1Y6UFFecXHFBeViVP8s16rpxbRvnB5G7DQ3K9h3
    gwitsuper:hav5XEQRrIlH8VaIMtLSTsVvYytzDc21IjHxnBkXT
  expire: False

package_update: false
package_upgrade: false
`

const ciMetaData = `
instance-id: iid-wmxgv
local-hostname: test-vm-1
`

const ciNetworkConfig = `
version: 2
ethernets:
  interface0:
     match:
       macaddress: "{{ .MacAddress }}"
     
     set-name: eth0
     addresses:
     - {{ .IpAddress }}/{{ .NetworkSubnet }}
     
     gateway4: {{ .NetworkGateway }}
     
     nameservers:
       search: [gateway-it.internal, ]
       addresses: [{{ .NetworkGateway }}, ]
`

const vmJsonConfig = `
{
    "cpu_sockets": "1",
    "cpu_cores": "1",
    "memory": "1G",
    "loader": "uefi",
    "live_status": "production",
    "os_type": "debian11",
    "os_comment": "Debian 11",
    "owner": "System",
    "parent_host": "hoster-test-0101",

    "networks": [
        {
            "network_adaptor_type": "virtio-net",
            "network_bridge": "internal",
            "network_mac": "58:9c:fc:75:22:73",
            "ip_address": "10.0.100.10",
            "comment": "Internal Network"
        }
    ],

    "disks": [
        {
            "disk_type": "virtio-blk",
            "disk_location": "internal",
            "disk_image": "disk0.img",
            "comment": "OS Drive"
        },
        {
            "disk_type": "ahci-cd",
            "disk_location": "internal",
            "disk_image": "seed.iso",
            "comment": "Cloud Init ISO"
        }
    ],

    "include_hostwide_ssh_keys": true,
    "vm_ssh_keys": [
        {
            "key_owner": "Yaroslav",
            "key_value": "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDvZbo8t8qSdcHZjCiRi0fVoGBHDXJVPYR+OynuxFc+Sle45xMfRiKjRwHnzBnUqTTppsE2wnLTxVMLfrV4Fqnx8irnr+Rll+YvfMPtCpwcR/Zxlu/zHXK+YUbKIOx0/qKJqfsJ8PaZffob7rGoIBgko8fO7iPd0Y5MK1T+uoVQnFRJMAr6RLlz1oMEppdWsEA2pTUhM6mUj57yqwmIznHIYjy44qIWauqOyR7NB9NV7ahxYh/K6lQtcHqxb9l3AE/dfV/RPAv5CDIEGhs1oCHN/1o1iKoKZKOGZJbn02tGNqA+8XUcKiT1Wh82fGhU6GKj/CWqhs0RNpjp12ETGabPzWEZ12OP6GkoIFmkQEFgdpq3fUlCcb1uRAywOncoCLN8njcMGyUMBR9lB+yWDrSY7psTjHGPGb7+nEl+4NvZB1Dhqci6lGYgO0/DS71UVPs8LTEcrqa18ir0yjcIvhEgc/IbPbnySMSm33PpIKHvartqmuIdo8d3kwFbuMDRWBk= yaroslav@ryzen-pc",
            "comment": "Fedora Ryzen PC Key"
        }
    ],

    "vnc_port": "5909",
    "vnc_password": "f7PJ2KcY",

    "description": "-"
}
`
