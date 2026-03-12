#!/bin/sh
set -e

# Use defaults if not provided
CRON_EXPRESSION=${CRON_EXPRESSION:-"0 0 1 * *"}
CRON_COMMAND=${CRON_COMMAND:-"/app/echarge-report status"}

# Supercronic does not require saving environment variables to a file since it natively
# runs within this process tree and passes environment variables downwards.
# We create a simple crontab file for supercronic.
echo "${CRON_EXPRESSION} ${CRON_COMMAND}" > /app/crontab

echo "======================================"
echo "Starting echarge-report supercronic container as non-root..."
echo "User ID         : $(id -u)"
echo "Group ID        : $(id -g)"
echo "Cron expression : $CRON_EXPRESSION"
echo "Command to run  : $CRON_COMMAND"
echo "Env vars loaded : $(env | grep -c '^GOEREPORT_') config variables"
echo "======================================"
echo ""

# Run supercronic in the foreground
exec supercronic /app/crontab
