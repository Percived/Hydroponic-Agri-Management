package telemetry

import "sync"

type CachedRecord struct {
	SensorChannelID uint64
	MetricCode      string
	Value           float64
	QualityFlag     string
	CollectedAtUnix int64
}

type SensorStatusCache struct {
	mu   sync.RWMutex
	data map[uint64]*CachedRecord // key: sensor_channel_id
}

func NewSensorStatusCache() *SensorStatusCache {
	return &SensorStatusCache{
		data: make(map[uint64]*CachedRecord),
	}
}

func (c *SensorStatusCache) Set(record CachedRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := record
	c.data[record.SensorChannelID] = &cp
}

func (c *SensorStatusCache) Get(channelID uint64) (*CachedRecord, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	rec, ok := c.data[channelID]
	return rec, ok
}

func (c *SensorStatusCache) GetMulti(channelIDs []uint64) map[uint64]*CachedRecord {
	if len(channelIDs) == 0 {
		return map[uint64]*CachedRecord{}
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make(map[uint64]*CachedRecord, len(channelIDs))
	for _, id := range channelIDs {
		if rec, ok := c.data[id]; ok {
			result[id] = rec
		}
	}
	return result
}
