package metrics

import (
	"log"
	"time"
)

type Cleaner struct {
	m                         *Metrics
	removeWhenInactiveMinutes int
}

func NewCleaner(m *Metrics, removeWhenInactiveMinutes int) *Cleaner {
	return &Cleaner{m, removeWhenInactiveMinutes}
}

func (c *Cleaner) Start() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				c.m.pm.lock.Lock()
				for source, sourceMetrics := range c.m.pm.metrics {
					if time.Since(sourceMetrics.lastAccessed) > time.Duration(c.removeWhenInactiveMinutes)*time.Minute {
						statusTopic, topicExists := sourceMetrics.Metrics["status_topic"].(string)
						statusNetHostname, hostnameExists := sourceMetrics.Metrics["status_net_hostname"].(string)
						statusDeviceName, deviceNameExists := sourceMetrics.Metrics["status_device_name"].(string)
						for pmk := range sourceMetrics.Metrics {
							if topicExists && hostnameExists && deviceNameExists {
								if gauge, ok := c.m.gauges[pmk]; ok {
									gauge.DeleteLabelValues(source, statusTopic, statusNetHostname, statusDeviceName)
								}
							}
						}
						delete(c.m.pm.metrics, source)
						log.Println("removed inactive source: " + source)
					}
				}
				c.m.pm.lock.Unlock()
			}
		}
	}()
	log.Println("scheduled metrics cleanup")
}
