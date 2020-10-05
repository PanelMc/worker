package container

import (
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/PanelMc/worker"
	"github.com/docker/docker/api/types"
)

func (c *dockerContainer) Stats() (worker.ContainerStats, error) {
	stats, err := c.stats(false, time.Millisecond*100)
	if err != nil {
		return worker.ContainerStats{}, err
	}

	return *<-stats, nil
}

func (c *dockerContainer) StatsChan() (<-chan *worker.ContainerStats, error) {
	if c.statsChan == nil {
		stats, err := c.stats(true, time.Second*1)
		if err != nil {
			return nil, err
		}

		c.statsChan = stats
	}

	return c.statsChan, nil
}

func (c *dockerContainer) stats(stream bool, delay time.Duration) (<-chan *worker.ContainerStats, error) {
	stats, err := c.client.ContainerStats(context.TODO(), c.ContainerID, false)
	if err != nil {
		return nil, err
	}

	statsChan := make(chan *worker.ContainerStats)

	go func() {
		defer func() {
			close(statsChan)
			stats.Body.Close()
		}()

		daemonOSType := stats.OSType
		dec := json.NewDecoder(stats.Body)
		var v *types.StatsJSON

		for {
			if err := dec.Decode(&v); err != nil {
				if err == io.EOF {
					// No more content, exit loop and close everything
					break
				}

				// Create a new decoder with the remaining data from the current decoder
				// in combination with the stats stream reader
				dec = json.NewDecoder(io.MultiReader(dec.Buffered(), stats.Body))

				time.Sleep(delay)
				continue
			}

			statsChan <- mapStats(daemonOSType, v)

			if !stream {
				break
			}
		}
	}()

	return statsChan, nil
}

func mapStats(daemonOSType string, v *types.StatsJSON) *worker.ContainerStats {
	var cpuPercent, memPerc float64
	var blkRead, blkWrite, mem, memLimit uint64

	if daemonOSType != "windows" {
		// MemoryStats.Limit will never be 0 unless the container is not running and we haven't
		// got any data from cgroup
		if v.MemoryStats.Limit != 0 {
			memPerc = float64(v.MemoryStats.Usage) / float64(v.MemoryStats.Limit) * 100.0
		}
		cpuPercent = calculateCPUPercentUnix(v.PreCPUStats.CPUUsage.TotalUsage, v.PreCPUStats.SystemUsage, v)
		blkRead, blkWrite = calculateBlockIO(v.BlkioStats)
		mem = v.MemoryStats.Usage
		memLimit = v.MemoryStats.Limit
	} else {
		cpuPercent = calculateCPUPercentWindows(v)
		blkRead = v.StorageStats.ReadSizeBytes
		blkWrite = v.StorageStats.WriteSizeBytes
		mem = v.MemoryStats.PrivateWorkingSet
	}
	netRx, netTx := calculateNetwork(v.Networks)

	return &worker.ContainerStats{
		CPUPercentage:    cpuPercent,
		Memory:           mem,
		MemoryPercentage: memPerc,
		MemoryLimit:      memLimit,
		NetworkDownload:  netRx,
		NetworkUpload:    netTx,
		DiscRead:         blkRead,
		DiscWrite:        blkWrite,
	}
}

func calculateCPUPercentUnix(previousCPU, previousSystem uint64, v *types.StatsJSON) float64 {
	var (
		cpuPercent = 0.0
		// calculate the change for the cpu usage of the container in between readings
		cpuDelta = float64(v.CPUStats.CPUUsage.TotalUsage) - float64(previousCPU)
		// calculate the change for the entire system between readings
		systemDelta = float64(v.CPUStats.SystemUsage) - float64(previousSystem)
	)

	if systemDelta > 0.0 && cpuDelta > 0.0 {
		cpuPercent = (cpuDelta / systemDelta) * float64(len(v.CPUStats.CPUUsage.PercpuUsage)) * 100.0
	}
	return cpuPercent
}

func calculateCPUPercentWindows(v *types.StatsJSON) float64 {
	// Max number of 100ns intervals between the previous time read and now
	possIntervals := uint64(v.Read.Sub(v.PreRead).Nanoseconds()) // Start with number of ns intervals
	possIntervals /= 100                                         // Convert to number of 100ns intervals
	possIntervals *= uint64(v.NumProcs)                          // Multiple by the number of processors

	// Intervals used
	intervalsUsed := v.CPUStats.CPUUsage.TotalUsage - v.PreCPUStats.CPUUsage.TotalUsage

	// Percentage avoiding divide-by-zero
	if possIntervals > 0 {
		return float64(intervalsUsed) / float64(possIntervals) * 100.0
	}
	return 0.00
}

func calculateBlockIO(blkio types.BlkioStats) (blkRead uint64, blkWrite uint64) {
	for _, bioEntry := range blkio.IoServiceBytesRecursive {
		switch strings.ToLower(bioEntry.Op) {
		case "read":
			blkRead = blkRead + bioEntry.Value
		case "write":
			blkWrite = blkWrite + bioEntry.Value
		}
	}
	return
}

func calculateNetwork(network map[string]types.NetworkStats) (uint64, uint64) {
	var rx, tx uint64

	for _, v := range network {
		rx += v.RxBytes
		tx += v.TxBytes
	}

	return rx, tx
}
