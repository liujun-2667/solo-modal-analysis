#!/bin/sh

/usr/local/bin/modal-analysis &
nginx -g "daemon off;"
