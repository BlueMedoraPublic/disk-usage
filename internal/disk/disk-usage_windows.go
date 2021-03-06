// +build windows

package disk

import (
	"fmt"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows"
)

// Call the Partitions function to get an array all drevices (local disk, remote, usb, cdrom)
// Only append valid drvies to the drives array (Only local disks)
func (c *Config) getDisks() error {
	devices, err := disk.Partitions(true)
	if err != nil {
		return err
	}

	for _, device := range devices {
		path, err := windows.UTF16PtrFromString(device.Mountpoint)
		if err != nil {
			return err
		}

		typeRet := windows.GetDriveType(path)

		if validDrive(int(typeRet)) {
			d := Device{
				Name:       device.Device,
				MountPoint: device.Mountpoint,
				Type:       device.Fstype,
			}
			c.Host.Devices = append(c.Host.Devices, d)
		}
	}

	return nil
}

// Kick off an alert for each drive that has a high consumption
func (c *Config) getUsage() error {
	for i, device := range c.Host.Devices {
		path := device.MountPoint

		fs, err := disk.Usage(path + "\\")
		if err != nil {
			log.Error(fmt.Sprintf("failed to read path %s: %s", path, err.Error()))
			continue
		}
		percentage := int(fs.UsedPercent)
		c.Host.Devices[i].UsagePercent = percentage
		log.Trace(fmt.Sprintf("disk %s usage %d", device.Name, percentage))

	}
	return nil
}

func getDevType(driveType uintptr) string {
	switch driveType {
	case 2:
		return "Removable"
	case 3:
		return "Local"
	case 4:
		return "Network"
	case 5:
		return "CDROM"
	default:
		return "Unknown"
	}
}

// Local drives are the only drives that should be considered for alerting
func validDrive(driveType int) bool {
	if driveType == 3 {
		return true
	} else {
		return false
	}
}
