package main

import (
	"fmt"
	"os"
	gabs "github.com/Jeffail/gabs"
	pflag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
	"gopkg.in/resty.v1"
	"strconv"
)

func ReadConfig() (string, string, bool) {
	viper.SetConfigName(".go_grub")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/")
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		fatalStr := "Fatal error config file: %s \n" +
          "Place config file in $HOME/.go_grub.yml\n" + 
          "yelp: \n  api_url: apiurlgoeshere (not required has default set)\n" +
          "  api_token: apitokegoeshere \n" +
					"debug: true (not required)"
		panic(fmt.Errorf(fatalStr, err))
	}
	viper.SetDefault("yelp.api_url", "https://api.yelp.com/v3/")
	viper.SetDefault("debug", false)
	yelpAPIUrl := viper.Get("yelp.api_url")
	yelpAPIToken := viper.Get("yelp.api_token")
	debug := viper.GetBool("debug")
	if debug {
		fmt.Println("Here is the value of debugMode: ", debug)
	}
	return yelpAPIUrl.(string), yelpAPIToken.(string), debug
}

func ParseCLI(debugMode bool) (string, int, int) {
	var searchArg string
	var zipArg int
	var distanceArg int
	pflag.StringVarP(&searchArg, "search", "s", "", "Keyword to search for on Yelp. REQUIRED")
	pflag.IntVarP(&zipArg, "zip", "z", 12345, "Zip code to search around. REQUIRED")
	pflag.IntVarP(&distanceArg, "distance", "d", 10, "Distance in miles around the zip you are willing to look. NOT REQUIRED")
	pflag.Parse()
	if debugMode {
		fmt.Println("Here is the value of flag searchArg: ", searchArg)
		fmt.Println("Here is the value of flag zipArg: ", zipArg)
		fmt.Println("Here is the value of flag distanceArg: ", distanceArg)
	}
	if searchArg == "" {
		fmt.Println("No arguments passed. Use --help to find out more.")
		os.Exit(1)
	}
	distanceArgMeter := distanceArg * 1609
	return searchArg, zipArg, distanceArgMeter
}

func RestyConfig(yelpAPIUrl string, yelpAPIToken string, debug bool) {
	// Host URL for all request. So you can use relative URL in the request
	resty.SetHostURL(yelpAPIUrl)
	resty.SetDebug(debug)

	// Headers for all request
	resty.SetHeader("Accept", "application/json")
	resty.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "go_grub",
	})
	resty.SetAuthToken(yelpAPIToken)
}

func RequestBuisnessSearch(keywordSearch string, zipSearch int, distanceSearch int) []byte {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"term":     keywordSearch,
			"location": strconv.Itoa(zipSearch),
			"radius":   strconv.Itoa(distanceSearch),
      "categories": "restaurants",
			"open_now": "true",
			"sort_by":  "rating",
		}).
		Get("businesses/search")
	//fmt.Printf("\nResponse Body: %v", resp.String())
	if err != nil {
		fmt.Printf("\nResponse Err: %v", err)
	}
  return resp.Body()
}

func ParseResponse(input []byte) {
  jsonParsed, err := gabs.ParseJSON(input)
  if err != nil {
		fmt.Printf("\ngabs.ParseJson err: %v", err)
  }
  children, _ := jsonParsed.S("businesses").Children()
  for _, child := range children {
	  fmt.Println("| Name:", child.Search("name").Data(), "| Review count: ", child.Search("review_count"), "| Rating:",  child.Search("rating").Data(), "| Price: ", child.Search("price"), "|")
  }
}

func main() {
	yelpAPIUrl, yelpAPIToken, debug := ReadConfig()
	keywordSearch, zipSearch, distanceSearch := ParseCLI(debug)
	RestyConfig(yelpAPIUrl, yelpAPIToken, debug)
	buisnessBody := RequestBuisnessSearch(keywordSearch, zipSearch, distanceSearch)
  ParseResponse(buisnessBody)
}
