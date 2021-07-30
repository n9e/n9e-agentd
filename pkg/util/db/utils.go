package db

import (
	"database/sql"
	"fmt"
	"net"
	"reflect"
	"strconv"
	"time"

	"github.com/n9e/n9e-agentd/pkg/config"
	"github.com/n9e/n9e-agentd/pkg/metrics"
	"k8s.io/klog/v2"
)

var (
	// AgentCheck methods to transformer name e.g. set_metadata -> metadata
	SUBMISSION_METHODS = map[string]string{
		"Gauge":          "gauge",
		"Count":          "count",
		"MonotonicCount": "monotonic_count",
		"Rate":           "rate",
		"Histogram":      "histogram",
		"Historate":      "historate",
		//"SetMetadata":    "metadata",
		// These submission methods require more configuration than just a name
		// and a value and therefore must be defined as a custom transformer.
		"ServiceCheck": "__service_check",
	}
)

func create_submission_transformer(submit_method interface{}) transformHandle {

	// During the compilation phase every transformer will have access to all the others and may be
	// passed the first arguments (e.g. name) that will be forwarded the actual AgentCheck methods.
	switch fn := submit_method.(type) {
	case func(string, float64, string, []string):
		return func(_ mapinterface, creation_args, modifiers_ interface{}) (interface{}, error) {
			modifiers, err := _mapinterface(modifiers_)
			if err != nil {
				return nil, err
			}

			metric_name, ok := creation_args.(string)
			if !ok {
				panic(fmt.Sprintf("metric_name must be a string %#v", creation_args))
			}

			return func(_ mapinterface, call_args, kwargs_ interface{}) (interface{}, error) {
				kwargs, err := _mapinterface(kwargs_)
				if err != nil {
					return nil, err
				}

				kwargs.update(modifiers)

				fn(metric_name, Float(call_args), "", kwargs.tags())
				return nil, nil
			}, nil
		}
	case func(string, metrics.ServiceCheckStatus, string, []string, string):
		return func(_ mapinterface, creation_args, modifiers_ interface{}) (interface{}, error) {
			modifiers, err := _mapinterface(modifiers_)
			if err != nil {
				return nil, err
			}

			check_name := String(creation_args)

			return func(_ mapinterface, call_args, kwargs_ interface{}) (interface{}, error) {
				kwargs, err := _mapinterface(kwargs_)
				if err != nil {
					return nil, err
				}

				kwargs.update(modifiers)

				status_string, ok := call_args.(string)
				if !ok {
					panic(fmt.Sprintf("metric_value must be a metrics.ServiceCheckStatus %#v", call_args))
				}

				status, ok := serviceCheckStatus[status_string]
				if !ok {
					status = metrics.ServiceCheckUnknown
				}

				fn(check_name, status, "", kwargs.tags(), "")
				return nil, nil
			}, nil
		}
	default:
		panic(fmt.Sprintf("unsupported sumbit_method type %s", reflect.TypeOf(submit_method)))
	}
}

func create_extra_transformer(column_transformer transformHandle, source ...string) (transformHandle, error) {
	// Every column transformer expects a value to be given but in the post-processing
	// phase the values are determined by references, so to avoid redefining every
	// transformer we just map the proper source to the value.
	if len(source) > 0 && source[0] != "" {
		return func(sources mapinterface, kwargs, _ interface{}) (interface{}, error) {
			return column_transformer(sources, sources[source[0]], kwargs)
		}, nil
		// Extra transformers that call regular transformers will want to pass values directly.
	} else {
		return column_transformer, nil
	}
}

type ConstantRateLimiter struct {
	rate_limit_s int
	period_s     time.Duration
	last_event   time.Time
}

// Basic rate limiter that sleeps long enough to ensure the rate limit is not exceeded. Not thread safe.
// :param rate_limit_s: rate limit in seconds
func NewConstantRateLimiter(rate_limit_s int) *ConstantRateLimiter {
	p := &ConstantRateLimiter{rate_limit_s: rate_limit_s}

	if rate_limit_s > 0 {
		p.period_s = time.Duration(1/rate_limit_s) * time.Second
	}

	return p
}

// Sleeps long enough to enforce the rate limit
func (p *ConstantRateLimiter) sleep() {
	now := time.Now()
	if t := p.period_s - (now.Sub(p.last_event)); t > 0 {
		time.Sleep(t)
	}
	p.last_event = now
}

func resolve_db_host(db_host string) string {
	agent_hostname := config.C.Hostname
	if len(db_host) == 0 || db_host == "localhost" || db_host == "127.0.0.1" {
		return agent_hostname
	}

	host_ip, err := net.LookupIP(db_host)
	if err != nil {
		klog.V(6).Infof("failed to resolve DB host '%s' due to %r. falling back to agent hostname: %s", db_host, err, agent_hostname)
		return agent_hostname
	}

	agent_host_ip, err := net.LookupIP(agent_hostname)
	if err != nil {
		klog.V(6).Infof("failed to resolve agent host '%s' due to socket.gaierror(%s). using DB host: %s", agent_hostname, err, db_host)
	}

	if len(agent_host_ip) > 0 && len(host_ip) > 0 && agent_host_ip[0].String() == host_ip[0].String() {
		return agent_hostname
	}

	return db_host
}

func total_time_to_temporal_percent(total_time float64, scales ...int) float64 {
	scale := MILLISECOND
	if len(scales) > 0 {
		scale = scales[0]
	}
	// This is really confusing, sorry.
	//
	// We get the `total_time` in `scale` since the start and we want to compute a percentage.
	// Since the time is monotonically increasing we can't just submit a point-in-time value but
	// rather it needs to be temporally aware, thus we submit the value as a rate.
	//
	// If we submit it as-is, that would be `scale` per second but we need seconds per second
	// since the Agent's check run interval is internally represented as seconds. Hence we divide
	// by 1000, for example, if the `scale` is milliseconds.
	//
	// At this point we have a number that will be no greater than 1 when compared to the last run.
	//
	// To turn it into a percentage we multiply by 100.
	//
	// Example:
	//
	// Say we have 2 moments in time T, tracking a monotonically increasing value X in milliseconds,
	// and the difference between each T is the default check run interval (15s).
	//
	// T1 = 100, X1 = 2,000 / 1,000 * 100 = 200
	// T2 = 115, X2 = 5,000 / 1,000 * 100 = 500
	//
	// See: https://github.com/DataDog/datadog-agent/blob/7.25.x/pkg/metrics/rate.go#L37
	//
	// V = (X2 - X1) / (T2 - T1) = (500 - 200) / (115 - 100) = 20%
	//
	// which is correct because 3000 ms = 3s and 3s of 15s is 20%
	return total_time / float64(scale) * 100
}

func Int(a interface{}) int64 {
	switch v := a.(type) {
	case int64:
		return v
	case *[]byte:
		i, _ := strconv.ParseInt(string(*v), 10, 0)
		return i
	case *string:
		i, _ := strconv.ParseInt(*v, 10, 0)
		return i
	case string:
		i, _ := strconv.ParseInt(v, 10, 0)
		return i
	case *float64:
		return int64(*v)
	case float64:
		return int64(v)
	case *int64:
		return *v
	default:
		panic(fmt.Sprintf("unsupported type %s", reflect.TypeOf(a)))
	}
}

func Float(a interface{}) float64 {
	switch v := a.(type) {
	case *sql.RawBytes:
		i, _ := strconv.ParseFloat(string(*v), 0)
		return i
	case *string:
		f, _ := strconv.ParseFloat(*v, 0)
		return f
	case string:
		f, _ := strconv.ParseFloat(v, 0)
		return f
	case *float64:
		return *v
	case float64:
		return v
	default:
		panic(fmt.Sprintf("unsupported type %s", reflect.TypeOf(a)))
	}
}

func String(a interface{}) string {
	switch v := a.(type) {
	case *sql.RawBytes:
		return string(*v)
	case string:
		return v
	default:
		panic(fmt.Sprintf("unsupported type %s", reflect.TypeOf(a)))
	}
}

func _result(in interface{}) (interface{}, error) {
	if in == nil {
		return nil, fmt.Errorf("get empty")
	}
	if e, ok := in.(error); ok {
		return nil, e
	}

	return in, nil
}

func _transformer(in interface{}) (transformHandle, error) {
	t, ok := in.(transformHandle)
	if !ok {
		t, ok = in.(func(mapinterface, interface{}, interface{}) (interface{}, error))
	}

	if !ok {
		return nil, fmt.Errorf("typeof %s is not a transformer", reflect.TypeOf(in))
	}
	return t, nil
}

func _transformer4(in interface{}, arg1 mapinterface, arg2, arg3 interface{}) (transformHandle, error) {
	t, err := _transformer(in)
	if err != nil {
		return nil, err
	}

	t2, err := t(arg1, arg2, arg3)
	if err != nil {
		return nil, err
	}

	return _transformer(t2)
}

func _transformer5(transformers mapinterface, key string, arg1 mapinterface, arg2, arg3 interface{}) (transformHandle, error) {
	return _transformer4(transformers[key], arg1, arg2, arg3)
}

func _mapinterface(in interface{}) (mapinterface, error) {
	v, ok := in.(mapinterface)
	if !ok {
		return nil, fmt.Errorf("typeof %s is not a map[string]interface{}", reflect.TypeOf(in))
	}

	return v, nil
}

func _string(in interface{}) (string, error) {
	v, ok := in.(string)
	if !ok {
		return "", fmt.Errorf("typeof %s is not a string", reflect.TypeOf(in))
	}

	return v, nil
}

func _float(in interface{}) (float64, error) {
	v, ok := in.(float64)
	if !ok {
		return 0, fmt.Errorf("typeof %s is not a string", reflect.TypeOf(in))
	}

	return v, nil
}
