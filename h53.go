//
// h53 is a small utility to attempt to query DNS over HTTPS (DoH) against
// CloudFlare DoH server 1.1.1.1
//
// Use Case: Some companies use DNS blocking without actually blocking access to destinations IPs
// In this case we can use CloudFlare's DoH service to resolve the IP and further connect to site.
//
// This client interfaces with the DoH service API:
// Links: 	https://developers.cloudflare.com/1.1.1.1/dns-over-https/
// 			https://developers.cloudflare.com/1.1.1.1/dns-over-https/json-format/
//
// Usage:
//  -T int
//        Query Timeout (sec.) Ex.: 10 (default 10)
//  -d    Debug Lookups
//  -n string
//        Query Name Ex.: example.com
//  -t string
//        Query Type (either a numeric value or text) Ex: A, AAAA.
//        Note: list of types can be found here: https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4
//  -v    Display Verbose processing
//
// Examples:
// 		h53 -t MX -n ibm.com  -v
//				0: ibm.com. - 129.42.38.10
//		h53 -t A -n tencent.com
//0: tencent.com. - 113.105.73.141
//1: tencent.com. - 113.105.73.142
//2: tencent.com. - 113.105.73.148
//3: tencent.com. - 113.107.238.11
//4: tencent.com. - 113.107.238.12
//5: tencent.com. - 113.107.238.14
//6: tencent.com. - 113.107.238.15
//7: tencent.com. - 113.107.238.27
//8: tencent.com. - 119.147.33.33
//9: tencent.com. - 119.147.33.36
//10: tencent.com. - 119.147.253.23
//11: tencent.com. - 119.147.253.25
//12: tencent.com. - 183.56.150.142
//13: tencent.com. - 183.56.150.144
//14: tencent.com. - 183.56.150.146
//15: tencent.com. - 183.56.150.155
//
// 		h53 -t A -n ibm.com  -d -v
// 		h53 -t A -n ibm.com

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"
)

type Question struct {
	Name string `json:"name"`
	Type int    `json:"type"`
}

type Answer struct {
	Name string `json:"name"`
	Type int    `json:"type"`
	TTL  int    `json:"TTL"`
	Data string `json:"data"`
}

type DNSJ struct {
	Status    int        `json:"Status"`
	TC        bool       `json:"TC"`
	RD        bool       `json:"RD"`
	RA        bool       `json:"RA"`
	AD        bool       `json:"AD"`
	CD        bool       `json:"CD"`
	Questions []Question `json:"Question"`
	Answers   []Answer   `json:"Answer"`
}

func main() {

	// Options
	var optType string
	var optName string
	var optTimeout int
	var optVerbose bool
	var optDebug bool

	flag.BoolVar(&optDebug, "d", false,
		"Debug Lookups")
	flag.BoolVar(&optVerbose, "v", false,
		"Display Verbose processing")
	flag.StringVar(&optType, "t", "",
		"Query Type (either a numeric value or text) Ex: A, AAAA. "+
			"\nNote: list of types can be found here: https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-4 ")
	flag.StringVar(&optName, "n", "",
		"Query Name Ex.: example.com")
	flag.IntVar(&optTimeout, "T", 10,
		"Query Timeout (sec.) Ex.: 10")

	flag.Parse()

	flagset := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) { flagset[f.Name] = true })

	if !flagset["t"] || !flagset["n"] {
		fmt.Fprint(os.Stderr, "Query Type (-t) or Name (-d) is NOT set.\n")
		os.Exit(1)
	}

	// Client
	var rdump []byte
	var u url.URL

	c := http.Client{Timeout: time.Duration(optTimeout) * time.Second}

	u.Scheme = "https"
	u.Host = "cloudflare-dns.com"
	u.Path = "dns-query"

	q := u.Query()
	q.Set("name", optName)
	q.Set("type", optType)
	u.RawQuery = q.Encode()

	if flagset["d"] {
		log.Printf("Host: %s, Query: %s\n", u.Host, u.RawQuery)
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create a GET request: %v\n", err)
		os.Exit(2)

	}
	req.Header.Set("accept", "application/dns-json")

	if flagset["d"] {
		rdump, err = httputil.DumpRequest(req, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to dump outgoing request: %v\n", err)
		} else {
			fmt.Printf("[DEBUG:REQUEST] \n%s\n", rdump)
		}

	}

	res, err := c.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error fetching response from the provider: %v", err)
		os.Exit(3)
	}

	defer res.Body.Close()

	if flagset["d"] {
		rdump, _ = httputil.DumpResponse(res, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to dump outgoing request: %v\n", err)
		} else {
			fmt.Printf("[DEBUG:RESPONSE] \n%s\n", rdump)
		}
	}

	// Parse response
	jdns := new(DNSJ)
	err = json.NewDecoder(res.Body).Decode(jdns)
	if err != nil {
		fmt.Fprintf(os.Stderr,
			"Error decoding response from cf: %v. May want to rerun with debug on.", err)
		os.Exit(4)
	}

	if jdns.Status != 0 {
		fmt.Printf("Unsuccessful DNS Return code: %d", jdns.Status)
		fmt.Printf("See: %s to determine the cause",
			"https://www.iana.org/assignments/dns-parameters/dns-parameters.xhtml#dns-parameters-6")
	}
	if flagset["v"] {
		fmt.Printf("Status: %d, TC: %t, RD: %t RA: %t, AD: %t, CD: %t\n",
			jdns.Status, jdns.TC, jdns.RD, jdns.RA, jdns.AD, jdns.CD)

		fmt.Printf("Questions: %d\n", len(jdns.Questions))
		if len(jdns.Questions) != 0 {
			for i, a := range jdns.Questions {
				fmt.Printf("%d: Name: %s Type: %d \n", i+1, a.Name, a.Type)
			}
		}
		fmt.Printf("Answers: %d\n", len(jdns.Answers))

		if len(jdns.Answers) != 0 {
			for i, a := range jdns.Answers {
				fmt.Printf("%d: %s - %s \n", i+1, a.Name, a.Data)
			}
		} else {
			fmt.Println("NOT FOUND")
			os.Exit(4)
		}
	} else {
		if len(jdns.Answers) != 0 {
			for i, a := range jdns.Answers {
				fmt.Printf("%d: %s - %s \n", i, a.Name, a.Data)
			}
		} else {
			fmt.Println("NOT FOUND")
			os.Exit(4)
		}
	}

}
