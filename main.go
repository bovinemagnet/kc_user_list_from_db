package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"

	flag "github.com/spf13/pflag"
)

// Used for storing the results from the database.
type UserEntity struct {
	ID               string
	Email            sql.NullString
	FirstName        sql.NullString
	LastName         sql.NullString
	RealmName        sql.NullString
	Username         string
	CreatedTimestamp int64
}

const (
	DB_USERNAME     = "keycloak"
	DB_PASSWORD     = "password"
	DB_HOST         = "localhost"
	DB_NAME         = "keycloak"
	DB_PORT         = 5432
	MAX_AGE_IN_DAYS = 30
	EMPTY_DAYS      = -1
	KC_REALM        = "master"
)

// Environment constants
const (
	ENV_CLIENT_REALM = "KC_CLIENT_REALM"
	ENV_USERNAME     = "KC_DB_USERNAME"
	ENV_PASSWORD     = "KC_DB_PASSWORD"

	// Deletion on days.
	ENV_MAX_AGE_IN_DATE = "KC_MAX_AGE_IN_DATE"
	ENV_MAX_AGE_IN_DAYS = "KC_MAX_AGE_IN_DAYS"
)

var (
	username            string
	password            string
	realm               string
	dbName              string
	dbHost              string
	dbPort              int
	maxAgeInDays        int
	deleteDate          string
	incUsername         bool
	incEmail            bool
	incId               bool
	incFirstName        bool
	incLastName         bool
	incCreatedTimestamp bool
)

func main() {

	// Get the path to the executable file
	exePath, err := os.Executable()
	if err != nil {
		fmt.Println("[M]  Error:", err)
		return
	}

	// Get the name of the executable file
	exeName := filepath.Base(exePath)

	// Parse command-line arguments

	flag.NewFlagSet("Find Disconnected IDP Users in Keycloak", flag.ExitOnError)

	flag.StringVarP(&username, "username", "U", DB_USERNAME, "Database username.")
	flag.StringVarP(&password, "password", "W", DB_PASSWORD, "Database password.")
	flag.StringVarP(&dbName, "dbname", "d", DB_NAME, "Database Name.")
	flag.StringVarP(&dbHost, "host", "h", DB_HOST, "Specifies the host name of the machine on which the server is running.")
	flag.IntVarP(&dbPort, "port", "p", DB_PORT, "Database Port.")

	flag.StringVarP(&realm, "realm", "r", KC_REALM, "Keycloak Realm")

	flag.IntVar(&maxAgeInDays, "days", EMPTY_DAYS, "the number of days, after which users are deleted")
	flag.StringVar(&deleteDate, "deleteDate", "", "The date after which users will be deleted. Format: YYYY-MM-DD")

	flag.BoolVarP(&incId, "incId", "0", false, "Include id in the output")
	flag.BoolVarP(&incUsername, "incUsername", "1", true, "Include username in the output")
	flag.BoolVarP(&incEmail, "incEmail", "2", false, "Include email in the output")
	flag.BoolVarP(&incFirstName, "incFirstName", "3", false, "Include first name in the output")
	flag.BoolVarP(&incLastName, "incLastName", "4", false, "Include last name in the output")
	flag.BoolVarP(&incCreatedTimestamp, "incCreatedTimestamp", "5", false, "Include created timestamp in the output")
	flag.CommandLine.SortFlags = false
	flag.Parse()

	log.Printf("[START] %s", exeName)

	//urlExample := "postgres://username:password@localhost:5432/database_name"
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%d/%s",
		username,
		password,
		dbHost, dbPort, dbName)
	conn, err := pgx.Connect(context.Background(), dbUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// get the offset for searhing in the db.
	//var offset int64 = 0
	maxAge, err := generateEpoch(maxAgeInDays, deleteDate)
	if err != nil {
		fmt.Println("[M]  Error:", err)
		return
	}

	log.Println("[INFO]  maxAge:", maxAge, " date:", keycloakEpochToDate(maxAge))
	// Create the SQL Query to run against the keycloak database
	sqlQuery := fmt.Sprintf("select ue.id, ue.email, ue.first_name, ue.last_name,r.name as realm_name, ue.username, ue.created_timestamp FROM user_entity as UE left join REALM as R on (ue.realm_id = r.id) where r.name = '%s' AND ue.created_timestamp <= %d;", realm, maxAge)

	rows, err := conn.Query(context.Background(), sqlQuery)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var userEntities []UserEntity
	var count int = 0
	for rows.Next() {
		count++
		var userEntity UserEntity
		err := rows.Scan(
			&userEntity.ID,
			&userEntity.Email,
			&userEntity.FirstName,
			&userEntity.LastName,
			&userEntity.RealmName,
			&userEntity.Username,
			&userEntity.CreatedTimestamp,
		)
		if err != nil {
			log.Fatal(err)
		}

		userEntities = append(userEntities, userEntity)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		fmt.Println()
	}
	var newLn bool = false
	// Print the parsed objects
	for _, userEntity := range userEntities {

		if incId {
			fmt.Printf("%s", userEntity.ID)
			newLn = true
		}
		if incUsername {
			if newLn {
				fmt.Printf(",")
			}
			fmt.Printf("%s", userEntity.Username)
			newLn = true
		}
		if incEmail {
			if newLn {
				fmt.Printf(",")
			}
			fmt.Printf("%s", userEntity.Email.String)
			newLn = true
		}
		if incFirstName {
			if newLn {
				fmt.Printf(",")
			}
			fmt.Printf("%s", userEntity.FirstName.String)
			newLn = true
		}
		if incLastName {
			if newLn {
				fmt.Printf(",")
			}
			fmt.Printf("%s", userEntity.LastName.String)
			newLn = true
		}
		if incCreatedTimestamp {
			if newLn {
				fmt.Printf(",")
			}
			fmt.Printf("%d", userEntity.CreatedTimestamp)
			newLn = true
		}
		if newLn {
			fmt.Printf("\n")
			newLn = false
		}
	}
	fmt.Println()
	log.Printf("dbhost=%s  dbname=%s  realm=%s  %s=%d\n", dbHost, dbName, realm, "rows", count)
	log.Printf("[END] %s", exeName)
}

func parseEnvVariables() {

	envUsername := os.Getenv(ENV_USERNAME)
	if envUsername != "" {
		username = envUsername
	}

	envPassword := os.Getenv(ENV_PASSWORD)
	if envPassword != "" {
		password = envPassword
	}

	envClientRealm := os.Getenv(ENV_CLIENT_REALM)
	if envClientRealm != "" {
		realm = envClientRealm
	}

	var err error

	envDays := os.Getenv(ENV_MAX_AGE_IN_DAYS)
	if envDays != "" {
		maxAgeInDays, err = strconv.Atoi(envDays)
		if err != nil {
			log.Fatal("Error parsing days from env variable: ", err)
			panic("Error parsing threads from env variable: " + ENV_MAX_AGE_IN_DAYS + err.Error())
		}
	}
	logCmdLineArgs()
}

// Send the command line arguments to the log file
func logCmdLineArgs() {

	log.Println("[KeyCloak Delete via DB Tool (Day/Date Based)]")
	log.Println("  Authentication:")
	log.Println("    username:", username)
	log.Println("    password:", password)
	log.Println("    host:", dbHost)
	log.Println("    port:", dbPort)
	log.Println("    database:", dbName)
	log.Println("  Deletion Criteria")

	if maxAgeInDays > EMPTY_DAYS {
		log.Println("    maxDaysInAge:", maxAgeInDays, "[", daysToKCEpoch(maxAgeInDays), "]")
	} else {
		log.Println("    maxDaysInAge:", "disabled")
	}
	if deleteDate != "" {
		log.Println("    deleteDate:", deleteDate)
	} else {
		log.Println("    deleteDate:", "Disabled")
	}
	log.Println(" ")
}

// Convert days to a Keycloak Epoch value
func daysToKCEpoch(days int) int64 {
	return int64(days * 86400)
}

func generateEpoch(days int, date string) (int64, error) {

	if date != "" {
		t, err := time.Parse("2006-01-02", date)
		if err != nil {
			return 0, err
		}
		return t.Unix() * 1000, nil
	}
	return time.Now().AddDate(0, 0, -days).Unix() * 1000, nil
}

func keycloakEpochToDate(epoch int64) string {
	return time.Unix(epoch/1000, 0).Format("2006-01-02")
}
