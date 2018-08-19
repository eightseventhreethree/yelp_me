# yelp_me
A CLI tool for finding food on Yelp using the Yelp API. 
Requires an API token from Yelp. 
Guide found here: https://www.yelp.com/developers/documentation/v3/authentication

````
./yelp_me --help
Usage of ./yelp_me_linux:
  -d, --distance int    Distance in miles around the zip you are willing to look. NOT REQUIRED (default 10)
  -s, --search string   Keyword to search for on Yelp. REQUIRED
  -z, --zip int         Zip code to search around. REQUIRED (default 12345)
````
To build from source. 
1. Install Go. Guide here: https://golang.org/doc/install
2. Install the required packages. 
  ````
   go get github.com/Jeffail/gabs
   go get github.com/spf13/pflag
   go get github.com/spf13/viper
   go get gopkg.in/resty.v1
  ````
3. ````go build ./yelp_me.go````
4. Profit.
