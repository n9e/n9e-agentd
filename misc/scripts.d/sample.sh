#!/bin/bash
echo '[
{"metric":"test1", "value":1, "type":"GUAGE", "tags":{"a":"1"}},
{"metric":"test2", "value":2, "type":"COUNTER", "tags":{"a":"2"}},
{"metric":"test3", "value":3, "type":"MONOTONIC_COUNT", "tags":{"a":"3"}},
{"metric":"test4", "value":4, "tags":{"a":"4"}}
]'
