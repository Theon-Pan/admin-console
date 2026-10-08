package cli // import "admin-console/internal/cli"
import (
	"admin-console/internal/model"
	"admin-console/internal/storage"
	"admin-console/internal/utils"
	"admin-console/internal/validator"
	"log/slog"
)

// func createAdminUserFromEnvironmentVariables(store *storage.Storage) {
// 	createAdminUser(store, config.Opts.AdminUsername(), config.Opts.AdminPassword())
// }

func createAdminUserFromInteractiveTerminal(store *storage.Storage) {
	username, password := askCredentials()
	createAdminUser(store, username, password)
}

func createAdminUser(store *storage.Storage, username, password string) {
	userCreationRequest := &model.UserCreationRequest{
		UserID:   utils.GenerateID(),
		Username: username,
		Password: password,
		IsAdmin:  true,
	}

	if store.UserExists(userCreationRequest.Username) {
		slog.Info("Skipping admin user creation because it already exists",
			slog.String("username", userCreationRequest.Username),
		)
		return
	}

	if validationErr := validator.ValidateUserCreationWithPassword(store, userCreationRequest); validationErr != nil {
		printErrorAndExit(validationErr.Error())
	}

	if user, err := store.CreateUser(userCreationRequest); err != nil {
		printErrorAndExit(err)
	} else {
		slog.Info("Create new amdin user",
			slog.String("username", user.Username),
			slog.Int64("user's system seq id", int64(user.ID)),
			slog.Int64("user's internal id", user.UserID),
		)
	}
}
