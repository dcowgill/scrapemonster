#!/bin/bash

MYSQLDUMP=/usr/bin/mysqldump
BACKUPS_DIR=/data/backups/mysql
NOW=`/bin/date "+%F_%H-%M-%S"`

function backupdb {
	db=$1
	$MYSQLDUMP $db | /bin/bzip2 > ${BACKUPS_DIR}/${db}_${NOW}.sql.bz2
}

backupdb coupang
backupdb scrapemonster
