package cli // import "admin-console/internal/cli"
import (
	"admin-console/internal/config"
	"admin-console/internal/database"
	"admin-console/internal/redis"
	"admin-console/internal/storage"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	flagConfigFileHelp        = "Load configuration file"
	flagCreateAdminHelp       = "Create an admin user from an interactive terminal"
	flagDebugModeHelp         = "Show debug logs"
	flagMigrateHelp           = "Run SQL migrations"
	flagJWTSecretGenerateHelp = "Generate JWT secret for APIs authentication, please specify the path in which the secret is saved"
)

// Parse parses command line arguments.
func Parse() {
	var (
		err                   error
		flagConfigFile        string
		flagCreateAdmin       bool
		flagDebugMode         bool
		flagMigrate           bool
		flagJWTSecretGenerate string
	)

	flag.BoolVar(&flagCreateAdmin, "create-admin", false, flagCreateAdminHelp)
	flag.StringVar(&flagConfigFile, "config-file", "", flagConfigFileHelp)
	flag.BoolVar(&flagDebugMode, "debug", false, flagDebugModeHelp)
	flag.BoolVar(&flagMigrate, "migrate", false, flagMigrateHelp)
	flag.StringVar(&flagJWTSecretGenerate, "jwt-secret-gen", "", flagJWTSecretGenerateHelp)
	flag.Parse()

	cfg := config.NewConfigParser()

	if flagConfigFile != "" {
		config.Opts, err = cfg.ParseFile(flagConfigFile)
		if err != nil {
			printErrorAndExit(err)
		}
	}

	if flagJWTSecretGenerate != "" {
		if !filepath.IsAbs(flagJWTSecretGenerate) {
			printfAndExit("unable to write JWTKey to the path: %s , please specify an absolute path.", flagJWTSecretGenerate)
		}
		generateJWTSecret(flagJWTSecretGenerate)
		return
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
	slog.Info("Database connected.")
	defer db.Close()

	redisClient, err := redis.NewClient(
		config.Opts.RedisAddress(),
		config.Opts.RedisPassword(),
		1,
		2,
		20,
		30*time.Minute,
	)
	slog.Info("Redis/Valkey connected.")
	if err != nil {
		printfAndExit("unable to connect to Redis/Valkey: %v", err)
	}
	defer redisClient.Close()

	store := storage.NewStorage(db, redisClient)
	if err := store.Ping(); err != nil {
		printErrorAndExit(err)
	}

	if flagCreateAdmin {
		createAdminUserFromInteractiveTerminal(store)
		return
	}

	if flagMigrate {
		if err := database.Migrate(db); err != nil {
			printErrorAndExit(err)
		}
		return
	}

	startDaemon(store)

}

func printErrorAndExit(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func printfAndExit(format string, args ...any) {
	err := fmt.Errorf(format, args...)
	printErrorAndExit(err)
}
