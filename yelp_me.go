package main

import (
	"fmt"
	"os"
	"strconv"

	gabs "github.com/Jeffail/gabs"
	pflag "github.com/spf13/pflag"
	viper "github.com/spf13/viper"
	"gopkg.in/resty.v1"
)

// Define methods to access all the input values from the CLI or the config file
type Config interface {
	GetAPIURL() string
	GetAPIToken() string
	GetVerbose() bool
	GetSearchValue() string
	GetZipCode() int
	GetDistanceValue() int
}

// Define all the values for accessing
type Input struct {
	getAPIURL        string
	getAPIToken      string
	getVerbose       bool
	getSearchValue   string
	getZipCode       int
	getDistanceValue int
}

// Get the value from Base
func (i *Input) GetAPIURL() string {
	if i.getAPIURL == "" {
		return "<not implemented>"
	}
	return i.getAPIURL
}
// Get the value from Base
func (i *Input) GetAPIToken() string {
	if i.getAPIToken == "" {
		return "<not implemented>"
	}
	return i.getAPIToken
}
// Get the value from Base
func (i *Input) GetVerbose() bool {
	return i.getVerbose
}
// Get the value from Base
func (i *Input) GetSearchValue() string {
	if i.getSearchValue == "" {
		fmt.Println("No arguments passed. Use --help to find out more.")
		os.Exit(1)
	}
	return i.getSearchValue
}
// Get the value from Base
func (i *Input) GetZipCode() int {
	return i.getZipCode
}
// Get the value from Base
func (i *Input) GetDistanceValue() int {
	return i.getDistanceValue
}

// Get config from the yml file and parse the cli arguments and store them in Input struct
func Base() Config {
	config := viper.New()
	config.SetConfigName(".go_grub")
	config.AddConfigPath(".")
	config.AddConfigPath("$HOME/")
	config.SetConfigType("yaml")
	err := config.ReadInConfig() // Find and read the config file
	if err != nil {              // Handle errors reading the config file
		fatalStr := "Fatal error config file: %s \n" +
			"Place config file in $HOME/.go_grub.yml\n" +
			"yelp: \n  api_url: apiurlgoeshere (not required has default set)\n" +
			"  api_token: apitokegoeshere \n"
		panic(fmt.Errorf(fatalStr, err))
	}
	config.SetDefault("yelp.api_url", "https://api.yelp.com/v3/")

	var searchValue string
	var zipCode int
	var distance int
	var verbose bool
	pflag.StringVarP(&searchValue, "search", "s", "", "Keyword to search for on Yelp. REQUIRED")
	pflag.IntVarP(&zipCode, "zip", "z", 12345, "Zip code to search around. REQUIRED")
	pflag.IntVarP(&distance, "distance", "d", 10, "Distance in miles around the zip you are willing to look. NOT REQUIRED")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
	pflag.Parse()
	if verbose {
		fmt.Println("Here is the value of flag searchValue: ", searchValue)
		fmt.Println("Here is the value of flag zipCode: ", zipCode)
		fmt.Println("Here is the value of flag distanceArg: ", distance)
	}

	return &Input{
		getAPIURL:        config.Get("yelp.api_url").(string),
		getAPIToken:      config.Get("yelp.api_token").(string),
		getVerbose:       verbose,
		getSearchValue:   searchValue,
		getZipCode:       zipCode,
		getDistanceValue: distance * 1609,
	}
}

// Create base for talking to Yelp
type Yelp struct {
	yelpAPIUrl   string
	yelpAPIToken string
	debug        bool
	keyword      string
	zipCode      int
	distance     int
}

// Configure Resty to make calls
func (y *Yelp) RestyConfig() {
	// Host URL for all request. So you can use relative URL in the request
	resty.SetHostURL(y.yelpAPIUrl)
	resty.SetDebug(y.debug)

	// Headers for all request
	resty.SetHeader("Accept", "application/json")
	resty.SetHeaders(map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   "go_grub",
	})
	resty.SetAuthToken(y.yelpAPIToken)
}

// Search for businesses with the defined values from the CLI sorted by rating
func (y *Yelp) RequestBuisnessSearch() []byte {
	resp, err := resty.R().
		SetQueryParams(map[string]string{
			"term":       y.keyword,
			"location":   strconv.Itoa(y.zipCode),
			"radius":     strconv.Itoa(y.distance),
			"categories": "restaurants",
			"open_now":   "true",
			"sort_by":    "rating",
		}).
		Get("businesses/search")
	//fmt.Printf("\nResponse Body: %v", resp.String())
	if err != nil {
		fmt.Printf("\nResponse Err: %v", err)
	}
	if resp.StatusCode() != 200 {
		fmt.Println("Error: ", resp.String())
	}
	return resp.Body()
}

// Parse the json response and put into a table format
func (y *Yelp) ParseResponse(input []byte) {
	jsonParsed, err := gabs.ParseJSON(input)
	if err != nil {
		fmt.Printf("\ngabs.ParseJson err: %v", err)
	}
	children, _ := jsonParsed.S("businesses").Children()
	for _, child := range children {
		fmt.Println("| Name:", child.Search("name").Data(), "| Review count: ", child.Search("review_count"), "| Rating:", child.Search("rating").Data(), "| Price: ", child.Search("price"), "|")
	}
}

// Call Base and Yelp
func main() {
	base := Base()
	yelp := Yelp{base.GetAPIURL(), base.GetAPIToken(), base.GetVerbose(), base.GetSearchValue(), base.GetZipCode(), base.GetDistanceValue()}
	yelp.RestyConfig()
	resp := yelp.RequestBuisnessSearch()
	yelp.ParseResponse(resp)
}
