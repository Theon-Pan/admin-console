package cli // import "admin-console/internal/cli"
import (
	"admin-console/internal/config"
	"admin-console/internal/database"
	"flag"
	"fmt"
	"io"
	"os"
)

const (
	flagMigrateHelp    = "Run SQL migrations"
	flagDebugModeHelp  = "Show debug logs"
	flagConfigFileHelp = "Load configuration file"
)

// Parse parses command line arguments.
func Parse() {
	var (
		err            error
		flagMigrate    bool
		flagDebugMode  bool
		flagConfigFile string
	)

	flag.BoolVar(&flagMigrate, "migrate", false, flagMigrateHelp)
	flag.BoolVar(&flagDebugMode, "debug", false, flagDebugModeHelp)
	flag.StringVar(&flagConfigFile, "config-file", "", flagConfigFileHelp)
	flag.Parse()

	cfg := config.NewConfigParser()

	if flagConfigFile != "" {
		config.Opts, err = cfg.ParseFile(flagConfigFile)
		if err != nil {
			printErrorAndExit(err)
		}
	}

	config.Opts, err = cfg.ParseEnvironmentVariables()
	if err != nil {
		printErrorAndExit(err)
	}

	if flagDebugMode {
		config.Opts.SetLogLevel("debug")
	}

	logFile := config.Opts.LogFile()
	var logFileHandler io.Writer
	switch logFile {
	case "stderr":
		logFileHandler = os.Stderr
	case "stdout":
		logFileHandler = os.Stdout
	default:
		logFileHandler, err = os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
		if err != nil {
			printfAndExit("unable to open log file: %v", err)
		}
		defer logFileHandler.(*os.File).Close()
	}

	if err := InitializeDefaultLogger(config.Opts.LogLevel(), logFileHandler, config.Opts.LogFormat(), config.Opts.LogDateTime()); err != nil {
		printErrorAndExit(err)
	}

	db, err := database.NewConnectionPool(
		config.Opts.DatabaseURL(),
		config.Opts.DatabaseMinConns(),
		config.Opts.DatabaseMaxConns(),
		config.Opts.DatabaseConnectionLifetime(),
	)

	if err != nil {
		printfAndExit("unable to connect to database: %v", err)
	}
	defer db.Close()

	if flagMigrate {
		println("Begin to do db migration")
		if err := database.Migrate(db); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		println("Migration is done!")
	}

}

func printErrorAndExit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func printfAndExit(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	printErrorAndExit(err)
}
