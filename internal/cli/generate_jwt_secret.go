package cli // import "admin-console/internal/cli"
import (
	"admin-console/internal/utils"
	"encoding/base64"
	"os"
)

func generateJWTSecret(path string) {
	secret, err := utils.GenerateJWTSecret()
	if err != nil {
		printErrorAndExit(err)
	}

	// O_EXCL prevents accidental rotation/overwrite of an existing key.
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if err != nil {
		printErrorAndExit(err)
	}
	_, writeErr := file.Write([]byte(base64.StdEncoding.EncodeToString(secret)))
	if writeErr != nil {
		printErrorAndExit(writeErr)
	}
}
