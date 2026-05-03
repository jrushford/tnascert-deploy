
## About

**tnascert-deploy** is a tool used to deploy TLS certificates to one or more TrueNAS-SCALE systems. The tool supports the TrueNAS-SCALE JSON-RPC 2.0 API and TrueNAS RESTful API v2.0.  You must select an API version to use for your NAS using the **client_api** configuration parameter, see the [**Configuration File**](#configuration-file) section below on how to set the **client_api** parameter.

With the proper setting the tool supports deploying certificates to the following TrueNAS versions:

| Product/Version | Supported API |
| --- | --- |
| TrueNAS-CORE 13.3 | RESTful API v2.0 |
| zVault 13.3-MASTER-202505042329 | RESTful API v2.0 |
| TrueNAS-SCALE 24.10 | RESTful API v2.0 |
| TrueNAS-SCALE 25+ | JSON-RPC 2.0 websocket API |

**tnas-certdeploy** is written in Go and, when compiled for your target system, there are no other dependencies but the binary itself: **tnascert-deploy**.

The tool connects to the JSON-RPC 2.0 WebSocket API or RESTful v2.0 API endpoint in order to deploy the certificates and private key for use as:

* TrueNAS UI certificate
* FTPS service certificate
* Docker App TLS certificates (for TrueNAS-SCALE systems utilizing Docker Apps)

**tnascert-deploy** utilizes an INI configuration file where multiple TrueNAS systems may be configured in separate sections of the file.  The user of the tool specifies one or more TrueNAS systems by their section name on the commandline defined in the configuration file in order to deploy certificates.

The tool may be utilized as part of an ACME (Automated Certificate Management Environment) process to deploy new or renewal certficates to TrueNAS systems, see the [sample-scripts](/sample-scripts) directory for examples.  The command line usage is as follows:

```
Usage: tnascert-deploy [-hv] [-c value] config_section ... config_section

-c, --config="full path to the configuration file [tnas-cert.ini]".
-h, --help print usage information and exit.
-v, --version print version information and exit
```

Example to deploy certficates to two TrueNAS machines nas01 and nas02:

    $ tnascert-deploy -c /etc/tnas-cert.ini nas01 nas02

##  Getting Started

Precompiled releases of **tnascert-deploy** are available for FreeBSD, Debian Linux, MacOS, or Windows 11. See the [Releases](https://github.com/jrushford/tnascert-deploy/releases) section of this repository. The current Release is [2.2](https://github.com/jrushford/tnascert-deploy/releases/tag/v2.2).

> [!NOTE]
> If your NAS is currently running with a self-signed or expired certificate, please set `tls_skip_verify` to `true` in the configuration file to avoid connection TLS verification errors.


### Testing

To build and test on any system with Go installed, clone this repository and run unit tests using:

    go test ./...

Alternatively, if you have make installed:

    make test

### Release Builds

For a statically-linked build of **tnascert-deploy**, use either of the following commands:

    make

or

    CGO_ENABLED=0 go build -ldflags="-s -w"

For a dynamically-linked binary, use:

    go build
   
Copy **tnascert-deploy** for use either as a command line tool or as part of your ACME deployment scripts and create an INI configuration file that lists all of your TrueNAS systems. 
    
##  Configuration file

By default, configuration information is loaded from the file **tnas-cert.ini** in the current working directory. To override this, pass the location of your configuration file using the `-c filename` option.

The configuration file uses the INI format that lists section names in square brackets followed by named value pairs separated by an equal sign. The `deploy_default` section name, if defined, will be used if no other section name is listed on the commandline.

The following shows an example configuration file with three TrueNAS systems configured.  In the example there are 3 sections defined: `deploy_default`, `nas02`, and `nas03`.  If no section is listed on the **tnascert-deploy** commandline, the `deploy_default` configuration will be loaded and certificates will be deployed to the TrueNAS host defined in that section.  Each individual NAS configuration can be loaded by listing only that desired section on the commandline. All 3 sections can be loaded and have certificates deployed in turn by listing all 3 sections on the commandline:

    tnascert-deploy deploy_default nas02 nas03

Since version 2.2, all the key values in the INI file may be loaded from the OS environment using the syntax `${VARIABLE_NAME}`.  The variable must be set in the environment in order to use it.  If the environment variable is not set, the program will exit with an error message while loading the configuration file. 

> [!IMPORTANT]
> Where needed, make sure you EXPORT the environment variables so they are available to the **tnascert-deploy** process.

For example, to set the FQDN of the `connect_host` and other sensitive fields from the environment use:

    connect_host = ${CONNECT_HOST}
    api_key = ${API_KEY}
    username = ${USERNAME}
    password = ${PASSWORD}

You can include multiple environment variables in a single entry:

    connect_host = ${CONNECT_HOST}.${DOMAIN_NAME}

### Sample configuration files

This is a basic config file using the default section and only hard-coded values.

```ini
[deploy_default]  
api_key = 1-ZFhoN97YrxqWg5GIR3XjhPNuaO7NKAwDBbwCashgTCi0z4Mfy9sYo8e8g4WPMCO2  
private_key_path = test_files/privkey.pem  
full_chain_path = test_files/fullchain.pem  
cert_basename = letsencrypt  
client_api = wsapi  
connect_host = nas01.mydomain.com  
protocol = wss  
tls_skip_verify = false  
delete_old_certs = true  
strict_basename_match = false  
add_as_ui_certificate = true  
add_as_ftp_certificate = true  
add_as_app_certificate = true  
app_list = webdav  
timeoutSeconds = 10  
debug = false  
```

```ini
# sample production config
[nas02]  
api_key = 1-ZFhoN97YrxqWg5GIR3XjhPNuaO7NKAwDBbwCashgTCi0z4Mfy9sYo8e8g4WPMCO2  
private_key_path = test_files/privkey.pem  
full_chain_path = test_files/fullchain.pem  
cert_basename = letsencrypt  
client_api = restapi  
connect_host = nas02.mydomain.com  
protocol = https  
tls_skip_verify = false  
delete_old_certs = true  
strict_basename_match = false  
add_as_ui_certificate = true  
add_as_ftp_certificate = true  
add_as_app_certificate = true  
app_list = gitea, webdav  
timeoutSeconds = 10  
debug = false  
  
# sample production config
[nas03]  
api_key = 2-AFhoB89YqxrWg5GIR3XjhPFUao7NKAwDBbWcAshgTCi0z47fM9sYo8e8g4wpMCO2  
cert_basename = letsencrypt  
private_key_path = test_files/privkey.pem  
full_chain_path = test_files/fullchain.pem  
client_api = wsapi  
connect_host = nas03.mydomain.com  
protocol = wss  
tls_skip_verify = true  
delete_old_certs = true  
strict_basename_match = false  
add_as_ui_certificate = false  
add_as_ftp_certificate = true  
add_as_app_certificate = true  
app_list = gitea, webdav, frigate  
timeoutSeconds = 10  
debug = false  
```

```ini
# sample configuration using some environment variables
[nas04]  
api_key = ${API_KEY}  
cert_basename = letsencrypt  
private_key_path = test_files/privkey.pem  
full_chain_path = test_files/fullchain.pem  
client_api = wsapi  
connect_host = ${CONNECT_HOST}.${DOMAIN_NAME}  
protocol = wss  
tls_skip_verify = true  
delete_old_certs = true  
strict_basename_match = false  
add_as_ui_certificate = false  
add_as_ftp_certificate = true  
add_as_app_certificate = true  
app_list = gitea, webdav, frigate  
timeoutSeconds = 10  
debug = false  
```

### Configuration File settings

In order to authenticate with a TrueNAS system, the user must either use the TrueNAS UI to generate and copy an **api_key** or use an admin **username** and **password** in the configuration file.  The **api_key** is preferred and if all three are defined in the configuration file, only the **api_key** will be used.  Do not include the **api_key** if you wish to use the **username** and **password**.

[id1]: ## "You must use either api_key or username/password"

| Key Name | Required |Default | Description |
| --- | --- | --- | --- |
| **api_key** [:information_source:][id1] | N | - | TrueNAS 64 byte API Key for login (this is the preferred login method). |
| **username** [:information_source:][id1] | N | - | TrueNAS username with admin privileges (API key is preferred for login). |
| **password** [:information_source:][id1] | N | - | TrueNAS password for user with admin privileges (API key is preferred for login). |
| **cert_basename** | N | **tnascert-deploy** | Basename for the certificate naming in TrueNAS. |
| **connect_host** | Y | - | TrueNAS DNS Fully Qualified Domain Name (FQDN) or IP address. |
| **client_api** | N | **wsapi** | The TrueNAS API to use: `wsapi` for the JSON-RPC 2.0 websocket API or `restapi` for the RESTful v2.0 API. |
| **delete_old_certs** | N | **false** | Whether to remove old certificates with the same basename after the new one has been installed. |
| **strict_basename_match** | N | **false** | When `true`, certificate names are checked more strictly before being deleted to reduce the chance of the basename matching incorrect certs. |
| **full_chain_path** | Y | - | Full path name to the certificate (full_chain.pem). |
| **private_key_path** | Y | - | Full path name to the certificate (private_key.pem). |
| **port** | N | **443** | TrueNAS API endpoint port. |
| **protocol**[^1] | N | **wss** | Using websockets: `ws` for insecure websockets or `wss` for secure websockets<br>Using RESTAPI: `http` for insecure HTTP or `https` for secure HTTP. |
| **tls_skip_verify** | N | **false** | Strict SSL cert verification of the endpoint. If your NAS is currently running with a self-signed or invalid certificate. Set this to avoid TLS verification errors. |
| **add_as_ui_certificate** | N | **false** | Install as the active UI certificate if `true`. |
| **add_as_ftp_certificate** | N | **false** | Install as the active FTP certificate if `true`. |
| **add_as_app_certificate** | N | **false** | If `true`, install the certificate for apps listed in the `app_list` |
| **app_list** | N | - | A comma separated list of docker apps that you wish to have the newly imported certificate used. Only works if they have a certificate assigned already. You must enable `add_as_app_certificate` to process the list. |
| **timeoutSeconds** | N | **10** | The number of seconds after which the TrueNAS client calls fail. |
| **debug** | N | **false** | Debug logging is enabled if `true`. |

[^1]: Websockets (`ws` and `wss`) are only for TrueNAS-SCALE systems utilizing the JSON-RPC 2.0 websocket API.  Use `http` or `https` for systems utilizing the RESTful v2.0 API.

## Notes

This tool uses the TrueNAS RESTful v2.0 API or TrueNAS Scale JSON-RPC 2.0 API and the TrueNAS client API module.

Support for **TrueNAS 25.04** systems or later, TrueNAS 24.10, TrueNAS-CORE, and zVault systems are supported using the RESTful v2.0 API.

## See Also

+ [TrueNAS api_client_golang](https://github.com/truenas/api_client_golang)
+ [TrueNAS websocket API documentaion](https://www.truenas.com/docs/api/scale_websocket_api.html)

## Sponsor

If you find this tool useful, consider buying me a cup of coffee or
sponsoring me by hitting the ***Sponsor*** button at the top of this
page..  
### Contact
+ John J. Rushford
+ jrushford@apache.org
