package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"strings"

	"os"

	"flag"

	"github.com/ChimeraCoder/anaconda"
)

type FilterConfig struct {
	Follow []string `json:"follow"`
	Track  []string `json:"track"`
}

var filterConfig FilterConfig

func loadConfig(configPath string, conf *FilterConfig) error {

	file, err := ioutil.ReadFile(configPath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(file, &conf); err != nil {
		return err
	}

	return nil
}

type Client struct {
	API anaconda.TwitterApi
}

func NewClient(consumerKey, consumerSecret, accessToken, accessSecret string) *Client {

	anaconda.SetConsumerKey(consumerKey)
	anaconda.SetConsumerSecret(consumerSecret)
	api := anaconda.NewTwitterApi(accessToken, accessSecret)
	return &Client{
		API: *api,
	}
}

func (client *Client) getUserIdsFromScreenNames(screenNames []string) ([]string, error) {
	v := url.Values{}
	v.Set("screen_name", strings.Join(filterConfig.Follow, ","))
	friendShips, err := client.API.GetFriendshipsLookup(v)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, v := range friendShips {
		ids = append(ids, v.Id_str)
	}
	return ids, nil
}

func (client *Client) getTwitterStream(filterConfig FilterConfig) (*anaconda.Stream, error) {

	ids, err := client.getUserIdsFromScreenNames(filterConfig.Follow)
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	if len(ids) != 0 {
		v.Set("follow", strings.Join(ids, ","))
	}
	if len(filterConfig.Track) != 0 {
		v.Set("track", strings.Join(filterConfig.Track, ","))
	}
	return client.API.PublicStreamFilter(v), nil
}

func print(tweet anaconda.Tweet) {
	header := fmt.Sprintf("\x1b[32m%s @%s\x1b[0m (Follows: %d Followers: %d) https://twitter.com/%s/status/%d",
		tweet.User.Name, tweet.User.ScreenName, tweet.User.FriendsCount, tweet.User.FollowersCount, tweet.User.ScreenName, tweet.Id)
	text := fmt.Sprintf("\x1b[36m%s\x1b[0m", tweet.Text)
	footer := fmt.Sprintf("RT: %d Fav: %d %s\n-----------------------------------------------",
		tweet.RetweetCount, tweet.FavoriteCount, tweet.CreatedAt)

	fmt.Println(header)
	fmt.Println(text)
	fmt.Println(footer)
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "c", "config.json", "The config file path")
	flag.Parse()
	if err := loadConfig(configPath, &filterConfig); err != nil {
		fmt.Println(err)
		return
	}

	consumerKey := os.Getenv("TWITTER_CONSUMER_KEY")
	consumerSecret := os.Getenv("TWITTER_CONSUMER_SECRET")
	accessToken := os.Getenv("TWITTER_ACCESS_TOKEN")
	accessSecret := os.Getenv("TWITTER_ACCESS_SECRET")

	if consumerKey == "" || consumerSecret == "" || accessToken == "" || accessSecret == "" {
		fmt.Println("set TWITTER_CONSUMER_KEY, TWITTER_CONSUMER_SECRET, TWITTER_ACCESS_TOKEN and TWITTER_ACCESS_SECRET as environment variables")
		return
	}

	client := NewClient(consumerKey, consumerSecret, accessToken, accessSecret)
	twitterStream, err := client.getTwitterStream(filterConfig)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		select {
		case x := <-twitterStream.C:
			switch tweet := x.(type) {
			case anaconda.Tweet:
				print(tweet)
			default:

			}
		}
	}

}
