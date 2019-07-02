test_container_devices_nic_macvlan() {
  ensure_import_testimage
  ensure_has_localhost_remote "${LXD_ADDR}"

  ct_name="nt$$"
  ipRand=$(shuf -i 0-9 -n 1)

  # Create dummy interface for use as parent.
  ip link add "${ct_name}" type dummy
  ip link set "${ct_name}" up

  # Test pre-launch profile config is applied at launch.
  lxc profile copy default "${ct_name}"
  lxc profile device set "${ct_name}" eth0 parent "${ct_name}"
  lxc profile device set "${ct_name}" eth0 nictype "macvlan"
  lxc profile device set "${ct_name}" eth0 mtu "1400"

  lxc launch testimage "${ct_name}" -p "${ct_name}"
  lxc exec "${ct_name}" -- ip addr add "192.0.2.1${ipRand}/24" dev eth0
  lxc exec "${ct_name}" -- ip addr add "2001:db8::1${ipRand}/64" dev eth0

  # Check custom MTU is applied if feature available in LXD.
  if lxc info | grep 'network_phys_macvlan_mtu: "true"' ; then
    if ! lxc exec "${ct_name}" -- ip link show eth0 | grep "mtu 1400" ; then
      echo "mtu invalid"
      false
    fi
  fi

  #Spin up another container with multiple IPs.
  lxc launch testimage "${ct_name}2" -p "${ct_name}"
  lxc exec "${ct_name}2" -- ip addr add "192.0.2.2${ipRand}/24" dev eth0
  lxc exec "${ct_name}2" -- ip addr add "2001:db8::2${ipRand}/64" dev eth0

  # Check comms between containers.
  lxc exec "${ct_name}" -- ping -c2 -W1 "192.0.2.2${ipRand}"
  lxc exec "${ct_name}" -- ping6 -c2 -W1 "2001:db8::2${ipRand}"
  lxc exec "${ct_name}2" -- ping -c2 -W1 "192.0.2.1${ipRand}"
  lxc exec "${ct_name}2" -- ping6 -c2 -W1 "2001:db8::1${ipRand}"

  # Test hot plugging a container nic with different settings to profile with the same name.
  lxc config device add "${ct_name}" eth0 nic \
    nictype=macvlan \
    name=eth0 \
    parent="${ct_name}" \
    mtu=1401

  # Check custom MTU is applied on hot-plug.
  if ! lxc exec "${ct_name}" -- ip link show eth0 | grep "mtu 1401" ; then
    echo "mtu invalid"
    false
  fi

  lxc config device remove "${ct_name}" eth0

  # Test hot plugging macvlan device based on vlan parent.
  ip link set "${ct_name}" up
  lxc config device add "${ct_name}" eth0 nic \
    nictype=macvlan \
    parent="${ct_name}" \
    name=eth0 \
    vlan=10 \
    mtu=1402

  # Check custom MTU is applied.
  if ! lxc exec "${ct_name}" -- ip link show eth0 | grep "mtu 1402" ; then
    echo "mtu invalid"
    false
  fi

  # Cleanup.
  lxc config device remove "${ct_name}" eth0
  lxc delete "${ct_name}" -f
  lxc delete "${ct_name}2" -f
  ip link delete "${ct_name}" type dummy
}
