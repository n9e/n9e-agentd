// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-present Datadog, Inc.

package dogstatsd

import (
	"bytes"
	"encoding/json"
	"expvar"
	"fmt"
	"net"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/telemetry"
	telemetry_utils "github.com/n9e/n9e-agentd/pkg/telemetry/utils"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/aggregator/ckey"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/dogstatsd/listeners"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/dogstatsd/mapper"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/metrics"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/status/health"
	"github.com/n9e/n9e-agentd/staging/datadog-agent/pkg/util"
	"k8s.io/klog/v2"
)

var (
	dogstatsdExpvars                  = expvar.NewMap("statsd")
	dogstatsdServiceCheckParseErrors  = expvar.Int{}
	dogstatsdServiceCheckPackets      = expvar.Int{}
	dogstatsdEventParseErrors         = expvar.Int{}
	dogstatsdEventPackets             = expvar.Int{}
	dogstatsdMetricParseErrors        = expvar.Int{}
	dogstatsdMetricPackets            = expvar.Int{}
	dogstatsdPacketsLastSec           = expvar.Int{}
	dogstatsdUnterminatedMetricErrors = expvar.Int{}

	tlmProcessed = telemetry.NewCounter("statsd", "processed",
		[]string{"message_type", "state"}, "Count of service checks/events/metrics processed by dogstatsd")
	tlmProcessedErrorTags = map[string]string{"message_type": "metrics", "state": "error"}
	tlmProcessedOkTags    = map[string]string{"message_type": "metrics", "state": "ok"}
)

func init() {
	dogstatsdExpvars.Set("ServiceCheckParseErrors", &dogstatsdServiceCheckParseErrors)
	dogstatsdExpvars.Set("ServiceCheckPackets", &dogstatsdServiceCheckPackets)
	dogstatsdExpvars.Set("EventParseErrors", &dogstatsdEventParseErrors)
	dogstatsdExpvars.Set("EventPackets", &dogstatsdEventPackets)
	dogstatsdExpvars.Set("MetricParseErrors", &dogstatsdMetricParseErrors)
	dogstatsdExpvars.Set("MetricPackets", &dogstatsdMetricPackets)
	dogstatsdExpvars.Set("UnterminatedMetricErrors", &dogstatsdUnterminatedMetricErrors)
}

// Server represent a statsd server
type Server struct {
	// listeners are the instantiated socket listener (UDS or UDP or both)
	listeners []listeners.StatsdListener
	// aggregator is a pointer to the aggregator that the dogstatsd daemon
	// will send the metrics samples, events and service checks to.
	aggregator *aggregator.BufferedAggregator
	// running in their own routine, workers are responsible of parsing the packets
	// and pushing them to the aggregator
	workers []*worker

	packetsIn                 chan listeners.Packets
	sharedPacketPool          *listeners.PacketPool
	sharedFloat64List         *float64ListPool
	Statistics                *util.Stats
	Started                   bool
	stopChan                  chan bool
	health                    *health.Handle
	metricPrefix              string
	metricPrefixBlacklist     []string
	defaultHostname           string
	histToDist                bool
	histToDistPrefix          string
	extraTags                 []string
	Debug                     *dsdServerDebug
	mapper                    *mapper.MetricMapper
	eolTerminationUDP         bool
	eolTerminationUDS         bool
	eolTerminationNamedPipe   bool
	telemetryEnabled          bool
	entityIDPrecedenceEnabled bool
	// disableVerboseLogs is a feature flag to disable the logs capable
	// of flooding the logger output (e.g. parsing messages error).
	// NOTE(remy): this should probably be dropped and use a throttler logger, see
	// package (pkg/trace/logutils) for a possible throttler implemetation.
	disableVerboseLogs bool

	// ServerlessMode is set to true if we're running in a serverless environment.
	ServerlessMode     bool
	UdsListenerRunning bool
}

// metricStat holds how many times a metric has been
// processed and when was the last time.
type metricStat struct {
	Name     string    `json:"name"`
	Count    uint64    `json:"count"`
	LastSeen time.Time `json:"lastSeen"`
	Tags     string    `json:"tags"`
}

type dsdServerDebug struct {
	sync.Mutex
	// Enabled is an atomic int used as a boolean
	Enabled uint64                         `json:"enabled"`
	Stats   map[ckey.ContextKey]metricStat `json:"stats"`
	// counting number of metrics processed last X seconds
	metricsCounts metricsCountBuckets
	// keyGen is used to generate hashes of the metrics received by dogstatsd
	keyGen *ckey.KeyGenerator
}

// metricsCountBuckets is counting the amount of metrics received for the last 5 seconds.
// It is used to detect spikes.
type metricsCountBuckets struct {
	counts     [5]uint64
	bucketIdx  int
	currentSec time.Time
	metricChan chan struct{}
	closeChan  chan struct{}
}

// NewServer returns a running DogStatsD server.
// If extraTags is nil, they will be read from DD_DOGSTATSD_TAGS if set.
func NewServer(aggregator *aggregator.BufferedAggregator, extraTags []string) (*Server, error) {
	var stats *util.Stats
	cf := config.C.Statsd
	if cf.StatsEnable {
		buff := cf.StatsBuffer
		s, err := util.NewStats(uint32(buff))
		if err != nil {
			klog.Errorf("statsd: unable to start statistics facilities")
		}
		stats = s
		dogstatsdExpvars.Set("PacketsLastSecond", &dogstatsdPacketsLastSec)
	}

	var metricsStatsEnabled uint64 // we're using an uint64 for its atomic capacity
	if cf.MetricsStatsEnable {
		klog.Info("statsd: metrics statistics will be stored.")
		metricsStatsEnabled = 1
	}

	packetsChannel := make(chan listeners.Packets, cf.QueueSize)
	tmpListeners := make([]listeners.StatsdListener, 0, 2)

	// sharedPacketPool is used by the packet assembler to retrieve already allocated
	// buffer in order to avoid allocation. The packets are pushed back by the server.
	sharedPacketPool := listeners.NewPacketPool(cf.BufferSize)

	udsListenerRunning := false

	socketPath := cf.Socket
	if len(socketPath) > 0 {
		unixListener, err := listeners.NewUDSListener(packetsChannel, sharedPacketPool)
		if err != nil {
			klog.Errorf(err.Error())
		} else {
			tmpListeners = append(tmpListeners, unixListener)
			udsListenerRunning = true
		}
	}
	if cf.Port > 0 {
		udpListener, err := listeners.NewUDPListener(packetsChannel, sharedPacketPool)
		if err != nil {
			klog.Errorf(err.Error())
		} else {
			tmpListeners = append(tmpListeners, udpListener)
		}
	}

	pipeName := cf.PipeName
	if len(pipeName) > 0 {
		namedPipeListener, err := listeners.NewNamedPipeListener(pipeName, packetsChannel, sharedPacketPool)
		if err != nil {
			klog.Errorf("named pipe error: %v", err.Error())
		} else {
			tmpListeners = append(tmpListeners, namedPipeListener)
		}
	}

	if len(tmpListeners) == 0 {
		return nil, fmt.Errorf("listening on neither udp nor socket, please check your configuration")
	}

	// check configuration for custom namespace
	metricPrefix := cf.MetricNamespace
	if metricPrefix != "" && !strings.HasSuffix(metricPrefix, ".") {
		metricPrefix = metricPrefix + "."
	}
	metricPrefixBlacklist := cf.MetricNamespaceBlacklist

	defaultHostname, err := util.GetHostname()
	if err != nil {
		klog.Errorf("statsd: unable to determine default hostname: %s", err.Error())
	}

	histToDist := cf.HistogramCopyToDistribution
	histToDistPrefix := cf.HistogramCopyToDistributionPrefix

	if extraTags == nil {
		extraTags = cf.Tags
	}

	entityIDPrecedenceEnabled := cf.EntityIdPrecedence

	eolTerminationUDP := false
	eolTerminationUDS := false
	eolTerminationNamedPipe := false

	for _, v := range cf.EolRequired {
		switch v {
		case "udp":
			eolTerminationUDP = true
		case "uds":
			eolTerminationUDS = true
		case "named_pipe":
			eolTerminationNamedPipe = true
		default:
			klog.Errorf("Invalid dogstatsd_eol_required value: %s", v)
		}
	}

	s := &Server{
		Started:                   true,
		Statistics:                stats,
		packetsIn:                 packetsChannel,
		sharedPacketPool:          sharedPacketPool,
		sharedFloat64List:         newFloat64ListPool(),
		aggregator:                aggregator,
		listeners:                 tmpListeners,
		stopChan:                  make(chan bool),
		health:                    health.RegisterLiveness("dogstatsd-main"),
		metricPrefix:              metricPrefix,
		metricPrefixBlacklist:     metricPrefixBlacklist,
		defaultHostname:           defaultHostname,
		histToDist:                histToDist,
		histToDistPrefix:          histToDistPrefix,
		extraTags:                 extraTags,
		eolTerminationUDP:         eolTerminationUDP,
		eolTerminationUDS:         eolTerminationUDS,
		eolTerminationNamedPipe:   eolTerminationNamedPipe,
		telemetryEnabled:          telemetry_utils.IsEnabled(),
		entityIDPrecedenceEnabled: entityIDPrecedenceEnabled,
		disableVerboseLogs:        cf.DisableVerboseLogs,
		Debug: &dsdServerDebug{
			Stats: make(map[ckey.ContextKey]metricStat),
			metricsCounts: metricsCountBuckets{
				counts:     [5]uint64{0, 0, 0, 0, 0},
				metricChan: make(chan struct{}),
				closeChan:  make(chan struct{}),
			},
			keyGen: ckey.NewKeyGenerator(),
		},
		UdsListenerRunning: udsListenerRunning,
	}

	// packets forwarding
	// ----------------------

	forwardHost := cf.ForwardHost
	forwardPort := cf.ForwardPort
	if forwardHost != "" && forwardPort != 0 {
		forwardAddress := fmt.Sprintf("%s:%d", forwardHost, forwardPort)
		con, err := net.Dial("udp", forwardAddress)
		if err != nil {
			klog.Warningf("Could not connect to statsd forward host : %s", err)
		} else {
			s.packetsIn = make(chan listeners.Packets, cf.QueueSize)
			go s.forwarder(con, packetsChannel)
		}
	}

	// start the workers processing the packets read on the socket
	// ----------------------

	s.handleMessages()

	// start the debug loop
	// ----------------------

	if metricsStatsEnabled == 1 {
		s.EnableMetricsStats()
	}

	// map some metric name
	// ----------------------

	cacheSize := cf.MapperCacheSize

	mappings := cf.MapperProfiles
	if len(mappings) != 0 {
		mapperInstance, err := mapper.NewMetricMapper(mappings, cacheSize)
		if err != nil {
			klog.Warningf("Could not create metric mapper: %v", err)
		} else {
			s.mapper = mapperInstance
		}
	}
	return s, nil
}

func (s *Server) handleMessages() {
	if s.Statistics != nil {
		go s.Statistics.Process()
		go s.Statistics.Update(&dogstatsdPacketsLastSec)
	}

	for _, l := range s.listeners {
		go l.Listen()
	}

	// Run min(2, GoMaxProcs-2) workers, we dedicate a core to the
	// listener goroutine and another to aggregator + forwarder
	workersCount := runtime.GOMAXPROCS(-1) - 2
	if workersCount < 2 {
		workersCount = 2
	}

	for i := 0; i < workersCount; i++ {
		worker := newWorker(s)
		go worker.run()
		s.workers = append(s.workers, worker)
	}
}

func (s *Server) forwarder(fcon net.Conn, packetsChannel chan listeners.Packets) {
	for {
		select {
		case <-s.stopChan:
			return
		case packets := <-packetsChannel:
			for _, packet := range packets {
				_, err := fcon.Write(packet.Contents)

				if err != nil {
					klog.Warningf("Forwarding packet failed : %s", err)
				}
			}
			s.packetsIn <- packets
		}
	}
}

// Flush flushes all the data to the aggregator to them send it to the Datadog intake.
func (s *Server) Flush() {
	klog.V(5).Infof("Received a Flush trigger")
	// make all workers flush their aggregated data (in the batcher) to the aggregator.
	for _, w := range s.workers {
		w.flush()
	}
	// flush the aggregator to have the serializer/forwarder send data to the backend.
	// We add 10 seconds to the interval to ensure that we're getting the whole sketches bucket
	s.aggregator.Flush(time.Now().Add(time.Second*10), true)
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

// ScanLines is an almost identical reimplementation of bufio.ScanLines, but also
// reports if the returned line is newline-terminated
func ScanLines(data []byte, atEOF bool) (advance int, token []byte, eol bool, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, false, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), true, nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), false, nil
	}
	// Request more data.
	return 0, nil, false, nil
}

func nextMessage(packet *[]byte, eolTermination bool) (message []byte) {
	if len(*packet) == 0 {
		return nil
	}

	advance, message, eol, err := ScanLines(*packet, true)
	if err != nil {
		return nil
	}

	if eolTermination && !eol {
		dogstatsdUnterminatedMetricErrors.Add(1)
		return nil
	}

	*packet = (*packet)[advance:]
	return message
}

func (s *Server) eolEnabled(sourceType listeners.SourceType) bool {
	switch sourceType {
	case listeners.UDS:
		return s.eolTerminationUDS
	case listeners.UDP:
		return s.eolTerminationUDP
	case listeners.NamedPipe:
		return s.eolTerminationNamedPipe
	}
	return false
}

func (s *Server) parsePackets(batcher *batcher, parser *parser, packets []*listeners.Packet, samples []metrics.MetricSample) []metrics.MetricSample {
	for _, packet := range packets {
		klog.V(6).Infof("statsd receive: %q", packet.Contents)
		for {
			message := nextMessage(&packet.Contents, s.eolEnabled(packet.Source))
			if message == nil {
				break
			}
			if len(message) == 0 {
				continue
			}
			if s.Statistics != nil {
				s.Statistics.StatEvent(1)
			}
			messageType := findMessageType(message)

			switch messageType {
			case serviceCheckType:
				serviceCheck, err := s.parseServiceCheckMessage(parser, message, packet.Origin)
				if err != nil {
					s.errLog("statsd: error parsing service check '%q': %s", message, err)
					continue
				}
				batcher.appendServiceCheck(serviceCheck)
			case eventType:
				event, err := s.parseEventMessage(parser, message, packet.Origin)
				if err != nil {
					s.errLog("statsd: error parsing event '%q': %s", message, err)
					continue
				}
				batcher.appendEvent(event)
			case metricSampleType:
				var err error
				samples = samples[0:0]

				samples, err = s.parseMetricMessage(samples, parser, message, packet.Origin)
				if err != nil {
					s.errLog("statsd: error parsing metric message '%q': %s", message, err)
					continue
				}
				for idx := range samples {
					if atomic.LoadUint64(&s.Debug.Enabled) == 1 {
						s.storeMetricStats(samples[idx])
					}
					batcher.appendSample(samples[idx])
					if s.histToDist && samples[idx].Mtype == metrics.HistogramType {
						distSample := samples[idx].Copy()
						distSample.Name = s.histToDistPrefix + distSample.Name
						distSample.Mtype = metrics.DistributionType
						batcher.appendSample(*distSample)
					}
				}
			}
		}
		s.sharedPacketPool.Put(packet)
	}
	batcher.flush()
	return samples
}

func (s *Server) errLog(format string, params ...interface{}) {
	if s.disableVerboseLogs {
		klog.V(5).Infof(format, params...)
	} else {
		klog.Errorf(format, params...)
	}
}

func (s *Server) parseMetricMessage(metricSamples []metrics.MetricSample, parser *parser, message []byte, origin string) ([]metrics.MetricSample, error) {
	sample, err := parser.parseMetricSample(message)
	if err != nil {
		dogstatsdMetricParseErrors.Add(1)
		tlmProcessed.IncWithTags(tlmProcessedErrorTags)
		return metricSamples, err
	}
	if s.mapper != nil {
		mapResult := s.mapper.Map(sample.name)
		if mapResult != nil {
			klog.V(6).Infof("statsd mapper: metric mapped from %q to %q with tags %v", sample.name, mapResult.Name, mapResult.Tags)
			sample.name = mapResult.Name
			sample.tags = append(sample.tags, mapResult.Tags...)
		}
	}
	metricSamples = enrichMetricSample(metricSamples, sample, s.metricPrefix, s.metricPrefixBlacklist, s.defaultHostname, origin, s.entityIDPrecedenceEnabled, s.ServerlessMode)

	if len(sample.values) > 0 {
		s.sharedFloat64List.put(sample.values)
	}

	for idx := range metricSamples {
		// All metricSamples already share the same Tags slice. We can
		// extends the first one and reuse it for the rest.
		if idx == 0 {
			metricSamples[idx].Tags = append(metricSamples[idx].Tags, s.extraTags...)
		} else {
			metricSamples[idx].Tags = metricSamples[0].Tags
		}
		dogstatsdMetricPackets.Add(1)
		tlmProcessed.IncWithTags(tlmProcessedOkTags)
	}
	return metricSamples, nil
}

func (s *Server) parseEventMessage(parser *parser, message []byte, origin string) (*metrics.Event, error) {
	sample, err := parser.parseEvent(message)
	if err != nil {
		dogstatsdEventParseErrors.Add(1)
		tlmProcessed.Inc("events", "error")
		return nil, err
	}
	event := enrichEvent(sample, s.defaultHostname, origin, s.entityIDPrecedenceEnabled)
	event.Tags = append(event.Tags, s.extraTags...)
	tlmProcessed.Inc("events", "ok")
	dogstatsdEventPackets.Add(1)
	return event, nil
}

func (s *Server) parseServiceCheckMessage(parser *parser, message []byte, origin string) (*metrics.ServiceCheck, error) {
	sample, err := parser.parseServiceCheck(message)
	if err != nil {
		dogstatsdServiceCheckParseErrors.Add(1)
		tlmProcessed.Inc("service_checks", "error")
		return nil, err
	}
	serviceCheck := enrichServiceCheck(sample, s.defaultHostname, origin, s.entityIDPrecedenceEnabled)
	serviceCheck.Tags = append(serviceCheck.Tags, s.extraTags...)
	dogstatsdServiceCheckPackets.Add(1)
	tlmProcessed.Inc("service_checks", "ok")
	return serviceCheck, nil
}

// Stop stops a running statsd server
func (s *Server) Stop() {
	close(s.stopChan)
	for _, l := range s.listeners {
		l.Stop()
	}
	if s.Statistics != nil {
		s.Statistics.Stop()
	}
	s.health.Deregister() //nolint:errcheck
	s.Started = false
}

func (s *Server) storeMetricStats(sample metrics.MetricSample) {
	now := time.Now()
	s.Debug.Lock()
	defer s.Debug.Unlock()

	// key
	util.SortUniqInPlace(sample.Tags)
	key := s.Debug.keyGen.Generate(sample.Name, "", sample.Tags)

	// store
	ms := s.Debug.Stats[key]
	ms.Count++
	ms.LastSeen = now
	ms.Name = sample.Name
	ms.Tags = strings.Join(sample.Tags, " ") // we don't want/need to share the underlying array
	s.Debug.Stats[key] = ms

	s.Debug.metricsCounts.metricChan <- struct{}{}
}

// EnableMetricsStats enables the debug mode of the DogStatsD server and start
// the debug mainloop collecting the amount of metrics received.
func (s *Server) EnableMetricsStats() {
	s.Debug.Lock()
	defer s.Debug.Unlock()

	// already enabled?
	if atomic.LoadUint64(&s.Debug.Enabled) == 1 {
		return
	}

	atomic.StoreUint64(&s.Debug.Enabled, 1)
	go func() {
		ticker := time.NewTicker(time.Millisecond * 100)
		var closed bool
		klog.V(5).Info("Starting the statsD debug loop.")
		for {
			select {
			case <-ticker.C:
				sec := time.Now().Truncate(time.Second)
				if sec.After(s.Debug.metricsCounts.currentSec) {
					s.Debug.metricsCounts.currentSec = sec

					if s.hasSpike() {
						klog.Warningf("A burst of metrics has been detected by statsd: here is the last 5 seconds count of metrics: %v", s.Debug.metricsCounts.counts)
					}

					s.Debug.metricsCounts.bucketIdx++
					if s.Debug.metricsCounts.bucketIdx >= len(s.Debug.metricsCounts.counts) {
						s.Debug.metricsCounts.bucketIdx = 0
					}

					s.Debug.metricsCounts.counts[s.Debug.metricsCounts.bucketIdx] = 0
				}
			case <-s.Debug.metricsCounts.metricChan:
				s.Debug.metricsCounts.counts[s.Debug.metricsCounts.bucketIdx]++
			case <-s.Debug.metricsCounts.closeChan:
				closed = true
				break
			}

			if closed {
				break
			}
		}
		klog.V(5).Info("Stopping the statsD debug loop.")
		ticker.Stop()
	}()
}

func (s *Server) hasSpike() bool {
	// compare this one to the sum of all others
	// if the difference is higher than all others sum, consider this
	// as an anomaly.
	var sum uint64
	for _, v := range s.Debug.metricsCounts.counts {
		sum += v
	}
	sum -= s.Debug.metricsCounts.counts[s.Debug.metricsCounts.bucketIdx]
	if s.Debug.metricsCounts.counts[s.Debug.metricsCounts.bucketIdx] > sum {
		return true
	}
	return false
}

// DisableMetricsStats disables the debug mode of the DogStatsD server and
// stops the debug mainloop.
func (s *Server) DisableMetricsStats() {
	s.Debug.Lock()
	defer s.Debug.Unlock()

	if atomic.LoadUint64(&s.Debug.Enabled) == 1 {
		atomic.StoreUint64(&s.Debug.Enabled, 0)
		s.Debug.metricsCounts.closeChan <- struct{}{}
	}

	klog.Info("Disabling StatsD debug metrics stats.")
}

// GetJSONDebugStats returns jsonified debug statistics.
func (s *Server) GetJSONDebugStats() ([]byte, error) {
	s.Debug.Lock()
	defer s.Debug.Unlock()
	return json.Marshal(s.Debug.Stats)
}

// FormatDebugStats returns a printable version of debug stats.
func FormatDebugStats(stats []byte) (string, error) {
	var dogStats map[uint64]metricStat
	if err := json.Unmarshal(stats, &dogStats); err != nil {
		return "", err
	}

	// put metrics in order: first is the more frequent
	order := make([]uint64, len(dogStats))
	i := 0
	for metric := range dogStats {
		order[i] = metric
		i++
	}

	sort.Slice(order, func(i, j int) bool {
		return dogStats[order[i]].Count > dogStats[order[j]].Count
	})

	// write the response
	buf := bytes.NewBuffer(nil)

	header := fmt.Sprintf("%-40s | %-20s | %-10s | %-20s\n", "Metric", "Tags", "Count", "Last Seen")
	buf.Write([]byte(header))
	buf.Write([]byte(strings.Repeat("-", len(header)) + "\n"))

	for _, key := range order {
		stats := dogStats[key]
		buf.Write([]byte(fmt.Sprintf("%-40s | %-20s | %-10d | %-20v\n", stats.Name, stats.Tags, stats.Count, stats.LastSeen)))
	}

	if len(dogStats) == 0 {
		buf.Write([]byte("No metrics processed yet."))
	}

	return buf.String(), nil
}