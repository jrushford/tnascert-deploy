#!/bin/sh
#
# Install the configuration file at /usr/local/etc
# Install tnascert-deploy at /usr/local/bin
CONFIG=/usr/local/etc/tnas-cert.ini
COMMAND=/usr/local/bin/tnascert-deploy

# change the name to my-rest-nas if using TrueNAS core or SCALE 24
if [ -f $CONFIG ] && [ -x $COMMAND ]; then
  $COMMAND -c $CONFIG my-websocket-nas
else
  echo "cannot find a configuration file or the tnascert-deploy command"
fi

