% TNASCERT-DEPLOY(1) Version 2.2 | User Commands

## NAME

tnascert-deploy - deploy certificates to a TrueNAS host

## SYNOPSIS

**tnascert-deploy** [OPTION]... [SECTION]...

## DESCRIPTION

Deploy certificates to TrueNAS based on settings in configuration file under SECTION(s).

A tool used to import a TLS certificate and private key into a TrueNAS SCALE host running one of **TrueNAS 24.14**, **TrueNAS 25+**, **TrueNAS-CORE**, or **zVault**.

Once a certificate is imported the tool can configure the TrueNAS host to use it as:

* The main UI TLS certificate
* The FTP certificate
* The certificate for Docker applications

The **tnas-cert.ini** file consists of multiple **sections**. The optional command line arguments for **section_name** may be used to load that particular configuration. This allows for multiple configurations in one configuration file; either different configurations for a single host, or separate configurations for several hosts.

You may list multiple **section_name** arguments on the command line to loop through certificate installation on multiple **TrueNAS** hosts.

If the optional argument ***section_name*** is not provided, The ***deploy_default*** section name is chosen to load the configuration if it exists.

For client authentication, you may use either a TrueNAS ***api_key*** or the ***username*** and ***password*** for a user with admin privileges. The ***api_key*** is the preferred method for authentication.  If all three variables are set in a configuration, the ***api_key*** login method will always be used.  If you prefer to use the ***username*** and ***password*** ensure only the ***username*** and ***password*** are set in your configuration.

See the sample **tnas-cert.ini** file.

## OPTIONS

**-c, --config=FILE**
: load configuration from FILE

**-h, --help**
: show brief command usage

**-v, --version**
: show version information

## EXAMPLES

The minimal command-line to upload certificates based on the section **deploy_default** in the file **tnas-cert.ini** in the current working directory:

```
tnascert-deploy
```

To specify a different location and/or filename for the configuration file:

```
tnascert-deploy --config /opt/truenas/deploy.ini
```

To deploy certificates using configuration in sections **nas01** and **nas02** with a non-default configuration file location:

```
tnascert-deploy --config ~/tnas-cert.ini nas01 nas02
```

## FILES

The default configuration file is named **tnas-cert.ini** in the current working directory. You may use the command line switch to use another file name and full path to the config file.

## CONFIGURATION

In order to authenticate with a TrueNAS system, the user must either use the TrueNAS UI to generate and copy an **api_key** or use an admin **username** and **password** in the configuration file.  The **api_key** is preferred and if all three are defined in the configuration file, only the **api_key** will be used.  Do not include the **api_key** if you wish to use the **username** and **password**.

All the key values in the configuration INI file may be interpolated from the OS environment using the syntax **${VARIABLE_NAME}** if desired. Environment variables must be set otherwise **tnascert-deploy** will exit with an error if you are trying to use an environment variable that is not set.

For example to set sensitive fields from the environment, you can use:

    api_key = ${API_KEY}
    username = ${USERNAME}
    password = ${PASSWORD}

You can also use multiple environment variables together in one entry:

    connect_host = ${HOSTNAME}.${DOMAIN_NAME}

The available keys are:

**api_key**  (optional, no default)  
: TrueNAS 64 byte API Key for login the preferred login method).

**username**  (optional, no default)  
: TrueNAS username with admin privileges (API key is preferred for login)

**password**  (optional, no default)
: TrueNAS password for user with admin privileges, (API key is preferred for login)

**cert_basename**  (optional, default is **"tnascert-deploy"**)
: basename for the certificate naming in TrueNAS.

**connect_host**  (required)
: TrueNAS DNS Fully Qualified Domain Name, FQDN, or IP address

**client_api**  (optional, default is "wsapi")
: The TrueNAS API to use. Choices are: 'wsapi' for the JSON-RPC 2.0 websocket API or 'restapi' for the RESTful v2.0 API.

**delete_old_certs**  (optional, default is **false**)
: whether to remove old certificates, default is false

**strict_basename_match**  (optional, default is **false**)
: when true, certificate names are checked more strictly before being deleted to reduce the chance of the basename matching incorrect certs

**full_chain_path**  (required)
: full path name to the certificate full_chain.pem

**private_key_path**  (required)
: full path name to the certificate private_key.pem

**port**  (optional, default is **443**)
: TrueNAS API endpoint port

**protocol**  (optional, default is **"wss"**)
: websocket protocol 'ws', 'wss', 'http', or 'https'. 'ws' and 'wss' are only for TrueNAS-SCALE systems using the JSON-RPC 2.0 websocket API. Use 'http' or 'https' for systems utilizing the RESTful v2.0 API.

**tls_skip_verify**  (optional, default is **false**)
: strict SSL cert verification of the endpoint.

**add_as_ui_certificate**  (optional, default is **false**)
: install as the active UI certificate if true

**add_as_ftp_certificate**  (optional, default is **false**)
: install as the active FTP service certificate if true

**add_as_app_certificate**  (optional, default is **false**)
: install as the active APP service certificate if true to the apps listed in the **app_list**

**app_list**  (optional, no default)
: A comma separated list of docker apps that you wish to have the newly imported certificate used. Apps in the list are only set to used the certificate if they have one assigned already. You must enable **add_as_app_certificate** to process the list.

**timeoutSeconds**  (optional, default is **10**)
: the number of seconds after which the truenas client calls fail

**debug**  (optional, default is **false**)
: debug logging if true

## NOTES

This tool uses the TrueNAS Scale JSON-RPC 2.0 API and the TrueNAS client API module. Product support:

| Product/Version | Supported API |
| ------ | --- |
| TrueNAS-CORE 13.3 | RESTful API v2.0 |
| zVault 13.3-MASTER-202505042329 | RESTful API v2.0 |
| TrueNAS-SCALE 24.10 | RESTful API v2.0 |
| TrueNAS-SCALE 25+ | JSON-RPC 2.0 websocket API |

See also:  
+ [TrueNAS api_client_golang](https://github.com/truenas/api_client_golang)\
+ [TrueNAS websocket API documentaion](https://www.truenas.com/docs/api/scale_websocket_api.html)

Clone this repository and build the tool using ***go build***.

## AUTHOR

Written by John J. Rushford (jrushford@apache.org)

## REPORTING BUGS

Report bugs to: https://github.com/jrushford/tnascert-deploy/issues
