#!/bin/sh

usage() {
    >&2 echo "usage: $0 [-a] [-e] [-d description] league start_year"
}

PSQL="/usr/local/pgsql/bin/psql"

active=false
exhibition=false
description="Season"
while getopts aed: flag; do
    case "$flag" in
        a) active=true;;
        e) exhibition=true;;
        d) description="$OPTARG";;
        ?) exit 1;;
    esac
done
shift "$(($OPTIND-1))"

if [ $# -ne 2 ]; then
    usage
    exit 1
fi

if [ -z "$RECAP_DB" ]; then
    >&2 echo "RECAP_DB not set"
    exit 1
fi

league=$(echo $1 | tr '[:lower:]' '[:upper:]')
start_year=$2

sql="select new_season('${league}', ${start_year}, '${description}', ${exhibition}, ${active});"
$PSQL $RECAP_DB -q -c "$sql" > /dev/null
