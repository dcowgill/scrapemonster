#!/bin/bash -x

CRAWL=/home/scrapemonster/scrapemonster/bin/crawl
LOG_DIR=/var/log/scrapemonster
SITES="tmon wmp"

for site in $SITES; do
    log_file="${LOG_DIR}/crawl_${site}.log"
    ${CRAWL} -s=${site} -db -q -v >>${log_file} 2>&1
done
