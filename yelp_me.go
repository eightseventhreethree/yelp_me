package main

import (
	"fmt"
	gabs "github.com/Jeffail/gabs"
	log "github.com/sirupsen/logrus"
	pflag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
	"gopkg.in/resty.v1"
	"os"
	"strconv"
)

// Get configuration from .yml file
func ReadConfig() (string, string) {
	viper.SetConfigName(".go_grub")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		fatalStr := "Fatal error config file: %s \n" +
			"Place config file in $HOME/.go_grub.yml\n" +
			"yelp: \n  api_url: apiurlgoeshere (not required has default set)\n" +
			"  api_token: apitokegoeshere \n"
		panic(fmt.Errorf(fatalStr, err))
	}
	viper.SetDefault("yelp.api_url", "https://api.yelp.com/v3/")
	yelpAPIUrl := viper.Get("yelp.api_url")
	yelpAPIToken := viper.Get("yelp.api_token")
	return yelpAPIUrl.(string), yelpAPIToken.(string)
}

// Get CLI arguments and evaluate them
func ParseCLI() (string, int, int) {
	var searchArg string
	var zipArg int
	var distanceArg int
	var verbose bool
	pflag.StringVarP(&searchArg, "search", "s", "", "Keyword to search for on Yelp. REQUIRED")
	pflag.IntVarP(&zipArg, "zip", "z", 12345, "Zip code to search around. REQUIRED")
	pflag.IntVarP(&distanceArg, "distance", "d", 10, "Distance in miles around the zip you are willing to look. NOT REQUIRED")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
	pflag.Parse()
	if verbose == false {
		log.SetLevel(log.WarnLevel)
	} else {
		log.SetLevel(log.DebugLevel)
		resty.SetDebug(true)
	}
	log.Debug("flag searchArg = ", searchArg)
	log.Debug("flag zipArg = ", zipArg)
	log.Debug("flag distanceArg = ", distanceArg)
	if searchArg == "" {
		fmt.Println("No arguments passed. Use --help to find out more.")
		os.Exit(1)
	}
	// Convert from miles to meters to use in request
	distanceArgMeter := distanceArg * 1609
	log.Debug("distanceArgMeter = ", distanceArgMeter)
	return searchArg, zipArg, distanceArgMeter
}

// Resty global configuration
func RestyConfig(yelpAPIUrl string, yelpAPIToken string) {
	// Host URL for all request. So you can use relative URL in the request
	resty.SetHostURL(yelpAPIUrl)

	// Headers for all request
	resty.SetHeader("Accept", "application/json")
	resty.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "go_grub",
	})
	resty.SetAuthToken(yelpAPIToken)
}

// Get the list of buisnesses based on the requested arguments
func RequestBuisnessSearch(keywordSearch string, zipSearch int, distanceSearch int) []byte {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"term":       keywordSearch,
			"location":   strconv.Itoa(zipSearch),
			"radius":     strconv.Itoa(distanceSearch),
			"categories": "restaurants",
			"open_now":   "true",
			"sort_by":    "rating",
		}).
		Get("businesses/search")
	log.Debug("Response Body: ", resp.String())
	if err != nil {
		fmt.Printf("\nResponse Err: %v", err)
	}
	return resp.Body()
}

// Parse the response from RequestBuisnessSearch
func ParseResponse(input []byte) {
	jsonParsed, err := gabs.ParseJSON(input)
	if err != nil {
		fmt.Printf("\ngabs.ParseJson err: %v", err)
	}
	children, _ := jsonParsed.S("businesses").Children()
	for _, child := range children {
		fmt.Println("| Name:", child.Search("name").Data(), "| Review count: ", child.Search("review_count"), "| Rating:", child.Search("rating").Data(), "| Price: ", child.Search("price"), "|")
	}
}

func main() {
	yelpAPIUrl, yelpAPIToken := ReadConfig()
	keywordSearch, zipSearch, distanceSearch := ParseCLI()
	RestyConfig(yelpAPIUrl, yelpAPIToken)
	buisnessBody := RequestBuisnessSearch(keywordSearch, zipSearch, distanceSearch)
	ParseResponse(buisnessBody)
}
