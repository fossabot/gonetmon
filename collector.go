package main

import (
	"errors"
	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
	log "github.com/sirupsen/logrus"
	"net"
	"strings"
	"sync"
)

// TODO : don't keep values here
var (
	snapshotLen int32 = 1024
	promiscuous       = false
	timeout           = defDisplayRefresh //10 * time.Second
)

type Devices struct {
	devices		[]net.Interface
	handles		[]*pcap.Handle
}

// InitialiseCapture opens device interfaces and associated handles to listen on, returns a map of these.
// If the interfaces parameter is not nil, only open those specified.
func InitialiseCapture(interfaces []string) (*Devices, error) {

	var err error

	devices := findDevices(interfaces)

	if devices == nil {
		return nil, err
	}

	devs := &Devices{
		devices: []net.Interface{},
		handles: []*pcap.Handle{},
	}
	err = nil
	for _, d := range devices {
		// Try to open all devices for capture
		if h, err := openDevice(d); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Error("Could not open device for capture.")
		} else {
			devs.devices = append(devs.devices, d)
			devs.handles = append(devs.handles, h)
		}
	}

	if len(devs.devices) == 0 {
		log.Error("Could not open any device interface.")
		return nil, errors.New("could not open any device interface")
	}

	return devs, nil
}

// findDevices gathers the list of interfaces of the machine that have their state flage UP.
// If the interfaces parameter is not nil, only list those specified if present.
func findDevices(interfaces []string) []net.Interface {
	devices, err := net.Interfaces()

	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Error("Error in finding network devices.")
		return nil
	}

	if len(devices) == 0 {
		log.Error("Could not find any network devices (but no error occurred).")
		return nil
	}

	// Purge interfaces that don't have their state flag UP
	for index, d := range devices {
		if d.Flags&(net.FlagUp) == 0 {
			// Flag is down, Interface is deactivated, purge element
			devices[index] = devices[len(devices)-1]
			devices = devices[:len(devices)-1]
		}
	}

	// If we want a custom list of interfaces
	if interfaces != nil {
		var tailoredList []net.Interface

		interfacesLoop:
		for _, i := range interfaces {

			for index, d := range devices {

				if d.Name == i {
					tailoredList = append(tailoredList, d)

					// Remove the found element from array to avoid it on next iteration
					// Won't affect current loop since Go uses a copy
					devices[index] = devices[len(devices)-1]
					devices = devices[:len(devices)-1]

					log.Info("Found requested interface ", i)

					continue interfacesLoop
				}
			}

			// Here, the requested interface is not in the found set
			log.Error("Could not find requested interface among activated interfaces : ", i)
		}

		if len(tailoredList) == 0 {
			log.Error("Could not find any requested network devices among : ", interfaces)
			return nil
		}

		devices = tailoredList
	}

	return devices
}

// openDevice opens a live listener on the interface designated by the device parameter and returns a corresponding handle
func openDevice(device net.Interface) (*pcap.Handle, error) {
	handle, err := pcap.OpenLive(device.Name, snapshotLen, promiscuous, timeout)
	if err != nil {
		log.WithFields(log.Fields{
			"interface": device.Name,
			"error":     err,
		}).Error("Could not open device.")

		return nil, err
	}

	log.WithFields(log.Fields{
		"interface": device.Name,
	}).Info("Opened device interface.")

	return handle, nil
}

// Closes listening on a device
func closeDevice(h *pcap.Handle) {
	h.Close()
}

func closeDevices(devices *Devices) {
	for index, dev := range devices.devices {
		log.Info("Closing device on interface ", dev.Name)
		closeDevice(devices.handles[index])
	}
}

// addFilter adds a BPF filter to the handle to filter sniffed traffic
func addFilter(handle *pcap.Handle, filter string) error {
	return handle.SetBPFFilter(filter)
}

// sniffApplicationLayer tells whether the packet contains the filter string
func sniffApplicationLayer(packet gopacket.Packet, filter string) bool {
	var isApp = false
	applicationLayer := packet.ApplicationLayer()
	if applicationLayer != nil {
		payload := applicationLayer.Payload()
		if strings.Contains(string(payload), filter) {
			isApp = true
		}
	}

	return isApp
}


// getRemoteIP extracts the IP address of the remote peer from packet
func getRemoteIP(packet gopacket.Packet, deviceIP string) string {
	src, dst := packet.NetworkLayer().NetworkFlow().Endpoints()

	// The deviceIP is among these two, so we return the other
	if strings.Compare(deviceIP, src.String()) == 0{
		return dst.String()
	}
	return src.String()
}


// getDeviceIP extracts the interface's local IP address
func getDeviceIP(device *net.Interface) (string, error) {
	add, err := device.Addrs()
	if err != nil {
		return "", err
	}
	address := add[0].String()[:strings.IndexByte(add[0].String(), '/')]
	return address, nil
}


// capturePacket continuously listens to a device interface managed by handle, and extracts relevant packets from traffic
// to send it to packetChan
func capturePackets(device net.Interface, handle *pcap.Handle, filter *Filter, wg *sync.WaitGroup, packetChan chan<- packetMsg) {
	defer wg.Done()

	log.Info("Capturing packets on ", device.Name)

	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())

	// This will loop on a channel that will send packages, and will quit when the handle is closed by another caller
	for packet := range packetSource.Packets() {
		if sniffApplicationLayer(packet, filter.Application) {

			ip, err := getDeviceIP(&device)
			if err != nil {
				log.WithFields(log.Fields{
					"interface": device.Name,
					"error":     err,
				}).Error("Could not extract IP from local network interface")
			}

			packetChan <- packetMsg{
				dataType:  filter.Type,
				device:    device.Name,
				deviceIP: ip,
				remoteIP: getRemoteIP(packet, ip),
				rawPacket: packet,
			}
		}
	}

	log.Info("Stopping capture on ", device.Name)
}

// Collector listens on all network devices for relevant traffic and sends packets to packetChan
func Collector(parameters *Parameters, devices *Devices, packetChan chan packetMsg, syncChan <-chan struct{}, syncwg *sync.WaitGroup) {

	wg := sync.WaitGroup{}

	for index, dev := range devices.devices {
		wg.Add(1)
		h := devices.handles[index]
		if err := addFilter(h, parameters.PacketFilter.Network); err != nil {
			log.WithFields(log.Fields{
				"interface": dev.Name,
				"error":     err,
			}).Error("Could not set filter on device. Closing.")
			closeDevice(h)
		}
		go capturePackets(dev, h, &parameters.PacketFilter, &wg, packetChan)
	}

	// Wait until sync to stop
	<-syncChan

	// Inform goroutines to stop
	closeDevices(devices)

	// Wait for goroutines to stop
	log.Info("Collector waiting for subs...")
	wg.Wait()
	log.Info("Collector terminating")
	syncwg.Done()
}
