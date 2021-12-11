.PHONY: build

build:
	bundle install && \
	bundle exec omnibus build agentd

mock:
	bundle install && \
	MAKE_MOCK=true bundle exec omnibus build test

