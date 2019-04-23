package worker

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/ViBiOh/httputils/pkg/db"
	"github.com/ViBiOh/httputils/pkg/errors"
	"github.com/ViBiOh/httputils/pkg/logger"
	"github.com/ViBiOh/httputils/pkg/tools"
	"github.com/ViBiOh/iot/pkg/enedis"
	"github.com/ViBiOh/iot/pkg/provider"
)

var (
	_ provider.Worker = &App{}
)

const (
	loginURL   = "https://espace-client-connexion.enedis.fr/auth/UI/Login"
	consumeURL = "https://espace-client-particuliers.enedis.fr/group/espace-particuliers/suivi-de-consommation?"

	frenchDateFormat = "02/01/2006"
)

// Config of package
type Config struct {
	email    *string
	password *string
	timezone *string
	hour     *int
	minute   *int
}

// App of package
type App struct {
	email    string
	password string
	cookie   string

	location *time.Location
	hour     int
	minute   int

	db *sql.DB
}

// Flags adds flags for configuring package
func Flags(fs *flag.FlagSet, prefix string) Config {
	return Config{
		email:    fs.String(tools.ToCamel(fmt.Sprintf("%sEmail", prefix)), "", "[enedis] Email"),
		password: fs.String(tools.ToCamel(fmt.Sprintf("%sPassword", prefix)), "", "[enedis] Password"),
		timezone: fs.String(tools.ToCamel(fmt.Sprintf(`%sTimezone`, prefix)), `Europe/Paris`, `[enedis] Timezone`),
		hour:     fs.Int(tools.ToCamel(fmt.Sprintf(`%sSchedulerHour`, prefix)), 8, `[enedis] Scheduler hour`),
		minute:   fs.Int(tools.ToCamel(fmt.Sprintf(`%sSchedulerMinute`, prefix)), 0, `[enedis] Scheduler minute`),
	}
}

// New creates new App from Config
func New(config Config, db *sql.DB) *App {
	locationStr := strings.TrimSpace(*config.timezone)
	location, err := time.LoadLocation(locationStr)
	if err != nil {
		logger.Warn(`%+v`, errors.WithStack(err))
	}

	return &App{
		email:    strings.TrimSpace(*config.email),
		password: strings.TrimSpace(*config.password),
		db:       db,
		location: location,
		hour:     *config.hour,
		minute:   *config.minute,
	}
}

// GetSource returns source name
func (a *App) GetSource() string {
	return enedis.Source
}

func (a *App) fetchAndSaveData(ctx context.Context, date time.Time) (err error) {
	var data *enedis.Consumption

	data, err = a.GetData(ctx, date, true)
	if err != nil {
		return
	}

	var tx *sql.Tx
	if tx, err = db.GetTx(a.db, nil); err != nil {
		return
	}

	defer func() {
		if endErr := db.EndTx(tx, err); endErr != nil {
			logger.Error("%+v", endErr)
		}
	}()

	for _, value := range data.Graphe.Data {
		if err = a.saveValue(value, tx); err != nil {
			return
		}
	}

	return
}

// Start the package
func (a *App) Start() {
	if !a.Enabled() {
		logger.Warn("no config provided")
		return
	}

	if err := a.Login(); err != nil {
		logger.Error("%+v", err)
		return
	}

	lastTimestamp, err := a.getLastFetch()
	if err != nil {
		logger.Error("%+v", err)
		return
	}

	nextSync, _ := a.getNextSyncTime(a.hour, a.minute)
	lastSync := lastTimestamp.Truncate(notificationInterval).AddDate(0, 0, 1)
	for lastSync.Before(nextSync) {
		logger.Info("Fetching data for %s", lastSync.Format(frenchDateFormat))
		if err := a.fetchAndSaveData(context.Background(), lastSync); err != nil {
			logger.Error("%+v", err)
		}

		lastSync = lastSync.AddDate(0, 0, 1)
	}

	go a.startScheduler()
}

// Enabled checks if worker is enabled
func (a *App) Enabled() bool {
	return a.email != "" && a.password != ""
}
