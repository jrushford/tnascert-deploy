
#### Sample Deployment scripts

The following is a sample **tnas-cert.ini** configuration file showing
two TrueNAS entries.  One entry is for the newer TrueNAS-SCALE version 25+
using the websocket API, wsapi.  The 2nd entry is a sample for the older
TrueNAS core or TrueNAS-SCALE version 24 using the restapi.

Choose the appropriate configuration for your TrueNAS version and then
edit and add an api_key, connect_host, private_key_path, and 
full_chain_path to suit your environment. See the Documentation at the
main repository page for configuration file details.

The sample **deploy-hook.sh** script assumes that you install the 
configuration file to /usr/local/etc/tnas-cert.ini and that the 
**tnascert-deploy** executable binary is installed at /usr/local/bin.
Modify the script for your needs.  The script is pretty basic but should
work fine.

Example acme.sh usage with the deploy-hook.sh:

  acme.sh --install-cert -d mydomain.org --deploy-hook deploy-hook.sh

