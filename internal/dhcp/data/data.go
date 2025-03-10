// Package data is an interface between DHCP backend implementations and the DHCP server.
package data

import (
	"net"
	"net/netip"
	"net/url"
	"strings"

	"github.com/insomniacslk/dhcp/dhcpv4"
	"go.opentelemetry.io/otel/attribute"
)

// Packet holds the data that is passed to a DHCP handler.
type Packet struct {
	// Peer is the address of the client that sent the DHCP message.
	Peer net.Addr
	// Pkt is the DHCP message.
	Pkt *dhcpv4.DHCPv4
	// Md is the metadata that was passed to the DHCP server.
	Md *Metadata
}

// Metadata holds metadata about the DHCP packet that was received.
type Metadata struct {
	// IfName is the name of the interface that the DHCP message was received on.
	IfName string
	// IfIndex is the index of the interface that the DHCP message was received on.
	IfIndex int
}

// DHCP holds the DHCP headers and options to be set in a DHCP handler response.
// This is the API between a DHCP handler and a backend.
type DHCP struct {
	MACAddress       net.HardwareAddr // chaddr DHCP header.
	IPAddress        netip.Addr       // yiaddr DHCP header.
	SubnetMask       net.IPMask       // DHCP option 1.
	DefaultGateway   netip.Addr       // DHCP option 3.
	NameServers      []net.IP         // DHCP option 6.
	Hostname         string           // DHCP option 12.
	DomainName       string           // DHCP option 15.
	BroadcastAddress netip.Addr       // DHCP option 28.
	NTPServers       []net.IP         // DHCP option 42.
	VLANID           string           // DHCP option 43.116.
	LeaseTime        uint32           // DHCP option 51.
	Arch             string           // DHCP option 93.
	DomainSearch     []string         // DHCP option 119.
	Disabled         bool             // If true, no DHCP response should be sent.
}

// Netboot holds info used in netbooting a client.
type Netboot struct {
	AllowNetboot  bool     `yaml:"allow_pxe,omitempty"`       // If true, the client will be provided netboot options in the DHCP offer/ack.
	IPXEScriptURL *url.URL `yaml:"ipxe_script_url,omitempty"` // Overrides a default value that is passed into DHCP on startup.
	IPXEScript    string   `yaml:"ipxe_script,omitempty"`     // Overrides a default value that is passed into DHCP on startup.
	Console       string   `yaml:"console,omitempty"`
	Facility      string   `yaml:"facility,omitempty"`
	OSIE          OSIE     `yaml:"osie,omitempty"`
}

// OSIE or OS Installation Environment is the data about where the OSIE parts are located.
type OSIE struct {
	// BaseURL is the URL where the OSIE parts are located.
	BaseURL *url.URL
	// Kernel is the name of the kernel file.
	Kernel string
	// Initrd is the name of the initrd file.
	Initrd string
}

type Power struct {
	// State is the power state of the machine.
	State string
	// DeviceId is the device ID of the machine.
	DeviceId string
	// SiteId is the site ID of the machine.
	SiteId string
	// Port is the port of the machine.
	Port int
	// Mode
	Mode string
}

// EncodeToAttributes returns a slice of opentelemetry attributes that can be used to set span.SetAttributes.
func (d *DHCP) EncodeToAttributes() []attribute.KeyValue {
	var ns []string
	for _, e := range d.NameServers {
		ns = append(ns, e.String())
	}

	var ntp []string
	for _, e := range d.NTPServers {
		ntp = append(ntp, e.String())
	}

	var ip string
	if d.IPAddress.Compare(netip.Addr{}) != 0 {
		ip = d.IPAddress.String()
	}

	var sm string
	if d.SubnetMask != nil {
		sm = net.IP(d.SubnetMask).String()
	}

	var dfg string
	if d.DefaultGateway.Compare(netip.Addr{}) != 0 {
		dfg = d.DefaultGateway.String()
	}

	var ba string
	if d.BroadcastAddress.Compare(netip.Addr{}) != 0 {
		ba = d.BroadcastAddress.String()
	}

	return []attribute.KeyValue{
		attribute.String("DHCP.MACAddress", d.MACAddress.String()),
		attribute.String("DHCP.IPAddress", ip),
		attribute.String("DHCP.SubnetMask", sm),
		attribute.String("DHCP.DefaultGateway", dfg),
		attribute.String("DHCP.NameServers", strings.Join(ns, ",")),
		attribute.String("DHCP.Hostname", d.Hostname),
		attribute.String("DHCP.DomainName", d.DomainName),
		attribute.String("DHCP.BroadcastAddress", ba),
		attribute.String("DHCP.NTPServers", strings.Join(ntp, ",")),
		attribute.Int64("DHCP.LeaseTime", int64(d.LeaseTime)),
		attribute.String("DHCP.DomainSearch", strings.Join(d.DomainSearch, ",")),
	}
}

// EncodeToAttributes returns a slice of opentelemetry attributes that can be used to set span.SetAttributes.
func (n *Netboot) EncodeToAttributes() []attribute.KeyValue {
	var s string

	if n == nil {
		return []attribute.KeyValue{}
	}

	if n.IPXEScriptURL != nil {
		s = n.IPXEScriptURL.String()
	}
	return []attribute.KeyValue{
		attribute.Bool("Netboot.AllowNetboot", n.AllowNetboot),
		attribute.String("Netboot.IPXEScriptURL", s),
	}
}

// EncodeToAttributes returns a slice of opentelemetry attributes that can be used to set span.SetAttributes.
func (o *OSIE) EncodeToAttributes() []attribute.KeyValue {
	var s string

	if o == nil {
		return []attribute.KeyValue{}
	}

	if o.BaseURL != nil {
		s = o.BaseURL.String()
	}
	return []attribute.KeyValue{
		attribute.String("OSIE.BaseURL", s),
		attribute.String("OSIE.Kernel", o.Kernel),
		attribute.String("OSIE.Initrd", o.Initrd),
	}
}

// EncodeToAttributes returns a slice of opentelemetry attributes that can be used to set span.SetAttributes.
func (p *Power) EncodeToAttributes() []attribute.KeyValue {

	if p == nil {
		return []attribute.KeyValue{}
	}

	return []attribute.KeyValue{
		attribute.String("Power.State", p.State),
		attribute.String("Power.DeviceID", p.DeviceId),
		attribute.String("Power.SiteID", p.SiteId),
		attribute.Int("Power.Port", p.Port),
		attribute.String("Power.Mode", p.Mode),
	}
}
