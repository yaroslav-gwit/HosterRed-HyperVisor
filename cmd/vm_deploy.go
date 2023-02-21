package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"
	"path"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"
)

var (
	vmDeployCmd = &cobra.Command{
		Use:   "deploy",
		Short: "Deploy the VM, using a pre-defined template",
		Long:  `Deploy the VM, using a pre-defined template`,
		// Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(args[0])
			ip, err := generateNewIp()
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(ip)
		},
	}
)

const ciUserDataTemplate = `
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

const ciMetaDataTemplate = `
instance-id: iid-wmxgv
local-hostname: test-vm-1
`

const ciNetworkConfigTemplate = `
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

const vmConfigFileTemlate = `
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

func generateNewIp() (string, error) {
	var existingIps []string
	for _, v := range getAllVms() {
		tempConfig := vmConfig(v)
		existingIps = append(existingIps, tempConfig.Networks[0].IPAddress)
	}

	networks, err := networkInfo()
	if err != nil {
		return "", errors.New(err.Error())
	}

	subnet := networks[0].Subnet
	rangeStart := networks[0].RangeStart
	rangeEnd := networks[0].RangeEnd

	var randomIp string
	// var err error
	randomIp, err = generateUniqueRandomIp(subnet)
	if err != nil {
		return "", errors.New("could not generate a random IP address: " + err.Error())
	}

	iteration := 0
	for {
		if slices.Contains(existingIps, randomIp) || !ipIsWinthinRange(randomIp, subnet, rangeStart, rangeEnd) {
			randomIp, err = generateUniqueRandomIp(subnet)
			if err != nil {
				return "", errors.New("could not generate a random IP address: " + err.Error())
			}
			iteration = iteration + 1
			if iteration > 400 {
				return "", errors.New("ran out of IP available addresses within this range")
			}
		} else {
			break
		}
	}

	return randomIp, nil
}

func generateUniqueRandomIp(subnet string) (string, error) {
	// Set the seed for the random number generator
	rand.Seed(time.Now().UnixNano())

	// Parse the subnet IP and mask
	ip, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		return "", errors.New(err.Error())
	}

	// Calculate the size of the address space within the subnet
	size, _ := ipNet.Mask.Size()
	numHosts := (1 << (32 - size)) - 2

	// Generate a random host address within the subnet
	host := rand.Intn(numHosts) + 1
	addr := ip.Mask(ipNet.Mask)
	addr[0] |= byte(host >> 24)
	addr[1] |= byte(host >> 16)
	addr[2] |= byte(host >> 8)
	addr[3] |= byte(host)

	stringAddress := fmt.Sprintf("%v", addr)
	return stringAddress, nil
}

func ipIsWinthinRange(ipAddress string, subnet string, rangeStart string, rangeEnd string) bool {
	// Parse the subnet IP and mask
	_, ipNet, err := net.ParseCIDR(subnet)
	if err != nil {
		panic(err)
	}

	// Define the range of allowed host addresses
	start := net.ParseIP(rangeStart).To4()
	end := net.ParseIP(rangeEnd).To4()

	// Parse the IP address to check
	ip := net.ParseIP(ipAddress).To4()

	// Check if the IP address is within the allowed range
	if ipNet.Contains(ip) && bytesInRange(ip, start, end) {
		return true
	} else {
		return false
	}
}

func bytesInRange(ip, start, end []byte) bool {
	for i := 0; i < len(ip); i++ {
		if start[i] > end[i] {
			log.Fatal("Make sure range start is lower than range end!")
		} else if ip[i] < start[i] || ip[i] > end[i] {
			return false
		}
	}
	return true
}

type NetworkInfoSt struct {
	Name            string `json:"network_name"`
	Gateway         string `json:"network_gateway"`
	Subnet          string `json:"network_subnet"`
	RangeStart      string `json:"network_range_start"`
	RangeEnd        string `json:"network_range_end"`
	BridgeInterface string `json:"bridge_interface"`
	ApplyBridgeAddr bool   `json:"apply_bridge_address"`
	Comment         string `json:"comment"`
}

func networkInfo() ([]NetworkInfoSt, error) {
	// JSON config file location
	execPath, err := os.Executable()
	if err != nil {
		return []NetworkInfoSt{}, err
	}
	networkConfigFile := path.Dir(execPath) + "/config_files/network_config.json"

	// Read the JSON file
	data, err := ioutil.ReadFile(networkConfigFile)
	if err != nil {
		return []NetworkInfoSt{}, err
	}

	// Unmarshal the JSON data into a slice of Network structs
	var networks = []NetworkInfoSt{}
	err = json.Unmarshal(data, &networks)
	if err != nil {
		return []NetworkInfoSt{}, err
	}

	return networks, nil
}
