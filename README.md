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
