package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/chromedp/cdproto/target"
	"github.com/chromedp/chromedp"
)

func main() {
	var (
		VideoDelay time.Duration // The delay between switching videos
		delay      int           // The delay time in seconds (used to set VideoDelay)
		PathToList string        // The path to the file containing a list of YouTube video links
		VideoLinks []string      // A slice containing the YouTube video links
		openedCtx  context.Context
	)

	const (
		btPlayer string = "#movie_player > div.ytp-cued-thumbnail-overlay > button"                                          // Selector for the "play" button on the video thumbnail
		btPlay   string = "#movie_player > div.ytp-chrome-bottom > div.ytp-chrome-controls > div.ytp-left-controls > button" // Selector for the "play" button on the video player controls
	)

	// Set the command-line flags
	flag.IntVar(&delay, "delay", 10, "How long it takes to switch to the next video (in seconds). The default setting is 10.")
	flag.StringVar(&PathToList, "list", "", "The path to your own list of YouTube videos. By default, there are 10 links in the script.")
	flag.Parse()

	// Set VideoDelay and VideoLinks
	VideoDelay = time.Duration(delay)
	VideoLinks = GetLinkList(&PathToList)
	fmt.Println(VideoLinks)

	// Set Chrome options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("mute-audio", false),
		chromedp.Flag("enable-automation", false),
		chromedp.Flag("start-fullscreen", true),
	)

	// Create an allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Create a new browser context
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	// Create a map to store opened videos
	OpenedVideos := make(map[string]context.Context)

	// Set the random seed
	rand.Seed(time.Now().UnixNano())

	// Select a random video link and play it
	randomIndex := rand.Intn(len(VideoLinks))
	videoLink := VideoLinks[randomIndex]
	chromedp.Run(ctx,
		chromedp.Navigate(videoLink),
		chromedp.Click(btPlayer, chromedp.ByQuery),
		chromedp.Sleep(VideoDelay*time.Second),
		chromedp.Click(btPlay, chromedp.ByQuery),
	)
	OpenedVideos[videoLink] = ctx

	// Loop through the list of video links
	for {
		// Check if the opened context for the current video link exists in the map of opened videos
		var ok bool
		randomIndex := rand.Intn(len(VideoLinks))
		videoLink := VideoLinks[randomIndex]
		if openedCtx, ok = OpenedVideos[videoLink]; ok {
			// If the opened context exists, switch to it
			c := chromedp.FromContext(openedCtx)
			chromedp.Run(ctx,
				target.ActivateTarget(c.Target.TargetID),
			)
			// Play the video
			chromedp.Run(openedCtx, chromedp.Click(btPlay, chromedp.ByQuery))
		} else {
			// If the opened context does not exist, create a new one
			openedCtx, _ = chromedp.NewContext(ctx)
			chromedp.Run(openedCtx,
				chromedp.Navigate(videoLink),
			)
			// Add the new context to the map of opened videos
			OpenedVideos[videoLink] = openedCtx
		}
		// Sleep for the specified delay time before switching to the next video
		time.Sleep(VideoDelay * time.Second)
		// Play the video
		chromedp.Run(openedCtx, chromedp.Click(btPlay, chromedp.ByQuery))
	}
}

// GetLinkList returns a list of YouTube video links from either the default list
// or a user-provided file.
// If the user does not provide a file path, it returns the default list.
// Otherwise, it reads the file and returns the list of links.
// Returns a slice of strings.
func GetLinkList(Path *string) []string {
	var Result []string

	// If the path to the list is empty, return the default list
	if *Path == "" {
		Result = []string{
			"https://youtu.be/QMSdOm6ZeZ4",
			"https://youtu.be/5W7cuu4fD20",
			"https://youtu.be/mYxaliDQ9-Y",
			"https://youtu.be/CbD-l5wHQxM",
			"https://youtu.be/Rpki5McDRls",
			"https://youtu.be/mo-FmkAYSDE",
			"https://youtu.be/xIg60rjirtI",
			"https://youtu.be/j7iy9n2T8ck",
			"https://youtu.be/As77sviLpHo",
			"https://youtu.be/HwbL8wyFju8",
		}
	} else {
		// If the path to the list is not empty, read the file and split it into lines
		data, err := ioutil.ReadFile(*Path)
		if err != nil {
			panic(err)
		}
		Result = strings.Split(string(data), "\n")
	}
	return Result
}
