
#### NAME

tnascert-deploy - A tool used to deploy UI certificates to a TrueNAS host

#### SYNOPSIS

tnascert-deploy [-h] [-c value] section_name ... section_name<br> 

 -c, --config="full path to tnas-cert.ini file"<br>
 -h, --help<br>
 -v, --version<br>

#### DESCRIPTION

A tool used to import a TLS certificate and private key into a TrueNAS
SCALE host running ***TrueNAS 24.14, TrueNAS 25+, TrueNAS-CORE, or zVault***
Once imported, the tool may be configred to activate the TrueNAS host to use
it as the main UI TLS certificate.  

The <b>tnas-cert.ini</b> file consists of multiple <b>sections</b> 
The optional command line arguments <b>section_name</b> may by
used to load that particular configuration.  This allows for maintaining 
multiple configurations in one tnas-cert.ini file where
each ***section_name*** may be an individual ***TrueNAS*** host.
You may list multiple ***_section_name*** on the command line to loop
through certificate installation on multiple ***TrueNAS*** hosts.

If the optional argument ***section_name*** is not provided, The
***deploy_default*** section name is chosen to load the configuration if
it exists.

For client authentication, you may use either a TrueNAS ***api_key*** or 
the ***username*** and ***password*** for user with admin privileges. The
***api_key*** is the preferred method for authentication.  If all three
variables are set in a configuration, the ***api_key*** login method will
always be used.  If you prefer to use the ***username*** and ***password***
ensure only the ***username*** and ***password*** are set in your configuration.

See the sample **tnas-cert.ini** file.

#### FILES

The default configuration file is named ***tnas-cert.ini*** in the current working
directory.  You may use the command line switch to use another file name and full
path to the config file.

#### CONFIG FILE SETTINGS

In order to authenticate with a TrueNAS system, the user must either use the
TrueNAS UI to generate and copy an **api_key** or use an admin **username**
and **password** in the configuration file.  The **api_key** is preferred and
if all three are defined in the configuration file, only the **api_key** will
be used.  Do not include the **api_key** if you wish to use the **username**
and **password**.

 - **api_key**                - (optional, no default) TrueNAS 64 byte API Key for login the 
                              preferred login method).
 - **username**               - (optional, no default) TrueNAS username with admin privileges 
                              (API key is preferred for login)
 - **password**               - (optional, no default) TrueNAS password for user with admin 
                              privileges, (API key is preferred for login)
 - **cert_basename**          - (optional, default is **"tnascert-deploy"**) basename 
                              for the certificate naming in TrueNAS.
 - **connect_host**           - (required), TrueNAS DNS Fully Qualified Domain Name, FQDN, or 
                              IP address
 - **client_api**             - (optional, default is "wsapi") The TrueNAS API to use. Choices
                              are: 'wsapi' for the JSON-RPC 2.0 websocket API or 'restapi' for
                              the RESTful v2.0 API.
 - **delete_old_certs**       - (optional, default is **false**) whether to remove old 
                              certificates, default is false
 - **full_chain_path**        - (required), full path name to the certificate full_chain.pem
 - **private_key_path**       - (required), full path name to the certificate private_key.pem
 - **port**                   - (optional, default is **443**) TrueNAS API endpoint port
 - **protocol**               - (optional, default is **"wss"**) websocket protocol 'ws', 'wss', 
                              'http', or 'https'.  'ws' and 'wss'are only for TrueNAS-SCALE
                              systems utilizing the JSON-RPC 2.0 websocket API.  Use 'http' or
                              'https' for systems utilizing the RESTful v2.0 API.
 - **tls_skip_verify**        - (optional, default is **false**) strict SSL cert verification of
							   the endpoint.
 - **add_as_ui_certificate**  - (optional, default is **false**) install as the active UI
                              certificate if true
 - **add_as_ftp_certificate** - (optional, default is **false**) install as the active FTP
                              service certificate if true
 - **add_as_app_certificate** - (optional, default is **false**) install as the active APP
                              service certificate if true to the apps listed in the 'app_list'
 - **app_list**               - (optional, no default) A comma separated list of docker apps
                              that you wish to have the newly imported certificate used.
                              Apps in the list are only set to used the certificate if they have
                              one assigned already. You must enable 'add_as_app_certificate' to
                              process the list.
 - **timeoutSeconds**         - (optional, default is **10**) the number of seconds after which
							   the truenas client calls fail
 - **debug**                  - (oprional, default is **false**) debug logging if true

#### NOTES

This tool uses the TrueNAS Scale JSON-RPC 2.0 API and the TrueNAS client API module. Supports versions of ***TrueNAS 25.04*** or later

See Also:  
+ [TrueNAS api_client_golang](https://github.com/truenas/api_client_golang)
+ [TrueNAS websocket API documentaion](https://www.truenas.com/docs/api/scale_websocket_api.html)


Clone this repository and build the tool using ***go build***

#### Contact
John J. Rushford<br>
jrushford@apache.org
