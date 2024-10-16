package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"reflect"
	"strings"

	"github.com/fogleman/gg"
	"github.com/google/go-github/v38/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
	"gopkg.in/yaml.v2"
)

type Line struct {
	Title string
	Icon  string
	Value LocalizedInt
}

type LocalizedInt int

func (val LocalizedInt) PrettyPrint() string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", val)
}

const (
	STATUS_YES     = 1
	STATUS_NO      = 0
	STATUS_PENDING = 2
	STATUS_NONE    = 3
)

const (
	TYPE_PHP     = "php"
	TYPE_LARAVEL = "laravel"
)

const (
	LOGO_PHP     = "![](assets/logos/php.png)"
	LOGO_LARAVEL = "![](assets/logos/laravel.png)"
	LOGO_CHECK   = "![](assets/logos/check.png)"
	LOGO_X       = "![](assets/logos/x.png)"
	LOGO_DOTS    = "![](assets/logos/dots.png)"
)

type Showcase struct {
	Repositories []ShowcaseRepository `yaml:"repositories"`
}

type ShowcaseRepositoryType string

func (t ShowcaseRepositoryType) GetLogo() string {
	switch t {
	case TYPE_LARAVEL:
		return LOGO_LARAVEL
	case TYPE_PHP:
		return LOGO_PHP
	}
	return ""
}

type ShowcaseRepositoryStatus int

func (b ShowcaseRepositoryStatus) GetLogo() string {
	switch b {
	case STATUS_YES:
		return LOGO_CHECK
	case STATUS_NO:
		return LOGO_X
	case STATUS_PENDING:
		return LOGO_DOTS
	}
	return ""
}

type ShowcaseRepository struct {
	Title              string                   `yaml:"name"`
	Type               ShowcaseRepositoryType   `yaml:"type"`
	New                bool                     `yaml:"new"`
	MinPHP             string                   `yaml:"min_php"`
	MinLaravel         string                   `yaml:"min_laravel"`
	HasPhpStan         ShowcaseRepositoryStatus `yaml:"has_php_stan"`
	SupportsLaravelTen ShowcaseRepositoryStatus `yaml:"supports_laravel10"`
	SupportsPHPEight   ShowcaseRepositoryStatus `yaml:"supports_php8"`
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

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	client := github.NewClient(tc)

	repos := GetRepos(ctx, client)

	log.Printf("Found %d repos", len(repos))

	commits, prs, issues, stargazzers := GetGitHubStats(ctx, client, repos)

	WriteReadme(ReadmeInformation{
		GenerateReadmeRepositoriesTable(),
		commits.PrettyPrint(),
		stargazzers.PrettyPrint(),
		issues.PrettyPrint(),
		prs.PrettyPrint(),
	})

	GenerateImage([]Line{
		{"Commits", "assets/icons/git-commit-outline.png", commits},
		{"Pull Requests", "assets/icons/git-pull-request-outline.png", prs},
		{"Issues", "assets/icons/bug-outline.png", issues},
		{"Stars", "assets/icons/star-outline.png", stargazzers},
	})
}

type RTable struct {
	Rows []RTableRow
}

type RTableRow struct {
	IsHeader bool
	Cols     []string
}

type ReadmeInformation struct {
	Repositories string `replace:"repositories"`
	Commits      string `replace:"commits"`
	Stars        string `replace:"stars"`
	Issues       string `replace:"issues"`
	PullRequests string `replace:"prs"`
}

func WriteReadme(data ReadmeInformation) {
	stub, err := ioutil.ReadFile("README.stub.md")
	if err != nil {
		log.Fatalf("error reading stub file: %v", err)
	}

	content := string(stub)

	for i := 0; i < reflect.TypeOf(&data).Elem().NumField(); i++ {
		field := reflect.TypeOf(&data).Elem().Field(i)
		name := string(field.Tag.Get("replace"))
		value := reflect.Indirect(reflect.ValueOf(data)).FieldByName(field.Name).String()
		content = strings.Replace(content, fmt.Sprintf("{%s}", name), value, -1)
	}

	err = ioutil.WriteFile("README.md", []byte(content), 0644)
	if err != nil {
		log.Fatalf("error writing readme: %v", err)
	}

	fmt.Println("Wrote readme")
}

func GenerateReadmeRepositoriesTable() string {
	showcase := Showcase{}
	data, err := ioutil.ReadFile("showcase.yml")
	if err != nil {
		log.Fatalf("error reading showcase file: %v", err)
	}

	err = yaml.Unmarshal([]byte(data), &showcase)
	if err != nil {
		log.Fatalf("error parsing showcase file: %v", err)
	}

	table := RTable{
		Rows: []RTableRow{
			{
				IsHeader: true,
				Cols:     []string{"Package", "^PHP", "^Laravel", "PHPStan", "Laravel 10", "PHP 8"},
			},
			{
				IsHeader: true,
				Cols:     []string{"---", "---", "---", "---", "---", "---"},
			},
		},
	}

	for _, repo := range showcase.Repositories {
		table.Rows = append(table.Rows, RTableRow{
			Cols: []string{
				repo.GetTableTitle(),
				repo.MinPHP,
				repo.MinLaravel,
				repo.HasPhpStan.GetLogo(),
				repo.SupportsLaravelTen.GetLogo(),
				repo.SupportsPHPEight.GetLogo(),
			},
		})
	}

	lines := make([]string, len(table.Rows))
	for _, row := range table.Rows {
		var line string
		for _, lstr := range row.Cols {
			line = line + lstr + "|"
		}
		lines = append(lines, "|"+line+"\n")
	}

	return strings.Join(lines[:], "")
}

func (repo ShowcaseRepository) GetTableTitle() string {
	return fmt.Sprintf("%s [**%s**](%s)", repo.Type.GetLogo(), repo.Title, fmt.Sprintf("https://github.com/romanzipp/%s", repo.Title))
}

func (repo ShowcaseRepository) GetBooleanImageUrl(val bool) string {
	switch val {
	case true:
		return LOGO_CHECK
	case false:
		return LOGO_X
	}

	return ""
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func GetGitHubStats(ctx context.Context, client *github.Client, repos []*github.Repository) (LocalizedInt, LocalizedInt, LocalizedInt, LocalizedInt) {
	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		panic(err)
	}

	fmt.Printf("Apparently you are %s\n", *user.Name)

	var stargazzers int
	var commits int
	var issues int
	var pulls int

	searchPulls, _, err := client.Search.Issues(ctx, fmt.Sprintf("type:pr author:%s", *user.Login), &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})

	pulls = searchPulls.GetTotal()

	searchIssues, _, err := client.Search.Issues(ctx, fmt.Sprintf("type:issue author:%s", *user.Login), &github.SearchOptions{
		ListOptions: github.ListOptions{
			PerPage: 100,
		},
	})

	issues = searchIssues.GetTotal()

	fmt.Println(pulls, issues)

	for _, r := range repos {
		com := GetRepoStats(ctx, client, user, r)

		fmt.Printf("Repo stats: %s %s.. com: %d, iss: %d, prs: %d\n", *r.FullName, strings.Repeat(".", max(0, 50-len(*r.FullName))), com, iss, pul)

		stargazzers = stargazzers + *r.StargazersCount
		commits = commits + com

	}

	fmt.Printf("Counted %d commits\n", commits)

	return LocalizedInt(commits), LocalizedInt(0), LocalizedInt(issues), LocalizedInt(stargazzers)
}

func GetRepoStats(ctx context.Context, client *github.Client, user *github.User, repo *github.Repository) (int, int, int) {
	if *repo.Size == 0 {
		return 0, 0, 0
	}

	pp := 100

	var commits int
	var issues int
	var pulls int

	// ----------------------- issues -----------------------

	ilatest, ires, err := client.Issues.ListByRepo(ctx, *repo.Owner.Login, *repo.Name, &github.IssueListByRepoOptions{
		Creator: *user.Login,
		ListOptions: github.ListOptions{
			PerPage: pp,
		},
	})

	issues = len(ilatest)

	if ires.LastPage > 0 {
		ioldest, _, err := client.Issues.ListByRepo(ctx, *repo.Owner.Login, *repo.Name, &github.IssueListByRepoOptions{
			Creator: *user.Login,
			ListOptions: github.ListOptions{
				Page:    ires.LastPage,
				PerPage: pp,
			},
		})

		if err != nil {
			panic(err)
		}

		if err != nil {
			issues = pp*(ires.LastPage-1) + len(ioldest)
		}
	}

	// ----------------------- commits -----------------------

	clatest, cres, err := client.Repositories.ListCommits(ctx, *repo.Owner.Login, *repo.Name, &github.CommitsListOptions{
		Author: *user.Login,
		ListOptions: github.ListOptions{
			PerPage: pp,
		},
	})

	if err != nil {
		panic(err)
	}

	commits = len(clatest)

	// if there are more pages, check the count in the last page

	if cres.LastPage > 0 {
		coldest, _, err := client.Repositories.ListCommits(ctx, *repo.Owner.Login, *repo.Name, &github.CommitsListOptions{
			Author: *user.Login,
			ListOptions: github.ListOptions{
				Page:    cres.LastPage,
				PerPage: pp,
			},
		})

		if err != nil {
			panic(err)
		}

		commits = pp*(cres.LastPage-1) + len(coldest)
	}

	// ----------------------- pull requests -----------------------
	//
	//log.Printf("%v\n", *user)

	plastest, pres, err := client.PullRequests.List(ctx, *repo.Owner.Login, *repo.Name, &github.PullRequestListOptions{
		State: "all",
		ListOptions: github.ListOptions{
			PerPage: pp,
		},
	})

	if err != nil {
		panic(err)
	}

	pulls = len(plastest)

	if pres.LastPage > 0 {
		poldest, _, err := client.PullRequests.List(ctx, *repo.Owner.Login, *repo.Name, &github.PullRequestListOptions{
			State: "all",
			ListOptions: github.ListOptions{
				Page:    pres.LastPage,
				PerPage: pp,
			},
		})

		if err != nil {
			panic(err)
		}

		pulls = pp*(pres.LastPage-1) + len(poldest)
	}

	return commits, issues, pulls
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

		dc.SetHexColor("#ef1818")
		dc.DrawStringAnchored(line.Value.PrettyPrint(), w-m-30, float64(lh*(i+1)), 1, 1)
	}

	dc.Clip()
	dc.SavePNG("assets/out.png")
}
