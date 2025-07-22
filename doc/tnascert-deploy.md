
#### NAME

tnascert-deploy - A tool used to deploy UI certificates to a TrueNAS host

#### SYNOPSIS

tnascert-deploy [-h] [-c value] section_name ... section_name<br> 

 -c, --config="full path to tnas-cert.ini file"<br>
 -h, --help<br>
 -v, --version<br>

#### DESCRIPTION

A tool used to import a TLS certificate and private key into a TrueNAS
SCALE host running ***TrueNAS 25.04*** or later.  Once imported, the tool 
may be configred to activate the TrueNAS host to use it as the main UI 
TLS certificate.  

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

    + api_key                string  - TrueNAS 64 byte API Key for login (preferred login method).
    + username               string  - TrueNAS username with admin privileges (API key is preferred for login)
    + password               string  - TrueNAS password for user with admin privileges, (API key is preferred for login)
    + cert_basename          string  - basename for cert naming in TrueNAS
    + connect_host           string  - TrueNAS DNS Fully Qualified Domain Name, FQDN, or IP address
    + delete_old_certs       bool    - whether to remove old certificates, default is false
    + full_chain_path        string  - path to full_chain.pem
    + port                   uint64  - TrueNAS API endpoint port
    + protocol               string  - websocket protocol 'ws' or 'wss' wss' is default
    + private_key_path       string  - path to private_key.pem
    + tls_skip_verify        bool    - strict SSL cert verification of the endpoint, false by default
    + add_as_ui_certificate  bool    - install as the active UI certificate if true
    + add_as_ftp_certificate bool    - install as the active FTP service certificate if true
    + add_as_app_certificate bool    - install as the active APP service certificate if true
    + timeoutSeconds         int64   - the number of seconds after which the truenas client calls fail
    + debug                  bool    - debug logging if true

#### NOTES

This tool uses the TrueNAS Scale JSON-RPC 2.0 API and the TrueNAS client API module. Supports versions of ***TrueNAS 25.04*** or later

See Also:  
+ [TrueNAS api_client_golang](https://github.com/truenas/api_client_golang)
+ [TrueNAS websocket API documentaion](https://www.truenas.com/docs/api/scale_websocket_api.html)


Clone this repository and build the tool using ***go build***

#### Contact
John J. Rushford<br>
jrushford@apache.org
