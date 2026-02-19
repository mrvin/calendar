#!/bin/sh

set -e

if [ "$ENV" = 'DEV' ]; then
	echo "Running Development Server"
	exec /app/bin/calendar
elif [ "$ENV" = 'TEST' ]; then
	echo "Running Tests"
	cd /app/
	make test
else
	echo "Invalid 'ENV': '$ENV'"
fi
