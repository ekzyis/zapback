package env

import (
	"log"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/namsral/flag"
)

var (
	Port           int
	Env            string
	PublicUrl      string
	CommitShortSha string
	CommitLongSha  string

	PhoenixdUrl                string
	PhoenixdLimitedAccessToken string

	VoltageUrl            string
	VoltageOrganizationId string
	VoltageEnvId          string
	VoltageWalletId       string
	VoltageApiKey         string
)

func Load(filenames ...string) error {
	if err := godotenv.Load(); err != nil {
		return err
	}
	flag.IntVar(&Port, "PORT", 4444, "Server port")
	flag.StringVar(&PublicUrl, "PUBLIC_URL", "", "Base URL")

	flag.StringVar(&PhoenixdUrl, "PHOENIXD_URL", "", "Phoenixd URL")
	flag.StringVar(&PhoenixdLimitedAccessToken, "PHOENIXD_LIMITED_ACCESS_TOKEN", "", "Phoenixd limited access token")

	flag.StringVar(&VoltageUrl, "VOLTAGE_URL", "", "Voltage URL")
	flag.StringVar(&VoltageOrganizationId, "VOLTAGE_ORGANIZATION_ID", "", "Voltage organization ID")
	flag.StringVar(&VoltageEnvId, "VOLTAGE_ENV_ID", "", "Voltage environment ID")
	flag.StringVar(&VoltageWalletId, "VOLTAGE_WALLET_ID", "", "Voltage wallet ID")
	flag.StringVar(&VoltageApiKey, "VOLTAGE_API_KEY", "", "Voltage API key")

	flag.StringVar(&Env, "ENV", "development", "Build environment")

	return nil
}

func Parse() {
	flag.Parse()
	CommitLongSha = execCmd("git", "rev-parse", "HEAD")
	CommitShortSha = execCmd("git", "rev-parse", "--short", "HEAD")
}

func execCmd(name string, args ...string) string {
	cmd := exec.Command(name, args...)
	stdout, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	return strings.TrimSpace(string(stdout))
}
