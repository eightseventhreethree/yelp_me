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

type ConfigObject interface {
	GetAPIURL() string
	GetAPIToken() string
	GetDebug() bool
	GetSearchValue() string
	GetZipCode() int
	GetDistanceValue() int
}

type InputStruct struct {
	getAPIURL        string
	getAPIToken      string
	getDebug         bool
	getSearchValue   string
	getZipCode       int
	getDistanceValue int
}

func (i *InputStruct) GetAPIURL() string {
	if i.getAPIURL == "" {
		return "<not implemented>"
	}
	return i.getAPIURL
}

func (i *InputStruct) GetAPIToken() string {
	if i.getAPIToken == "" {
		return "<not implemented>"
	}
	return i.getAPIToken
}

func (i *InputStruct) GetDebug() bool {
	return i.getDebug
}

func (i *InputStruct) GetSearchValue() string {
	if i.getSearchValue == "" {
		fmt.Println("No arguments passed. Use --help to find out more.")
		os.Exit(1)
	}
	return i.getSearchValue
}

func (i *InputStruct) GetZipCode() int {
	return i.getZipCode
}

func (i *InputStruct) GetDistanceValue() int {
	return i.getDistanceValue
}

func Input() ConfigObject {
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
			"  api_token: apitokegoeshere \n" +
			"debug: true (not required)"
		panic(fmt.Errorf(fatalStr, err))
	}
	config.SetDefault("yelp.api_url", "https://api.yelp.com/v3/")
	config.SetDefault("debug", false)
	debug := config.GetBool("debug")

	var searchValue string
	var zipCode int
	var distance int
	pflag.StringVarP(&searchValue, "search", "s", "", "Keyword to search for on Yelp. REQUIRED")
	pflag.IntVarP(&zipCode, "zip", "z", 12345, "Zip code to search around. REQUIRED")
	pflag.IntVarP(&distance, "distance", "d", 10, "Distance in miles around the zip you are willing to look. NOT REQUIRED")
	pflag.Parse()
	if debug {
		fmt.Println("Here is the value of flag searchValue: ", searchValue)
		fmt.Println("Here is the value of flag zipCode: ", zipCode)
		fmt.Println("Here is the value of flag distanceArg: ", distance)
	}

	return &InputStruct{
		getAPIURL:        config.Get("yelp.api_url").(string),
		getAPIToken:      config.Get("yelp.api_token").(string),
		getDebug:         debug,
		getSearchValue:   searchValue,
		getZipCode:       zipCode,
		getDistanceValue: distance * 1609,
	}
}


type Yelp struct {
	yelpAPIUrl   string
	yelpAPIToken string
	debug        bool
	keyword      string
	zipCode      int
	distance     int
}

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

func main() {
	input := Input()
	fmt.Println("input.GetAPIToken = ", input.GetAPIToken())
	yelp := Yelp{input.GetAPIURL(), input.GetAPIToken(), input.GetDebug(), input.GetSearchValue(), input.GetZipCode(), input.GetDistanceValue()}
	yelp.RestyConfig()
	resp := yelp.RequestBuisnessSearch()
	yelp.ParseResponse(resp)
}
