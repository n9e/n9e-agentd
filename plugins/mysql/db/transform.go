package db

import (
	"fmt"
	"strconv"

	"k8s.io/klog/v2"
)

var (
	COLUMNTRANSFORMERS = map[string]interface{}{
		"temporalPercent": getTemporalPercent,
		"monotonicGauge":  getMonotonicGauge,
		"tag":             getTag,
		"tagList":         getTagList,
		"match":           getMatch,
		"serviceCheck":    getServiceCheck,
		"timeElapsed":     getTimeElapsed,
	}
	EXTRATRANSFORMERS = map[string]interface{}{
		"expression": getExpression,
		"percent":    getPercent,
	}

	SECOND      = 1
	MILLISECOND = 1000
	MICROSECOND = 1000000
	NANOSECOND  = 1000000000

	TIME_UNITS = map[string]int{
		"microsecond": MICROSECOND,
		"millisecond": MILLISECOND,
		"nanosecond":  NANOSECOND,
		"second":      SECOND,
	}
)

func getMonotonicGauge() error { return nil }
func getTag() error            { return nil }
func getTagList() error        { return nil }
func getMatch() error          { return nil }
func getServiceCheck() error   { return nil }
func getTimeElapsed() error    { return nil }
func getExpression() error     { return nil }
func getPercent() error        { return nil }

// # type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], None]
// """
// Send the result as percentage of time since the last check run as a `rate`.
//
// For example, say the result is a forever increasing counter representing the total time spent pausing for
// garbage collection since start up. That number by itself is quite useless, but as a percentage of time spent
// pausing since the previous collection interval it becomes a useful metric.
//
// There is one required parameter called `scale` that indicates what unit of time the result should be considered.
// Valid values are:
//
// - `second`
// - `millisecond`
// - `microsecond`
// - `nanosecond`
//
// You may also define the unit as an integer number of parts compared to seconds e.g. `millisecond` is
// equivalent to `1000`.
// """
func getTemporalPercent(transformers, columnName string, modifiers map[string]string) error {
	var scale int
	if s := modifiers["scale"]; s == "" {
		return fmt.Errorf("the `scale` parameter is required")
	} else if i, ok := TIME_UNITS[s]; ok {
		scale = i
	} else if i, err := strconv.Atoi(s); err == nil {
		scale = i
	} else {
		return fmt.Errorf("the `scale` parameter must be an integer representing parts of a second e.g. 1000 for millisecond")
	}

	klog.V(5).Info(scale)

	return nil

	/*
		    rate = transformers["rate"](transformers, column_name, **modifiers)
			return func(_ interface, value, map[string]interface{}) {
				rate()
			}
	*/

}
