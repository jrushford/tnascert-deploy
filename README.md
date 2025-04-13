
#### NAME

tnascert-deploy - A tool used to deploy UI certificates to a TrueNAS host

#### SYNOPSIS

tnascert-deploy [-h] [-c value] section_name<br> 

 -c, --config="full path to tnas-cert.ini file"<br>
 -h, --help<br>

#### DESCRIPTION

A tool used to import a TLS certificate and private key into a TrueNAS
SCALE host.  Once imported, the tool may be configred to activate the
TrueNAS host to use it as the main UI TLS certificate.  

The <b>tnas-cert.ini</b> file consists of multiple <b>sections</b> 
The optional command line argument <b>section_name</b> may by
used to load that particular configuration.  This allows for maintaining 
multiple configurations in one tnas-cert.ini file where
each ***section_name*** may be an individual ***TrueNAS*** host.

If the optioanl argument ***section_name*** is not provided, The
***default*** section name is chosen to load the configuration if
the ***default*** section name exists in the configuration file,
otherwise a configuration error is printed and the program exits.

See the sample **tnas-cert.ini** file.

#### FILES

tnas-cert.ini<br>

#### BUILD NOTES

This tool uses the TrueNAS Scale JSON-RPC 2.0 API and the TrueNAS client API module.
To build, fetch the [truenas_api/truenas_api module at github](https://github.com/truenas/api_client_golang.git) 
and install it in your go build path.<br>

Then use ***go build***

#### Contact
John J. Rushford<br>
jrushford@apache.org
