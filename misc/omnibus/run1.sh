#!/bin/bash
unset PY_RUNTIMES
export PACKAGE_VERSION="5.0.0"
export RELEASE="alpha1"
bundle exec omnibus build agentd1 --log-level debug
