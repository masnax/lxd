(network-macvlan)=
# Macvlan network

The macvlan network type allows one to specify presets to use when connecting instances to a parent interface
using macvlan NICs. This allows the instance NIC itself to simply specify the `network` it is connecting to without
knowing any of the underlying configuration details.

(network-macvlan-options)=
## Configuration options

Key                             | Type      | Condition             | Default                   | Description
:--                             | :--       | :--                   | :--                       | :--
maas.subnet.ipv4                | string    | ipv4 address          | -                         | MAAS IPv4 subnet to register instances in (when using `network` property on nic)
maas.subnet.ipv6                | string    | ipv6 address          | -                         | MAAS IPv6 subnet to register instances in (when using `network` property on nic)
mtu                             | integer   | -                     | -                         | The MTU of the new interface
parent                          | string    | -                     | -                         | Parent interface to create macvlan NICs on
vlan                            | integer   | -                     | -                         | The VLAN ID to attach to
gvrp                            | boolean   | -                     | false                     | Register VLAN using GARP VLAN Registration Protocol
