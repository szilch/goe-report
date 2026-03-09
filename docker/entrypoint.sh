#!/bin/sh
set -e

# Use defaults if not provided
CRON_EXPRESSION=${CRON_EXPRESSION:-"0 0 1 * *"}
CRON_COMMAND=${CRON_COMMAND:-"/app/goe-report status"}

# The busybox cron daemon doesn't automatically load environment variables 
# from the Docker environment. We need to save them so they can be sourced by the cron job.
# We extract all GOEREPORT_ variables (the Viper prefix for this app).
env | grep '^GOEREPORT_' > /app/env.sh
echo "export PATH=\$PATH:/app" >> /app/env.sh

# Create the cron job. We add a prefix to source our saved environment variables.
# Output is redirected to /proc/1/fd/1 so that it appears in `docker logs`.
echo "${CRON_EXPRESSION} . /app/env.sh && ${CRON_COMMAND} > /proc/1/fd/1 2>&1" > /etc/crontabs/root

echo "======================================"
echo "Starting goe-report cron container..."
echo "Cron expression : $CRON_EXPRESSION"
echo "Command to run  : $CRON_COMMAND"
echo "Env vars loaded : $(cat /app/env.sh | grep -c '^GOEREPORT_') config variables"
echo "======================================"
echo ""

# Run busybox cron in the foreground (-f)
exec crond -f -l 2
