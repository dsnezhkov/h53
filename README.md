# h53
A utility to force query DNS over DNS over HTTPS (DoH) bounced off of CloudFlare API when DNS block is in place



h53 is a small utility to attempt to query DNS over HTTPS (DoH) against
CloudFlare DoH server 1.1.1.1

## Use Case: 
Some companies use DNS blocking without actually blocking access to destinations IPs

In this case we can use CloudFlare's DoH service to resolve the IP and further connect to site.

h53 interfaces with the DoH service API:
_Links:_
 	
- https://developers.cloudflare.com/1.1.1.1/dns-over-https/
- https://developers.cloudflare.com/1.1.1.1/dns-over-https/json-format/

## Usage:
```
h53 <options>:
  -T int
       Query Timeout (sec.) Ex.: 10 (default 10)
  -d    Debug Lookups
  -n string
        Query Name Ex.: example.com
  -t string
        Query Type (either a numeric value or text) Ex: A, AAAA.
        Note: list of types can be found here: https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4
  -v    Display Verbose processing

 Examples:
    h53 -t MX -n ibm.com  -v
    0: ibm.com. - 129.42.38.10
    
    h53 -t A -n tencent.com
    0: tencent.com. - 113.105.73.141
    1: tencent.com. - 113.105.73.142
    2: tencent.com. - 113.105.73.148
...

    h53 -t A -n google.com  -d -v
```
