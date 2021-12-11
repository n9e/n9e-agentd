#!/bin/bash
export PY_RUNTIMES=3
bundle exec omnibus build agentd --log-level debug
