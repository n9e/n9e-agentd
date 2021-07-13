package db

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/n9e/n9e-agentd/pkg/util/expr"
)

var (
	// Simple heuristic to not mistake a source for part of a string (which we also transform it into)
	SOURCE_PATTERN = `(?<!"|\')({})(?!"|\')`
)
var (
	COLUMN_TRANSFORMERS = mapinterface{
		"temporal_percent": get_temporal_percent,
		"monotonic_gauge":  get_monotonic_gauge,
		"tag":              get_tag,
		"tag_list":         get_tag_list,
		"match":            get_match,
		"service_check":    get_service_check,
		//"time_elapsed":     get_time_elapsed,
	}

	EXTRA_TRANSFORMERS = mapinterface{
		"expression": get_expression,
		"percent":    get_percent,
	}
)

func get_tag(transformers mapinterface, column_name, modifiers interface{}) (interface{}, error) {
	// # type: (Dict[str, Callable], str, Any) -> str
	// """
	// Convert a column to a tag that will be used in every subsequent submission.

	// For example, if you named the column `env` and the column returned the value `prod1`, all submissions
	// from that row will be tagged by `env:prod1`.

	// This also accepts an optional modifier called `boolean` that when set to `true` will transform the result
	// to the string `true` or `false`. So for example if you named the column `alive` and the result was the
	// number `0` the tag will be `alive:false`.
	// """
	return func(_ mapinterface, value, _ interface{}) (interface{}, error) {
		return fmt.Sprintf("%v:%v", column_name, String(value)), nil
	}, nil
}

func get_tag_list(transformers mapinterface, column_name, modifiers interface{}) (interface{}, error) {
	// # type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], List[str]]
	// """
	// Convert a column to a list of tags that will be used in every submission.

	// Tag name is determined by `column_name`. The column value represents a list of values. It is expected to be either
	// a list of strings, or a comma-separated string.

	// For example, if the column is named `server_tag` and the column returned the value `'us,primary'`, then all
	// submissions for that row will be tagged by `server_tag:us` and `server_tag:primary`.
	// """

	return func(_ mapinterface, value, _ interface{}) (interface{}, error) {
		var ss []string
		for _, v := range strings.Split(String(value), ",") {
			ss = append(ss, fmt.Sprintf("%v:%v", column_name, v))
		}
		return ss, nil
	}, nil
}

func get_monotonic_gauge(transformers mapinterface, column_name, modifiers interface{}) (interface{}, error) {
	// Send the result as both a `gauge` suffixed by `.total` and a `monotonic_count` suffixed by `.count`.
	gauge, err := _transformer5(transformers, "gauge",
		transformers, fmt.Sprintf("%v.total", column_name), modifiers)
	if err != nil {
		return nil, err
	}

	monotonic_count, err := _transformer5(transformers, "monotonic_count",
		transformers, fmt.Sprintf("%v.count", column_name), modifiers)
	if err != nil {
		return nil, err
	}

	return func(_ mapinterface, value, kwargs interface{}) (interface{}, error) {
		gauge(nil, value, kwargs)
		monotonic_count(nil, value, kwargs)
		return nil, nil
	}, nil
}

func get_temporal_percent(transformers mapinterface, column_name, modifiers_ interface{}) (interface{}, error) {
	// type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], None]
	// """
	// Send the result as percentage of time since the last check run as a `rate`.

	// For example, say the result is a forever increasing counter representing the total time spent pausing for
	// garbage collection since start up. That number by itself is quite useless, but as a percentage of time spent
	// pausing since the previous collection interval it becomes a useful metric.

	// There is one required parameter called `scale` that indicates what unit of time the result should be considered.
	// Valid values are:

	// - `second`
	// - `millisecond`
	// - `microsecond`
	// - `nanosecond`

	// You may also define the unit as an integer number of parts compared to seconds e.g. `millisecond` is
	// equivalent to `1000`.
	// """
	modifiers, err := _mapinterface(modifiers_)
	if err != nil {
		return nil, err
	}

	scale_, err := _string(modifiers.pop("scale"))
	if err != nil {
		return nil, fmt.Errorf("the `scale` parameter is required, err %s", err)
	}

	// try
	scale, err := strconv.Atoi(scale_)
	if err != nil {
		var ok bool
		if scale, ok = TIME_UNITS[scale_]; !ok {
			return nil, fmt.Errorf("the `scale` parameter must be one of: microsecond, millisecond, nanosecond, second")
		}
	}

	rate, err := _transformer5(transformers, "rate", transformers, column_name, modifiers)
	if err != nil {
		return nil, err
	}

	return func(_ mapinterface, value, kwargs interface{}) (interface{}, error) {
		rate(nil, total_time_to_temporal_percent(Float(value), scale), kwargs)
		return nil, nil
	}, nil
}

func get_match(transformers mapinterface, column_name, modifiers interface{}) (interface{}, error) {
	// # type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], None]
	// """
	// This is used for querying unstructured data.

	// For example, say you want to collect the fields named `foo` and `bar`. Typically, they would be stored like:

	// | foo | bar |
	// | --- | --- |
	// | 4   | 2   |

	// and would be queried like:

	// ```sql
	// SELECT foo, bar FROM ...
	// ```

	// Often, you will instead find data stored in the following format:

	// | metric | value |
	// | ------ | ----- |
	// | foo    | 4     |
	// | bar    | 2     |

	// and would be queried like:

	// ```sql
	// SELECT metric, value FROM ...
	// ```

	// In this case, the `metric` column stores the name with which to match on and its `value` is
	// stored in a separate column.

	// The required `items` modifier is a mapping of matched names to column data values. Consider the values
	// to be exactly the same as the entries in the `columns` top level field. You must also define a `source`
	// modifier either for this transformer itself or in the values of `items` (which will take precedence).
	// The source will be treated as the value of the match.

	// Say this is your configuration:

	// ```yaml
	// query: SELECT source1, source2, metric FROM TABLE
	// columns:
	//   - name: value1
	//     type: source
	//   - name: value2
	//     type: source
	//   - name: metric_name
	//     type: match
	//     source: value1
	//     items:
	//       foo:
	//         name: test.foo
	//         type: gauge
	//         source: value2
	//       bar:
	//         name: test.bar
	//         type: monotonic_gauge
	// ```

	// and the result set is:

	// | source1 | source2 | metric |
	// | ------- | ------- | ------ |
	// | 1       | 2       | foo    |
	// | 3       | 4       | baz    |
	// | 5       | 6       | bar    |

	// Here's what would be submitted:

	// - `foo` - `test.foo` as a `gauge` with a value of `2`
	// - `bar` - `test.bar.total` as a `gauge` and `test.bar.count` as a `monotonic_count`, both with a value of `5`
	// - `baz` - nothing since it was not defined as a match item
	// """
	// # Do work in a separate function to avoid having to `del` a bunch of variables
	m, err := _mapinterface(modifiers)
	if err != nil {
		return nil, err
	}

	compiled_items, err := _compile_match_items(transformers, m)
	if err != nil {
		return nil, err
	}

	return func(sources mapinterface, value, kwargs interface{}) (interface{}, error) {
		name, err := _string(value)
		if err != nil {
			return nil, err
		}
		if v, ok := compiled_items[name]; ok {
			v.transformer(sources, sources[v.source], kwargs)
		}
		return nil, nil
	}, nil
}

func get_service_check(transformers mapinterface, column_name, modifiers_ interface{}) (interface{}, error) {
	// # type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], None]
	// """
	// Submit a service check.

	// The required modifier `status_map` is a mapping of values to statuses. Valid statuses include:

	// - `OK`
	// - `WARNING`
	// - `CRITICAL`
	// - `UNKNOWN`

	modifiers, err := _mapinterface(modifiers_)
	if err != nil {
		return nil, err
	}

	// Any encountered values that are not defined will be sent as `UNKNOWN`.
	// Do work in a separate function to avoid having to `del` a bunch of variables
	status_map, err := _compile_service_check_statuses(modifiers)
	if err != nil {
		return nil, err
	}

	service_check_method, err := _transformer5(transformers, "__service_check",
		transformers, column_name, modifiers)
	if err != nil {
		return nil, err
	}

	return func(_ mapinterface, value, kwargs interface{}) (interface{}, error) {
		v, err := _string(value)
		if err != nil {
			return nil, fmt.Errorf("value must be a string, err %s", err)
		}

		status, err := _string(status_map[v])
		if err != nil {
			status = "UNKNOWN"
		}

		return service_check_method(nil, status, kwargs)
	}, nil
}

func get_time_elapsed(transformers mapinterface, column_name, modifiers_ interface{}) (interface{}, error) {

	//Send the number of seconds elapsed from a time in the past as a `gauge`.

	//For example, if the result is an instance of
	//[datetime.datetime](https://docs.python.org/3/library/datetime.html#datetime.datetime) representing 5 seconds ago,
	//then this would submit with a value of `5`.

	//The optional modifier `format` indicates what format the result is in. By default it is `native`, assuming the
	//underlying library provides timestamps as `datetime` objects. If it does not and passes them through directly as
	//strings, you must provide the expected timestamp format using the
	//[supported codes](https://docs.python.org/3/library/datetime.html#strftime-and-strptime-format-codes).

	//!!! note
	//    The code `%z` (lower case) is not supported on Windows.
	// modifiers, err := _mapinterface(modifiers_)
	// if err != nil {
	// 	return nil, err
	// }

	// time_format, err := _string(modifiers.pop("format", "native"))
	// if err != nil {
	// 	return nil, fmt.Errorf("the `format` parameter must be a string")
	// }

	// gauge, err := _transforer5(transformers, "gauge", transformers, column_name, modifiers)

	// if time_format == "native" {
	// 	return func(_ mapinterface, value, kwargs interface{}) (interface{}, error) {
	// 		value = ensure_aware_datetime(value)
	// 		gauge(_, (datetime.now(value.tzinfo) - value).total_seconds(), kwargs)
	// 	}, nil
	// } else {

	// 	return func(_ mapinterface, value, kwargs interface{}) (interface{}, error) {
	// 		value = ensure_aware_datetime(datetime.strptime(value, time_format))
	// 		gauge(_, (datetime.now(value.tzinfo) - value).total_seconds(), kwargs)
	// 	}, nil
	// }
	return nil, nil
}

//
//
func get_expression(transformers mapinterface, name, modifiers_ interface{}) (interface{}, error) {
	//# type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], Any]
	//"""
	//This allows the evaluation of a limited subset of Python syntax and built-in functions.

	//```yaml
	//columns:
	//  - name: disk.total
	//    type: gauge
	//  - name: disk.used
	//    type: gauge
	//extras:
	//  - name: disk.free
	//    expression: disk.total - disk.used
	//    submit_type: gauge
	//```

	//For brevity, if the `expression` attribute exists and `type` does not then it is assumed the type is
	//`expression`. The `submit_type` can be any transformer and any extra options are passed down to it.

	//The result of every expression is stored, so in lieu of a `submit_type` the above example could also be written as:

	//```yaml
	//columns:
	//  - name: disk.total
	//    type: gauge
	//  - name: disk.used
	//    type: gauge
	//extras:
	//  - name: free
	//    expression: disk.total - disk.used
	//  - name: disk.free
	//    type: gauge
	//    source: free
	//```

	//The order matters though, so for example the following will fail:

	//```yaml
	//columns:
	//  - name: disk.total
	//    type: gauge
	//  - name: disk.used
	//    type: gauge
	//extras:
	//  - name: disk.free
	//    type: gauge
	//    source: free
	//  - name: free
	//    expression: disk.total - disk.used
	//```

	//since the source `free` does not yet exist.
	//"""
	modifiers, err := _mapinterface(modifiers_)
	if err != nil {
		return nil, err
	}

	//available_sources, err := _mapinterface(modifiers.pop("sources"))
	//if err != nil {
	//	return nil, err
	//}

	expression := String(modifiers.pop("expression"))
	if expression == "" {
		return nil, fmt.Errorf("the `expression` parameter is required")
	}

	notations, err := expr.NewNotations([]byte(expression))
	if err != nil {
		return nil, fmt.Errorf("parse expr %s err %s", expression, err)
	}

	if _, ok := modifiers["submit_type"]; !ok {
		return func(sources mapinterface, kwargs, _ interface{}) (interface{}, error) {
			return notations.Calc(sources.getFloat)
		}, nil
	}

	submit_method, err := _transformer5(transformers,
		String(modifiers["submit_type"]), transformers, name, modifiers)
	if err != nil {
		return nil, err
	}
	return func(sources mapinterface, kwargs, _ interface{}) (interface{}, error) {
		result, err := notations.Calc(sources.getFloat)
		if err != nil {
			return nil, err
		}
		submit_method(sources, result, kwargs)
		return result, nil
	}, nil
}

func get_percent(transformers mapinterface, name, modifiers_ interface{}) (interface{}, error) {
	//# type: (Dict[str, Callable], str, Any) -> Callable[[Any, Any, Any], None]
	//"""
	//Send a percentage based on 2 sources as a `gauge`.

	//The required modifiers are `part` and `total`.

	//For example, if you have this configuration:

	//```yaml
	//columns:
	//  - name: disk.total
	//    type: gauge
	//  - name: disk.used
	//    type: gauge
	//extras:
	//  - name: disk.utilized
	//    type: percent
	//    part: disk.used
	//    total: disk.total
	//```

	//then the extra metric `disk.utilized` would be sent as a `gauge` calculated as `disk.used / disk.total * 100`.

	//If the source of `total` is `0`, then the submitted value will always be sent as `0` too.
	//"""
	modifiers, err := _mapinterface(modifiers_)
	if err != nil {
		return nil, err
	}

	available_sources, err := _mapinterface(modifiers.pop("sources"))
	if err != nil {
		return nil, err
	}

	part := String(modifiers.pop("part"))
	if part == "" {
		return nil, fmt.Errorf("the `part` parameter is required")
	}

	if _, ok := available_sources[part]; !ok {
		return nil, fmt.Errorf("the `part` parameter `%s` is not an available source", part)
	}

	total := String(modifiers.pop("total"))
	if total == "" {
		return nil, fmt.Errorf("the `total` parameter is required")
	}
	if _, ok := available_sources[total]; !ok {
		return nil, fmt.Errorf("the `total` parameter `%s` is not an available source", total)
	}

	gauge, err := _transformer5(transformers, "gauge", transformers, name, modifiers)
	if err != nil {
		return nil, err
	}
	gauge, err = create_extra_transformer(gauge)
	if err != nil {
		return nil, err
	}

	return func(sources mapinterface, kwargs, _ interface{}) (interface{}, error) {
		gauge(sources, compute_percent(Float(sources[part]), Float(sources[total])), kwargs)
		return nil, nil
	}, nil
}

func compute_percent(part, total float64) float64 {
	if total > 0 {
		return part / total * 100
	}

	return 0
}

func _compile_service_check_statuses(modifiers mapinterface) (mapinterface, error) {
	// # type: (Dict[str, Any]) -> Dict[str, ServiceCheckStatus]
	status_map_ := modifiers.pop("status_map")
	if status_map_ == nil {
		return nil, fmt.Errorf("the `status_map` parameter is required")
	}
	status_map, ok := status_map_.(mapinterface)
	if !ok {
		return nil, fmt.Errorf("the `status_map` parameter must be a mapping")
	}
	if len(status_map) == 0 {
		return nil, fmt.Errorf("the `status_map` parameter must not be empty")
	}

	for value, status_string_ := range status_map {
		status_string, ok := status_string_.(string)
		if !ok {
			return nil, fmt.Errorf("status `%v` for value `%v` of parameter `status_map` is not a string", status_string_, value)
		}

		switch s := strings.ToUpper(status_string); s {
		case "OK", "WARNING", "CRITICAL", "UNKNOWN":
			status_map[value] = s
		default:
			return nil, fmt.Errorf("invalid status `{}` for value `{}` of parameter `status_map`", status_string, value)
		}
	}
	return status_map, nil
}

func _compile_match_items(transformers, modifiers mapinterface) (map[string]sourceTransform, error) {
	//# type: (Dict[str, Any], Dict[str, Any]) -> Dict[str, Tuple[str, Any]]
	items_ := modifiers.pop("items")
	if items_ == nil {
		return nil, fmt.Errorf("the `items` parameter is required")
	}

	items, ok := items_.(mapinterface)
	if !ok {
		return nil, fmt.Errorf("the `items` parameter must be a mapping")
	}

	global_transform_source := modifiers.pop("source")

	compiled_items := map[string]sourceTransform{}
	for item, data_ := range items {
		data, ok := data_.(mapinterface)
		if !ok {
			return nil, fmt.Errorf("item `%s` is not a mapping", item)
		}

		transform_name_ := data.pop("name")
		if transform_name_ == nil {
			return nil, fmt.Errorf("the `name` parameter for item `%s` is required", item)
		}
		transform_name, ok := transform_name_.(string)
		if !ok {
			return nil, fmt.Errorf("the `name` parameter for item `%s` must be a string", item)
		}

		transform_type_ := data.pop("type")
		if transform_type_ == nil {
			return nil, fmt.Errorf("the `type` parameter for item `%s` is required", item)
		}
		transform_type, ok := transform_type_.(string)
		if !ok {
			return nil, fmt.Errorf("the `type` parameter for item `%s` must be a string", item)
		}

		if _, ok := transformers[transform_type]; !ok {
			return nil, fmt.Errorf("unknown type `%s` for item `%s`", transform_type, item)
		}

		transform_source_ := data.pop("source", global_transform_source)
		if transform_source_ == nil {
			return nil, fmt.Errorf("the `source` parameter for item `%s` is required", item)
		}

		transform_source, ok := transform_source_.(string)
		if !ok {
			return nil, fmt.Errorf("the `source` parameter for item `%` must be a string", item)
		}

		transform_modifiers := make(mapinterface)

		transformer, err := _transformer5(transformers, transform_name, transformers, transform_name, transform_modifiers)
		if err != nil {
			return nil, fmt.Errorf("unknown type `%s` for item `%s` err %s", transform_type, item, err)
		}

		transform_modifiers.update(modifiers)
		transform_modifiers.update(data)
		compiled_items[item] = sourceTransform{
			source:      transform_source,
			transformer: transformer,
		}
	}

	return compiled_items, nil
}
