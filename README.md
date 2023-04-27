# kc_user_list_from_db
Retrieve a list of users from a keycloak database, based on criteria passed in.

In this case it will retrieve a list of users from a realm where the user was created more than the specified days ago.

## Using It

`λ bin/kc_user_list_from_db --help` : Get the help information.


### The outcome of the help.

```bash
Usage of bin/kc_user_list_from_db:
-U, --username string       Database username. (default "keycloak21_u")
-W, --password string       Database password. (default "password")
-d, --dbname string         Database Name. (default "keycloak21_d")
-h, --host string           Specifies the host name of the machine on which the server is running. (default "localhost")
-p, --port int              Database Port. (default 5432)
-r, --realm string          Keycloak Realm (default "npe")
--days int              the number of days, after which users are deleted (default -1)
--deleteDate string     The date after which users will be deleted. Format: YYYY-MM-DD
-0, --incId                 Include id in the output
-1, --incUsername           Include username in the output (default true)
-2, --incEmail              Include email in the output
-3, --incFirstName          Include first name in the output
-4, --incLastName           Include last name in the output
-5, --incCreatedTimestamp   Include created timestamp in the output
```

### Example

`λ bin/kc_user_list_from_db --incUsername=false --realm npe  --days 2 -3 -5` : Get a list of users that were created more than 2 days ago and only include the email and created timestamp in the output.

