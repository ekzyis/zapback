package env

import (
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/namsral/flag"
)

var (
	Port                       int
	Env                        string
	PublicUrl                  string
	CommitShortSha             string
	CommitLongSha              string
	PhoenixdUrl                string
	PhoenixdLimitedAccessToken string
	PostgresUrl                string
	PostgresUrlWithoutPassword string
	Debug                      bool
)

func Load(filenames ...string) error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	flag.IntVar(&Port, "PORT", 4444, "Server port")
	flag.StringVar(&PublicUrl, "PUBLIC_URL", "", "Base URL")
	flag.StringVar(&PhoenixdUrl, "PHOENIXD_URL", "", "Phoenixd URL")
	flag.StringVar(&PhoenixdLimitedAccessToken, "PHOENIXD_LIMITED_ACCESS_TOKEN", "", "Phoenixd limited access token")
	flag.StringVar(&PostgresUrl, "POSTGRES_DB", "", "PostgreSQL connection URL")
	flag.StringVar(&Env, "ENV", "development", "Build environment")
	return nil
}

func Parse() {
	flag.Parse()
	Debug = Env == "development"
	CommitLongSha = execCmd("git", "rev-parse", "HEAD")
	CommitShortSha = execCmd("git", "rev-parse", "--short", "HEAD")
	if u, err := url.Parse(PostgresUrl); err == nil {
		if pw, ok := u.User.Password(); ok {
			PostgresUrlWithoutPassword = strings.Replace(PostgresUrl, fmt.Sprintf(":%s@", pw), ":*****@", 1)
		} else {
			PostgresUrlWithoutPassword = PostgresUrl
		}
	}
}

func execCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(stdout))
}
