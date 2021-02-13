// +build windows

package disk

import (
	"strconv"

	"github.com/bluemedorapublic/gopsutil/disk"
	log "github.com/golang/glog"
)

// Call the Partitions function to get an array all drevices (local disk, remote, usb, cdrom)
// Only append valid drvies to the drives array (Only local disks)
func (c *Config) getDisks() error {
	devices, err := disk.Partitions(true)
	if err != nil {
		return err
	}

	for _, device := range devices {
		if validDrive(int(device.Typeret)) == true {

			d := Device{
				Name: device.Device,
				MountPoint: device.Mountpoint,
				Type: device.Fstype,
			}
			c.Host.Devices = append(c.Host.Devices, d)

			c.Host.Drives = append(c.Host.Drives, string(device.Mountpoint))
		}
	}

	return nil
}

// Kick off an alert for each drive that has a high consumption
func (c Config) getUsage() error {
	var (
		createAlert bool   = false
		createLock  bool   = false
		message     string = c.Host.Name
	)

	for _, drive := range c.Host.Drives {

		fs, _ := disk.Usage(drive + "\\")
		log.Info(fs.Path, int(fs.UsedPercent), "%")
		usedSpace := strconv.Itoa(int(fs.UsedPercent)) + "%"

		if int(fs.UsedPercent) > c.Threshold {
			message = message + " high disk usage on drive " + drive + " " + usedSpace
			log.Info(message)
			createAlert = true
			createLock = true

		} else {
			log.Info("Disk usage healthy: ", drive)
		}
	}

	return c.handleLock(createLock, createAlert, message)
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
