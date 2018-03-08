package main

import "syscall"
import "bufio"
import "os"
import "strings"
import "runtime"
import "log"

var supported_fs map[string]bool

func getPathSizeStats(path string) (ServerFilesystem, error) {
	stat := syscall.Statfs_t{}
	err := syscall.Statfs(path, &stat)
	if err != nil {
		return ServerFilesystem{}, err
	}
	blocksize := uint64(stat.Bsize)
	fs := ServerFilesystem{
		Path: path,
		Size: blocksize * uint64(stat.Blocks),
		Used: uint64(stat.Blocks-stat.Bavail) * blocksize,
		Free: uint64(stat.Bavail) * uint64(blocksize),
		Pct:  uint64(((float64(stat.Blocks) - float64(stat.Bavail)) / float64(stat.Blocks)) * 100)}
	return fs, nil
}

func getMountsLinux() ([]string, error) {
	var paths []string
	file, err := os.Open("/proc/mounts")
	if err != nil {
		return paths, nil
	}
	defer file.Close()

	seen := map[string]bool{}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		parts := strings.Split(scanner.Text(), " ")
		// Sanity check to avoid choking on missing indexes
		if len(parts) < 5 {
			continue
		}
		device := parts[0]
		mount := parts[1]
		filesystem := parts[2]
		flags := strings.Split(parts[3], ",")
		// Ignore pointless FSs
		if !supported_fs[filesystem] {
			continue
		}
		// Ignore read-only FSs
		readOnly := false
		for _, option := range flags {
			if option == "ro" {
				readOnly = true
				break
			}
		}
		if readOnly {
			continue
		}
		// Try to avoid catching bindmounts by deduping on device
		if seen[device] {
			continue
		} else {
			seen[device] = true
			paths = append(paths, mount)
		}
	}
	return paths, nil
}

func GetFilesystems() ([]ServerFilesystem, error) {
	filesystems := []ServerFilesystem{}

	var mounts []string

	if runtime.GOOS == "linux" {
		mounts, _ = getMountsLinux()
	} else {
		log.Printf("Unsupported OS for getting filesystem info")
		return filesystems, nil
	}

	for _, mount := range mounts {
		stats, err := getPathSizeStats(mount)
		if err == nil {
			filesystems = append(filesystems, stats)
		}
	}

	return filesystems, nil
}

func init() {
	// Whitelist sane filesystems to keep track of to
	// ignore ones that don't matter at all
	supported_fs = map[string]bool{}
	supported_fs["ext4"] = true
	supported_fs["ext3"] = true
	supported_fs["ext2"] = true
	supported_fs["jfs"] = true
	supported_fs["xfs"] = true
	supported_fs["zfs"] = true
	supported_fs["vfat"] = true
	supported_fs["hfs"] = true
}
