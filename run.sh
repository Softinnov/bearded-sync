#!/bin/sh

docker run --rm -d --name bearded-sync -v /tmp/tickets:/tickets --link esc-db:db test/bearded-sync
