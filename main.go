package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/fogleman/gg"
	"github.com/google/go-github/v38/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

type Line struct {
	Title string
	Icon  string
	Value int
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No env file present")
	}

	token := os.Getenv("GH_TOKEN")
	if token == "" {
		log.Fatalln("GH_TOKEN env value not present")
	}

	//GenerateImage([]Line{
	//	{"Commits", "assets/icons/git-commit-outline.png", 1000},
	//	{"Pull Requests", "assets/icons/git-pull-request-outline.png", 122},
	//	{"Issues", "assets/icons/bug-outline.png", 205},
	//	{"Stars", "assets/icons/star-outline.png", 4566},
	//})
	//return

	commits, prs, issues, stargazzers := GetData(token)

	GenerateImage([]Line{
		{"Commits", "assets/icons/git-commit-outline.png", commits},
		{"Pull Requests", "assets/icons/git-pull-request-outline.png", prs},
		{"Issues", "assets/icons/bug-outline.png", issues},
		{"Stars", "assets/icons/star-outline.png", stargazzers},
	})
}

func GetData(token string) (int, int, int, int) {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Apparently you are %s\n", *user.Name)
	fmt.Println("Let's start counting your commits")

	var stargazzers int
	var commits int

	repos := GetRepos(ctx, client)
	for _, r := range repos {
		c := GetRepoCommitCount(ctx, client, user, r)
		//fmt.Printf("%d %s %s\n", c, strings.Repeat(" ", 8-len(fmt.Sprint(c))), *r.FullName)

		stargazzers = stargazzers + *r.StargazersCount
		commits = commits + c
	}

	fmt.Printf("Wow, these are some %d thicc commits\n", commits)

	return commits, 0, 0, stargazzers
}

func GetRepoCommitCount(ctx context.Context, client *github.Client, user *github.User, repo *github.Repository) int {
	pp := 50
	clatest, res, err := client.Repositories.ListCommits(ctx, *repo.Owner.Login, *repo.Name, &github.CommitsListOptions{
		Author: *user.Login,
		ListOptions: github.ListOptions{
			PerPage: pp,
		},
	})
	if err != nil {
		panic(err)
	}

	if res.LastPage == 0 {
		return len(clatest)
	}

	coldest, _, err := client.Repositories.ListCommits(ctx, *repo.Owner.Login, *repo.Name, &github.CommitsListOptions{
		Author: *user.Login,
		ListOptions: github.ListOptions{
			Page:    res.LastPage,
			PerPage: pp,
		},
	})
	if err != nil {
		panic(err)
	}

	return pp*(res.LastPage-1) + len(coldest)
}

func GetRepos(ctx context.Context, client *github.Client) []*github.Repository {
	var repos []*github.Repository
	var page int
	for {
		xrepos, res, err := client.Repositories.List(ctx, "", &github.RepositoryListOptions{
			Visibility: "all",
			ListOptions: github.ListOptions{
				Page: page,
			},
		})
		if err != nil {
			panic(err)
		}

		repos = append(repos, xrepos...)

		if res.NextPage == 0 {
			break
		}

		page = res.NextPage
	}

	return repos
}

func GetOrgs(ctx context.Context, client *github.Client) []*github.Organization {
	var orgs []*github.Organization
	var page int
	for {
		xorgs, res, err := client.Organizations.List(ctx, "", &github.ListOptions{Page: page})
		if err != nil {
			panic(err)
		}

		orgs = append(orgs, xorgs...)

		if res.NextPage == 0 {
			break
		}

		page = res.NextPage
	}

	return orgs
}

func GenerateImage(lines []Line) {
	// actual size: 600 x 400
	// border margins: 34
	const w = 600
	const h = 400
	const m = 34 + 20

	lh := (h - m*2) / len(lines)

	img, err := gg.LoadImage("assets/src.png")
	if err != nil {
		log.Fatal(err)
	}

	dc := gg.NewContext(600, 400)
	dc.SetRGB(0, 0, 0)

	if err := dc.LoadFontFace("assets/Inter-Bold.ttf", 35); err != nil {
		panic(err)
	}

	dc.DrawImage(img, 0, 0)

	for i, line := range lines {
		icn, err := gg.LoadImage(line.Icon)
		if err != nil {
			log.Fatal(err)
		}

		dc.DrawImage(icn, m+10, int(lh*(i+1)))

		dc.SetRGB(0, 0, 0)
		dc.DrawStringAnchored(line.Title, m+10+30+20, float64(lh*(i+1)), 0, 1)

		p := message.NewPrinter(language.English)

		dc.SetHexColor("#ef1818")
		dc.DrawStringAnchored(p.Sprintf("%d", line.Value), w-m-30, float64(lh*(i+1)), 1, 1)
	}

	dc.Clip()
	dc.SavePNG("assets/out.png")
}
