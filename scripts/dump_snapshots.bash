#!/bin/bash -x

DUMP=/home/scrapemonster/scrapemonster/bin/dumpSnapshots
DUMP_DIR=/home/coupang/snapshots
LOG_FILE=/var/log/scrapemonster/dumpSnapshots.log

YESTERDAY=`perl -e '@t=localtime(time()-86400); printf("%d-%02d-%02d\n",$t[5]+1900,$t[4]+1,$t[3]);'`
DAY=${1:-${YESTERDAY}}

${DUMP} -day=$DAY -dir=${DUMP_DIR} -compress >>${LOG_FILE} 2>&1
