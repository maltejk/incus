(network-ovn)=
# OVN network

<!-- Include start OVN intro -->
{abbr}`OVN (Open Virtual Network)` is a software-defined networking system that supports virtual network abstraction.
You can use it to build your own private cloud.
See [`www.ovn.org`](https://www.ovn.org/) for more information.
<!-- Include end OVN intro -->

The `ovn` network type allows to create logical networks using the OVN {abbr}`SDN (software-defined networking)`.
This kind of network can be useful for labs and multi-tenant environments where the same logical subnets are used in multiple discrete networks.

An Incus OVN network can be connected to an existing managed {ref}`network-bridge` or {ref}`network-physical` to gain access to the wider network.
By default, all connections from the OVN logical networks are NATed to an IP allocated from the uplink network.

See {ref}`network-ovn-setup` for basic instructions for setting up an OVN network.

## Unmanaged OVN networks

Incus can detect and use existing OVN logical switches as unmanaged networks.
These networks appear in the network list and can be used as uplinks for managed OVN networks or for direct NIC attachment to instances.

Unmanaged OVN networks are automatically detected from the OVN Northbound database.
They support the same features as managed OVN networks for NIC attachment, but do not allow configuration changes.

In case of OVN unavailability, instances connected to unmanaged OVN networks will lose network connectivity but remain running.
Errors are logged, and operations may retry transient failures.

### Infrastructure Requirements

For unmanaged OVN networks to provide full functionality, the following OVN infrastructure must be pre-configured:

- **Logical Switch**: The switch must exist with the desired name.
- **Logical Router**: A router should be attached to the switch for external connectivity.
- **DHCP Configuration**: DHCP options must be set on the switch for automatic IP assignment to instances.
- **Router Ports**: Proper linking between switch and router ports.
- **NAT Rules** (optional): For outbound internet access from instances.

Without these, instances may not receive IP addresses or have external connectivity.

### Setup Examples

Use `ovn-nbctl` commands to create the required infrastructure. Here's an example setup for a switch named `my-ovn-switch` with subnet `192.168.1.0/24`:

1. Create the logical switch:

   ```bash
   ovn-nbctl ls-add my-ovn-switch
   ```

2. Create a logical router:

   ```bash
   ovn-nbctl lr-add my-ovn-router
   ```

3. Create router port (internal interface):

   ```bash
   ovn-nbctl lrp-add my-ovn-router my-ovn-router-port 192.168.1.1/24
   ```

4. Create switch port for the router:

   ```bash
   ovn-nbctl lsp-add my-ovn-switch my-ovn-router-lsp
   ovn-nbctl lsp-set-addresses my-ovn-router-lsp router
   ovn-nbctl lsp-set-type my-ovn-router-lsp router
   ovn-nbctl lsp-set-options my-ovn-router-lsp router-port=my-ovn-router-port
   ```

5. Create DHCP options:

   ```bash
   DHCP_UUID=$(ovn-nbctl dhcp-options-create 192.168.1.0/24)
   ovn-nbctl dhcp-options-set-options $DHCP_UUID \
     server_id=192.168.1.1 \
     server_mac=aa:bb:cc:dd:ee:ff \
     lease_time=3600 \
     router=192.168.1.1
   ovn-nbctl ls-set-dhcpv4-options my-ovn-switch $DHCP_UUID
   ```

6. Add NAT rule for external access (assuming uplink interface):

   ```bash
   ovn-nbctl lr-nat-add my-ovn-router snat 192.168.1.1 192.168.1.0/24
   ```

### Naming Conventions

Incus does not enforce specific naming for unmanaged OVN switches, but for consistency with managed networks, consider naming them descriptively (e.g., `incus-unmanaged-<purpose>`). Managed OVN networks created by Incus use the pattern `incus-net-<network-name>`.

### Testing Unmanaged OVN Networks

To verify unmanaged OVN network functionality:

1. **Detection**: Run `incus network list` to confirm the unmanaged network appears.
2. **NIC Attachment**: Create an instance with a `nic` device attached to the unmanaged network.
3. **IP Assignment**: Check that the instance receives an IP via DHCP.
4. **Connectivity**: Test internal connectivity between instances on the same switch.
5. **External Access**: Verify outbound connectivity through the router.
6. **Error Handling**: Simulate OVN database unavailability and observe logged errors without instance crashes.
7. **Uplink Usage**: Use the unmanaged network as an uplink for a managed OVN network.

### Using OVN as a Layer 2 Switch

In scenarios where routing, DHCP, and DNS are already handled by existing network infrastructure (such as VXLAN networks, external routers, or dedicated DHCP/DNS servers), OVN can be used purely as a layer 2 switch. This allows Incus instances to integrate seamlessly with legacy or complex network setups without duplicating network services.

#### Key Characteristics

- **No OVN Routing**: Logical routers are not required; the switch operates at layer 2 only.
- **External IP Management**: DHCP and DNS are provided by external systems, not configured in OVN.
- **Layer 2 Connectivity**: Instances on the switch can communicate at the data link layer, with higher-level services handled externally.
- **Integration Benefits**: Useful for connecting to existing VXLAN tunnels, corporate networks, or multi-vendor environments.

#### Setup for Layer 2 Only

Create a basic logical switch without routers or DHCP options:

```bash
ovn-nbctl ls-add my-l2-switch
```

To integrate with existing VXLAN infrastructure:

- Ensure the OVN logical switch is connected to the appropriate chassis or external ports.
- Use OVN's support for Geneve or VXLAN encapsulation if needed for tunneling.
- Configure external routers or switches to route traffic to/from the OVN switch.

For example, to connect the logical switch to an existing VXLAN network with VNI 10824 and remote IP 192.168.1.100:

```bash
# Add a VXLAN port to the switch
ovn-nbctl lsp-add my-l2-switch vxlan-tunnel

# Set the port type to VXLAN
ovn-nbctl lsp-set-type vxlan-tunnel vxlan

# Configure VXLAN options
ovn-nbctl lsp-set-options vxlan-tunnel \
  remote_ip=192.168.1.100 \
  vni=10824
```

This creates a tunnel port that connects the OVN logical switch to the VXLAN infrastructure. Adjust the `remote_ip` to match your VXLAN endpoint and ensure the chassis running ovn-controller is configured to handle VXLAN encapsulation.

For unmanaged networks, the switch must already exist and be properly connected in your OVN deployment. Incus will detect it and allow NIC attachment, but all layer 3+ configuration remains external.

% Include content from [network_bridge.md](network_bridge.md)
```{include} network_bridge.md
    :start-after: <!-- Include start MAC identifier note -->
    :end-before: <!-- Include end MAC identifier note -->
```

(network-ovn-options)=

## Configuration options

The following configuration key namespaces are currently supported for the `ovn` network type:

- `bridge` (L2 interface configuration)
- `dns` (DNS server and resolution configuration)
- `ipv4` (L3 IPv4 configuration)
- `ipv6` (L3 IPv6 configuration)
- `security` (network ACL configuration)
- `user` (free-form key/value for user metadata)

```{note}
{{note_ip_addresses_CIDR}}
```

The following configuration options are available for the `ovn` network type:

% Include content from [config_options.txt](../config_options.txt)
```{include} ../config_options.txt
    :start-after: <!-- config group network_ovn-common start -->
    :end-before: <!-- config group network_ovn-common end -->
```

```{note}
The `bridge.external_interfaces` option supports an extended format allowing the creation of missing VLAN interfaces.
The extended format is `<interfaceName>/<parentInterfaceName>/<vlanId>`.
When the external interface is added to the list with the extended format, the system will automatically create the interface upon the network's creation and subsequently delete it when the network is terminated. The system verifies that the `<interfaceName>` does not already exist. If the interface name is in use with a different parent or VLAN ID, or if the creation of the interface is unsuccessful, the system will revert with an error message.
```

(network-ovn-features)=

## Supported features

The following features are supported for the `ovn` network type:

- {ref}`network-acls`
- {ref}`network-forwards`
- {ref}`network-integrations`
- {ref}`network-zones`
- {ref}`network-ovn-peers`
- {ref}`network-load-balancers`

```{toctree}
:maxdepth: 1
:hidden:

Set up OVN </howto/network_ovn_setup>
Create routing relationships </howto/network_ovn_peers>
Configure network load balancers </howto/network_load_balancers>
```
