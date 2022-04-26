#!/bin/bash
set -eo pipefail #abort if any command fails
# Program: migratedb.sh
# Rev: 1.0.0
# Created by: Alejandro Torres Rojas
# Date: 17-May-2020
# Developed on bash
# Tested on bash Linux osspg1 4.4.0-137-generic #163-Ubuntu SMP Mon Sep 24 13:14:43 UTC 2018 x86_64 x86_64 x86_64 GNU/Linux
#
# Update from the original to use to migrate from ICD old account to new account
# Date: 11-Dec-2020
#
# Sample program call:
# sh migratedb.sh test/null  Will run the process but not commit updates
# sh migratedb.sh commit Will run appliying all commits
#
# Arguments used:
#     flag : test or empty
#
# Will migrate Postgres databases instances from osspg1 to ICD service
# if test is passed as an argument or not argument it will run the proccess but won't commint any changes
# if commit is passed as an argument it will apply all changes to the ICD service

me=$(basename "$0")
# Current Production/Staging database configuration
MIGRATEDB_ROOT=$(dirname $(readlink -f $0))
#echo "MIGRATEDB_ROOT $MIGRATEDB_ROOT"
### NEW ACCOUNT PG TARGET SERVICE ###
# 
DB_SERVER_PRD=10.93.231.225
#doctorapitestgw
DB_SERVER_DEV=10.154.65.173
PG_ICD_USER="admin"
PG_ICD_DB="ibmclouddb"
PG_CERT_PATH="$MIGRATEDB_ROOT/fa1498a3-0bba-11ea-9a2f-deb1275e52d0"
#PGP_ADMIN_PWD=""
#PG_ICD_HOST="a2fb6b53-3b32-4721-bdff-e51cf305a50d.bn2a2uid0up8mv7mv2ig.databases.appdomain.cloud"
#PG_ICD_DB_PORT=31334
# Databases for PostgreSQL-OSS Prod
PG_ICD_HOST_PRD="aee8dfa4-5269-4f5d-a284-3036e8202a75.br37s45d0p54n73ffbr0.databases.appdomain.cloud"
PG_ICD_DB_PORT_PRD=31354
PGP_ADMIN_PWD_PRD=""
#Databases for PostgreSQL-OSS Staging
PG_ICD_HOST_STG="5c933ec1-1137-4743-a77e-e94fe90087f3.br37s45d0p54n73ffbr0.databases.appdomain.cloud"
PG_ICD_DB_PORT_STG=30436
PGP_ADMIN_PWD_STG=""
#Databases for PostgreSQL-OSS Test
#PG_ICD_HOST_DEV="a2fb6b53-3b32-4721-bdff-e51cf305a50d.bn2a2uid0up8mv7mv2ig.databases.appdomain.cloud"
#PG_ICD_DB_PORT_DEV=31334
#PGP_ADMIN_PWD_DEV=""
#Moving Dev back to Staging ICD service
PG_ICD_HOST_DEV="5c933ec1-1137-4743-a77e-e94fe90087f3.br37s45d0p54n73ffbr0.databases.appdomain.cloud"
PG_ICD_DB_PORT_DEV=30436
PGP_ADMIN_PWD_DEV=""


PG_PNP_DB_PASS="" # place holder for pwd rotation
PG_SUB_PASS="" # place holder for pwd rotation

SSL_MODE="verify-full"
#Load balancer CIS Range
PG_LB_HOST="pgtest.oss.cloud.ibm.com"
PG_LB_PORT=35432
LB_SSL_MODE="verify-ca"

# actual PG common password
PG_OSSPG_PWD=""

#kong production
DB_NAME_KONG_PRD=kong_prod
DB_USER_KONG_PRD=kong

#kong staging
DB_NAME_KONG_STG=kong
DB_USER_KONG_STG=kong_test

#kong development
DB_NAME_KONG_DEV=kong
DB_USER_KONG_DEV=doctor

#PnP production
DB_NAME_PNP_PRD=pnp_prod
DB_USER_PNP_PRD=apipnp

#PnP staging
DB_NAME_PNP_STG=pnp
DB_USER_PNP_STG=apipnp
DB_USER_NEW_PNP_STG=apipnp_test

#PnP development
DB_NAME_PNP_DEV=pnp
DB_NAME_NEW_PNP_DEV=pnp_dev
DB_USER_PNP_DEV=subscriptionapi_dev
DB_USER_NEW_PNP_DEV=apipnp_dev

#subscriptionapi production
DB_NAME_SUB_PRD=subscriptionapi
DB_USER_SUB_PRD=apitip

#subscriptionapi_test staging
DB_NAME_SUB_STG=subscriptionapi_test
DB_USER_SUB_STG=apitip
DB_USER_NEW_SUB_STG=apitip_test

#subscriptionapi_dev development
DB_NAME_SUB_DEV=subscriptionapi_dev
DB_USER_SUB_DEV=subscriptionapi_dev

timestamp=`date "+%Y%m%d%H%M"`
PG_HOME=/opt/postgresql/9.6.10/pg/bin
DATA_DIR=$MIGRATEDB_ROOT/../data
LOG_DIR=$MIGRATEDB_ROOT/../log
#logFileName=$LOG_DIR/$(basename "$0" | cut -d. -f1)_$timestamp.log
#temporary files to export tables from kong_prod database
category_table=kong_category.$timestamp
catalog_table=kong_catalog.$timestamp
subscription_table=kong_subscription.$timestamp
#temporary files to dump pnp and subscription databases
pnp_db=$DB_NAME_PNP_PRD.$timestamp
pnp_stg_db=$DB_NAME_PNP_STG.$timestamp
pnp_dev_db=$DB_NAME_PNP_STG.$timestamp

sub_prd_db=$DB_NAME_SUB_PRD.$timestamp
sub_stg_db=$DB_USER_SUB_STG.$timestamp
sub_dev_db=$DB_USER_SUB_STG.$timestamp

#Vault URI's to update secrets
VAULT_URI_DEV=/generic/crn/v1/dev/local/tip-oss-flow/global/apiplatform/pg
VAULT_URI_STG=/generic/crn/v1/staging/local/tip-oss-flow/global/apiplatform/pg
VAULT_URI_PRD=/generic/crn/v1/internal/local/tip-oss-flow/global/apiplatform/pg
#Vault config file
V_CONFIG=$DATA_DIR/.vConfig.json



helpMessage="\
Usage:
       $me -e dev -t GIT_TOKEN [Required] -c icd_cert_file
Migrate postgresql databases from osspg1 to ICD service.

  -h, --help               Show this help information.

  -e [Required]            Source environment for the database migration [dev/stg/prd]
                           Options:
                           dev     => doctorapitestgw baremetal
                           stg/prd => osspg barematal
                           The target environment will be IBM Cloud ICD pg service
  -t [Required]            Personal Git Token with access to Vault
                           to rotate pg users passwords
  -c                       ICD certificate path, if not privided wil try to use
                           fa1498a3-0bba-11ea-9a2f-deb1275e52d0 under the current location

"

 parse_args() {
     # Parse arg flags
     while : ; do
       if [[  $1 = "-t"  && -n $2 ]]; then
         GIT_TOKEN=$2
         shift 2
       elif [[ $1 = "-e" && -n $2  ]]; then
         target_env=$2
         shift 2
       elif [[ $1 = "-c" && -n $2  ]]; then
         icd_cert_path=$2
         shift 2
       elif [[ $1 = "-h" || $1 = "--help"  ]]; then
         echo "$helpMessage"
         exit 0
     else
         break
     fi
   done
 }

function chekDependencies {

  if [[  -z $target_env  ]]; then
    echo " Source environment must be provided [dev/stg/prd]"
    echo "$helpMessage"
    exit 1
  fi

  if [[  -z $GIT_TOKEN  ]]; then
    echo " GIT TOKEN is Required to write values in Vault"
    echo "$helpMessage"
    exit 1
  fi

  # check servers are alive
  if nc -z $DB_SERVER_PRD 22 2>/dev/null; then
    echo "Sucessfully connected to baremetal production pg server: $DB_SERVER_PRD ✓"
  else
    echo "Unable to connect to baremetal production pg server: $DB_SERVER_PRD ✗"
    exit 1
  fi

  if nc -z $DB_SERVER_DEV 22 2>/dev/null; then
    echo "Sucessfully connected to baremetal development pg server: $DB_SERVER_DEV ✓"
  else
    echo "Unable to connect to baremetal development pg server: $DB_SERVER_DEV ✗"
    exit 1
  fi

  if ping -c1 -W1 $PG_ICD_HOST_PRD 2>/dev/null; then
    echo "Sucessfully connected to ICD production pg server $PG_ICD_HOST_PRD ✓"
  else
    echo "Unable to connect to ICD production pg server: $PG_ICD_HOST ✗"
    exit 1
  fi

  if ping -c1 -W1 $PG_ICD_HOST_STG 2>/dev/null; then
    echo "Sucessfully connected to ICD staging/development pg server $PG_ICD_HOST_PRD ✓"
  else
    echo "Unable to connect to ICD staging/development pg server: $PG_ICD_HOST ✗"
    exit 1
  fi

  if [[ ! -z $icd_cert_path  ]]; then
      $PG_CERT_PATH = $icd_cert_path
  fi

  if [ -f "$PG_CERT_PATH" ]; then
    echo "$PG_CERT_PATH ✓"
  else
    echo "Missing $PG_CERT_PATH verify the location of the ICD certificate ✗"
    exit 1
  fi

  if [ -f "$PG_HOME/pg_dump" ]; then
    echo "$PG_HOME/pg_dump ✓"
  else
    echo "pg_dump does not exist at $PG_HOME, please check PG_HOME environment variable ✗"
    exit 1
  fi

  if [ -d "$DATA_DIR" ]; then
    echo "$DATA_DIR  exist ✓"
  else
    mkdir $DATA_DIR
    echo "Created  $DATA_DIR directory ✓"
  fi

  if [ -d "$LOG_DIR" ]; then
    echo "$LOG_DIR  exist ✓"
  else
    mkdir $LOG_DIR
    echo "Created $LOG_DIR ✓"
  fi
  logFileName=$LOG_DIR/$(basename "$0" | cut -d. -f1)_$tagert_env_$timestamp.log

  if [ -f "$V_CONFIG" ]; then
    echo "$V_CONFIG  exist ✓"
  else
    echo "Missing Vault config file:$V_CONFIG ✗"
    exit 1
  fi

  # Set VAULT_ADDR
  if [ -z ${VAULT_ADDR+x} ]; then  #check varible exist and it is set
    export VAULT_ADDR=https://vserv-us.sos.ibm.com:8200
  else
    logger "Notice: Using VAULT_ADDR value from environment ($VAULT_ADDR)"
  fi

  logger "Trying to login to Vault using token $GIT_TOKEN it most be a valid user token"
  #getting the Vault token for the GIT user to later use it to write Vault values
  vault login  -method=github token=$GIT_TOKEN >>$logFileName 2>&1
}

function logger {

  for var in "$@"; do
        echo -e $(date)" ${var}"
        echo -e $(date)" ${var}" >> $logFileName
  done
}


function paramIsNotNull() {
  #Checks parameter is not null
  var=$1 msg=$2
  if [ -z $var ]; then
    logger $msg
    exit 1
  fi

}

function  setVaulTokenForReding {
  target_env=$1
  logger "Trying to get Vault token for reading"
  # Set VAULT_TOKEN  for reading
  #if [ -z ${VAULT_TOKEN+x} ]; then
    token=$(curl --request POST --data @$DATA_DIR/dev.json $VAULT_ADDR/v1/auth/approle/login | jq ".auth.client_token")

    logger "token:$token"
    export VAULT_TOKEN=$token
    logger "Vault token was succesfuly set in the environment and ready to use ✓"
    if [ -z ${VAULT_TOKEN+x} ]; then
        logger "VAULT_TOKEN did not get exported"
        exit 1
    fi
  #else
  #   logger "Notice: Using VAULT_TOKEN:$VAULT_TOKEN value from environment"
  #fi
}

function getPGCredentials() {
  VAULT_URI=$1

  logger "Trying to retrieve pg credentials"
  #get pg_admin user
  logger "VAULT_URI=$VAULT_URI"
  #PGP_ADMIN_PWD=$(vault read  -field=value $VAULT_URI_PRD/pg_admin)
  #get pg_admin password
  #PG_ICD_USER=$(vault read  -field=value $VAULT_URI_PRD/pg_admin_pass)
  PG_OSSPG_PWD=$(vault read  -field=value $VAULT_URI/kong_db_password)
  if [ ! "$PGP_ADMIN_PWD" ] || [ ! "$PG_ICD_USER" ] || [ ! "$PG_OSSPG_PWD" ]; then
    logger "Error: Could not fetch pg credentials from Vault, values are required ✗"
    exit 1
  fi
}

function setDB2readOnly() {
  target_env=$1

  #ALTER DATABASE foobar SET default_transaction_read_only = true;

  logger "Setting databases for $target_env environment to READ ONLY mode"

  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/psql -U postgres  "host=$DB_SERVER " '\x' -c "ALTER DATABASE $DB_NAME_KONG SET default_transaction_read_only = true;"
  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/psql -U postgres  "host=$DB_SERVER " '\x' -c "ALTER DATABASE $DB_NAME_PNP SET default_transaction_read_only = true;"
  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/psql -U postgres  "host=$DB_SERVER " '\x' -c "ALTER DATABASE $DB_NAME_SUB SET default_transaction_read_only = true;"
}


function  setEnvValues() {
  target_env=$1
  logger "Setting variables for $target_env environment"

  case "$target_env" in
  "dev")
      #setVaulTokenForReding $target_env
      logger $(env |grep VAULT)
      # set default values as development
      PG_ICD_HOST=$PG_ICD_HOST_DEV
      PG_ICD_DB_PORT=$PG_ICD_DB_PORT_DEV
      PGP_ADMIN_PWD=$PGP_ADMIN_PWD_DEV
      DB_SERVER=$DB_SERVER_DEV
      DB_NAME_KONG=$DB_NAME_KONG_DEV
      DB_USER_KONG=$DB_USER_KONG_DEV
      DB_NAME_PNP=$DB_NAME_PNP_DEV
      DB_USER_PNP=$DB_USER_PNP_DEV
      DB_NAME_SUB=$DB_NAME_SUB_DEV
      DB_USER_SUB=$DB_USER_SUB_DEV
      VAULT_URI=$VAULT_URI_DEV
      #getPGCredentials $VAULT_URI
      ;;
  "stg" )
      #setVaulTokenForReding $target_env
      PG_ICD_HOST=$PG_ICD_HOST_STG
      PG_ICD_DB_PORT=$PG_ICD_DB_PORT_STG
      PGP_ADMIN_PWD=$PGP_ADMIN_PWD_STG
      DB_SERVER=$DB_SERVER_PRD  #starting and prd are located in the same bare metal
      DB_NAME_KONG=$DB_NAME_KONG_STG
      DB_USER_KONG=$DB_USER_KONG_STG
      DB_NAME_PNP=$DB_NAME_PNP_STG
      DB_USER_PNP=$DB_USER_PNP_STG
      DB_NAME_SUB=$DB_NAME_SUB_STG
      DB_USER_SUB=$DB_USER_SUB_STG
      VAULT_URI=$VAULT_URI_STG
      #getPGCredentials $VAULT_URI
      ;;
  "prd" )
      #setVaulTokenForReding $target_env
      PG_ICD_HOST=$PG_ICD_HOST_PRD
      PG_ICD_DB_PORT=$PG_ICD_DB_PORT_PRD
      PGP_ADMIN_PWD=$PGP_ADMIN_PWD_PRD
      DB_SERVER=$DB_SERVER_PRD
      DB_NAME_KONG=$DB_NAME_KONG_PRD
      DB_USER_KONG=$DB_USER_KONG_PRD
      DB_NAME_PNP=$DB_NAME_PNP_PRD
      DB_USER_PNP=$DB_USER_PNP_PRD
      DB_NAME_SUB=$DB_NAME_SUB_PRD
      DB_USER_SUB=$DB_USER_SUB_PRD
      VAULT_URI=$VAULT_URI_PRD
      #getPGCredentials $VAULT_URI
      ;;
  *)
    logger "Invalid environment use dev/stg/prd for development/staging/production environment"
    exit 1
    ;;
  esac

}


function dropDB() {
  DB_NAME=$1

  paramIsNotNull $DB_NAME "dropDB: Database name can't be empty"

  logger "Trying to drop $DB_NAME database"
  if [ "$( PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" )" = '1' ]
  then
    logger "$DB_NAME database already exist, dropping $DB_NAME database to refresh data"
    logger "  Closing any open connections "
    logger "    Revoking connection to $DB_NAME"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "REVOKE CONNECT ON DATABASE $DB_NAME FROM PUBLIC;" >>$logFileName 2>&1
    logger "    Killing any open connection to $DB_NAME before to drop the database"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "SELECT pg_terminate_backend (pid) FROM pg_stat_activity WHERE datname ='$DB_NAME';" >>$logFileName 2>&1
    logger "    Dropping $DB_NAME database"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "DROP DATABASE $DB_NAME;" >>$logFileName 2>&1
  else
    logger "Database $DB_NAME does not exist"
  fi
}


function createRole  {
  DB_USER=$1

  paramIsNotNull $DB_USER  "createRole: User name can't be empty"
  logger "Checking role $DB_USER"
  if [ "$( PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" )" = '1' ]
  then
    logger "Role $DB_USER already exist"
  else
    logger "Creating $DB_USER role"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "CREATE ROLE $DB_USER CREATEDB LOGIN;" >>$logFileName 2>&1
    logger "Role $DB_USER created ✓"
  fi
  # Grant role to admin to create a database using admin
  logger "Granting role $DB_USER to admin"
  PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "GRANT $DB_USER to admin;" >>$logFileName 2>&1
}

function createDB() {
  DB_NAME=$1  DB_USER=$2

  paramIsNotNull $DB_NAME "createDB: Database name can't be empty"
  paramIsNotNull $DB_USER "createDB: User name can't be empty"
  # Checks if database exist, if does, drops it
  logger "Trying to create $DB_NAME database"
  createRole $DB_USER
  if [ "$( PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" )" = '1' ]
  then
    logger "Database $DB_NAME already exist"
    dropDB $DB_NAME
  else
    logger "Creating $DB_NAME database"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "CREATE DATABASE $DB_NAME WITH OWNER $DB_USER;" >>$logFileName 2>&1
    logger "Database $DB_NAME created ✓"
  fi
}

function importDB() {
   DB_NAME=$1 DB_NAME_KONG=$2 catalog=$3 subscription=$4 category=$5 db_dump=$6 target_env=$7

   paramIsNotNull $DB_NAME "importDB: Database name can't be empty"
   paramIsNotNull $DB_NAME_KONG "importDB: Kong database name can't be empty"
   paramIsNotNull $catalog "importDB: File name to dump catalog is required"
   paramIsNotNull $subscription "importDB: File name to dump subscription is required"
   paramIsNotNull $category "importDB: File name to dump category is required"
   paramIsNotNull $db_dump  "importDB: File name to dump database is required"
   paramIsNotNull $target_env "importDB: target environment is required"
   logger " Importing $db_dump  into $DB_NAME database ✓"
   PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" < $db_dump >>$logFileName 2>&1
   if [[ $DB_NAME =~ "pnp" && $target_env != "prd" ]]; then
     importKongTables $DB_NAME $DB_NAME_KONG $target_env $catalog $subscription $category
   fi
   logger " Database $DB_NAME is imported ✓"
}

function  importKongTables {
  DB_NAME=$1 DB_NAME_KONG=$2 target_env=$3 catalog=$4 subscription=$5 category=$6

  paramIsNotNull $DB_NAME "importKongTables: Database name can't be empty"
  paramIsNotNull $DB_NAME_KONG "importKongTables: Kong database name can't be empty"
  paramIsNotNull $target_env "importKongTables: target environment is required"
  paramIsNotNull $catalog "importKongTables: File name to dump catalog is required"
  paramIsNotNull $subscription "importKongTables: File name to dump subscription is required"
  paramIsNotNull $category "importKongTables: File name to dump category is required"

   #Delete tables if they exist  before to iomport the ones from Kong
  #if [ "$target_env" != "dev" ]; then
    logger " Dropping $catalog $DB_NAME database if exists"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" -tAc "DROP TABLE IF EXISTS catalog CASCADE" >>$logFileName 2>&1
    logger " Dropping $subscription from $DB_NAME database if exists✓"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" -tAc "DROP TABLE IF EXISTS subscription CASCADE" >>$logFileName 2>&1
    logger " Dropping $category from $DB_NAME database  if exists ✓"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" -tAc "DROP TABLE IF EXISTS category CASCADE" >>$logFileName 2>&1

    logger " Importing $catalog from $DB_NAME_KONG into $DB_NAME database ✓"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" < $catalog >>$logFileName 2>&1
    logger " Importing $subscription from $DB_NAME_KONG into $DB_NAME database ✓"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" < $subscription >>$logFileName 2>&1
    logger " Importing $category from $DB_NAME_KONG into $DB_NAME database ✓"
    PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=verify-full" < $category >>$logFileName 2>&1
  # else
  #     logger "Table catalog already exist at $DB_NAME database in environment $target_env, table won't be uploaded assuming current data is in use"
  #     logger "Table subscription already exist at $DB_NAME database in environment $target_env, table won't be uploaded assuming current data is in use"
  #     logger "Table category already exist at $DB_NAME database in environment $target_env, table won't be uploaded assuming current data is in use"
  # fi
}

function exportKongTables {
  catalog=$1 subscription=$2 category=$3

  paramIsNotNull $catalog "exportKongTables: File name to dump catalog is required"
  paramIsNotNull $subscription "exportKongTables: File name to dump subscription is required"
  paramIsNotNull $category "exportKongTables: File name to dump category is required"
  logger "Exporting $DB_NAME_KONG tables connecting $DB_SERVER user:$DB_USER_KONG to database:$DB_NAME_KONG"
  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/pg_dump -h $DB_SERVER -U $DB_USER_KONG -d $DB_NAME_KONG -t catalog > $catalog
  logger "Catalog exported at $catalog ✓"
  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/pg_dump -h $DB_SERVER -U $DB_USER_KONG -d $DB_NAME_KONG -t subscription > $subscription
  logger "Subscription exported at $subscription ✓"
  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/pg_dump -h $DB_SERVER -U $DB_USER_KONG -d $DB_NAME_KONG -t category > $category
  logger "Category exported at $category ✓"
}

function dumpDB {
  DB_SERVER=$1 DB_NAME=$2 DB_USER=$3 db_dump=$4

  paramIsNotNull $DB_NAME "dumpDB: Database name can't be empty"
  paramIsNotNull $DB_USER  "dumpDB: User name can't be empty"
  paramIsNotNull $DB_SERVER "dumpDB: Database server can't be empty"
  paramIsNotNull $db_dump "dumpDB: File name to dump database is required"
  logger "Dumping $DB_NAME database"
  logger "  Connecting to server:$DB_SERVER database:$DB_NAME user:$DB_USER"
  PGPASSWORD=$PG_OSSPG_PWD $PG_HOME/pg_dump -h $DB_SERVER -U $DB_USER -d $DB_NAME | sed -E 's/(DROP|CREATE|COMMENT ON) EXTENSION/-- \1 EXTENSION/g' > $db_dump
  logger "  Completed dumping $DB_NAME at $db_dump ✓"

}

function changeOnweship() {
  oldOwner=$1 newOwner=$2 objFile=$3

  paramIsNotNull $oldOwner "changeOnweship: old role owner is required"
  paramIsNotNull $newOwner "changeOnweship: new role owner is required"
  paramIsNotNull $objFile "changeOnweship: target file is required"
  logger "Changing  $objFile onwership from $oldOwner to $newOwner"
  oldStr1="OWNER TO $oldOwner"
  newStr1="OWNER TO $newOwner"
  oldStr2="Owner: $oldOwner"
  newStr2="Owner: $newOwner"
  sed -i "s+$oldStr1+$newStr1+g" $objFile
  sed -i "s+$oldStr2+$newStr2+g" $objFile
  logger "Changed onwership from $oldOwner to $newOwner completed ✓"
}

function importDB2ICD() {
  #target_env=$ passed by the command line

  paramIsNotNull $target_env "importDB2ICD: target environment is required"
  setEnvValues $target_env
  #setDB2readOnly $target_env will be done manually
  logger "importDB2ICD  $DB_NAME_PNP and $DB_NAME_SUB for $target_env environment"
  # set teporary files for exporting data
  pnp_dump=$DATA_DIR/$target_env'_'$DB_NAME_PNP'_db'_$timestamp.sql
  sub_dump=$DATA_DIR/$target_env'_'$DB_NAME_SUB'_db'_$timestamp.sql
  # No need to move to PRD tables are already there
  if [ "$target_env" != "prd" ]; then
     catalog=$DATA_DIR/$target_env'_'$catalog_table.sql
     category=$DATA_DIR/$target_env'_'$category_table.sql
     subscription=$DATA_DIR/$target_env'_'$subscription_table.sql
     exportKongTables $catalog $subscription $category
  fi
  dumpDB $DB_SERVER $DB_NAME_PNP $DB_USER_PNP $pnp_dump
  logger "  Importing  $DB_NAME_PNP from $pnp_dump ✓"
  if [ "$target_env" = "dev" ]; then
     changeOnweship $DB_USER_PNP $DB_USER_NEW_PNP_DEV $pnp_dump
     DB_NAME_PNP=$DB_NAME_NEW_PNP_DEV
     DB_USER_PNP=$DB_USER_NEW_PNP_DEV
  elif [[ "$target_env" = "stg" ]]; then
     changeOnweship $DB_USER_PNP $DB_USER_NEW_PNP_STG $pnp_dump
     DB_USER_PNP=$DB_USER_NEW_PNP_STG
  fi
  if [ "$target_env" != "prd" ]; then
    changeOnweship $DB_USER_KONG $DB_USER_PNP $catalog
    changeOnweship $DB_USER_KONG $DB_USER_PNP $subscription
    changeOnweship $DB_USER_KONG $DB_USER_PNP $category
  fi
  dropDB $DB_NAME_PNP
  createDB $DB_NAME_PNP $DB_USER_PNP
  importDB $DB_NAME_PNP $DB_NAME_KONG $catalog $subscription $category $pnp_dump $target_env
  checkDB $DB_NAME_PNP $DB_USER_PNP
  updVaultDBValues $DB_NAME_PNP $DB_USER_PNP incident_table $target_env
  # checkLoadBalancer $DB_NAME_PNP $PG_ICD_USER $PGP_ADMIN_PWD incident_table We are not using Load Balancer we will connect to the DB directly
  # starting subscriptionapi workflow
  logger "  Importing  $DB_NAME_SUB from $sub_dump ✓"
  dumpDB $DB_SERVER $DB_NAME_SUB $DB_USER_SUB $sub_dump
  if [ "$target_env" = "stg" ]; then
    changeOnweship $DB_USER_SUB $DB_USER_NEW_SUB_STG $sub_dump
    DB_USER_SUB=$DB_USER_NEW_SUB_STG
  fi
  dropDB $DB_NAME_SUB
  createDB $DB_NAME_SUB $DB_USER_SUB
  importDB $DB_NAME_SUB $DB_NAME_KONG $catalog $subscription $category $sub_dump $target_env
  checkDB $DB_NAME_SUB $DB_USER_SUB
  updVaultDBValues $DB_NAME_SUB $DB_USER_SUB subscriptionapi $target_env
  # checkLoadBalancer $DB_NAME_SUB $PG_ICD_USER $PGP_ADMIN_PWD subscriptionapi We are not using Load Balancer we will connect to the DB directly
  logger "Final Vault updates from environment $target_env"
  pgHostUpd2Vault
}

function checkTableinDB() {
  DB_NAME=$1 DB_USER=$2 USER_PWD=$3 tableName=$4

  paramIsNotNull $DB_NAME "checkTableinDB: Database name can't be empty"
  paramIsNotNull $DB_USER "checkTableinDB: User name can't be empty"
  paramIsNotNull $USER_PWD "checkTableinDB: User password can't be empty"
  paramIsNotNull $tableName "checkTableinDB: Table name can't be empty"

  if (( "$(PGPASSWORD=$USER_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=verify-full" -tAc "SELECT count(*) FROM $tableName" )" > 0 ))
    then
    logger "Verification passed for $DB_NAME.$tableName connecting with user $DB_USER"
  else
    logger "Table $DB_NAME.$tableName is empty or unable to connect with user $DB_USER"
    exit 1
  fi
}


function checkDB() {
  DB_NAME=$1 DB_USER=$2

  paramIsNotNull $DB_NAME "checkDB: Database name can't be empty"
  paramIsNotNull $DB_NAME "checkDB: Database user can't be empty"

  logger "Checking $DB_NAME health"
  pg_isready_ouput=$LOG_DIR/pg_isready_$timestamp.log
  $(pg_isready -h $PG_ICD_HOST -p $PG_ICD_DB_PORT -d $DB_NAME > $pg_isready_ouput)
  if [ $? != 0 ]; then
     "There is a problem with the database ✗"
     $(cat $pg_isready_ouput)
     exit 1
  fi
  logger $(cat $pg_isready_ouput)
  logger "Setting temporary password ($PG_OSSPG_PWD) to user $DB_USER to test tables ownership"
  PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "ALTER USER $DB_USER WITH PASSWORD '$PG_OSSPG_PWD';" >>$logFileName 2>&1
  if [[ $DB_NAME =~ "pnp" ]]; then
    logger "Checking if Kong tables were transfered to $DB_NAME database"
    # checkTableinDB $DB_NAME $PG_ICD_USER $PGP_ADMIN_PWD catalog
    # checkTableinDB $DB_NAME $PG_ICD_USER $PGP_ADMIN_PWD category
    # checkTableinDB $DB_NAME $PG_ICD_USER $PGP_ADMIN_PWD incident_table
    checkTableinDB $DB_NAME $DB_USER $PG_OSSPG_PWD catalog
    checkTableinDB $DB_NAME $DB_USER $PG_OSSPG_PWD category
    checkTableinDB $DB_NAME $DB_USER $PG_OSSPG_PWD incident_table
    logger "Checking $DB_NAME tables completed ✓"
  else
    logger "Checking subscriptionapi table"
    #checkTableinDB $DB_NAME $PG_ICD_USER $PG_OSSPG_PWD subscriptionapi
    checkTableinDB $DB_NAME $DB_USER $PG_OSSPG_PWD  subscriptionapi
    logger "Checking $DB_NAME tables completed ✓"
  fi
}

function rotateDBUserPwd() {
  DB_NAME=$1 DB_USER=$2 tblName=$3

  paramIsNotNull $DB_NAME "rotateDBUserPwd: Database name can't be empty"
  paramIsNotNull $DB_USER "rotateDBUserPwd: User name can't be empty"
  paramIsNotNull $tblName "rotateDBUserPwd: Table name can't be empty"

  #Get new password
  NEW_PWD=`openssl rand -base64 15`
  if [ -n "$NEW_PWD" ]; then
      logger "Setting password: $NEW_PWD for $DB_USER"
      PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "ALTER USER $DB_USER WITH PASSWORD '$NEW_PWD';" >>$logFileName 2>&1
      logger "Testing connecting to database:$DB_NAME with user:$DB_USER query:SELECT count(*) FROM $tableName"
      if (( "$(PGPASSWORD=$NEW_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=verify-full" -tAc "SELECT count(*) FROM $tableName" )" > 0 ))
      then
        logger "New password is working for $DB_NAME.$DB_USER"
        logger "Saving new password into Vault"
        if [[ $DB_NAME =~ "pnp" ]] ; then
          logger "vault write $VAULT_URI/pg_pnp_db_pass 'value=$NEW_PWD'"
          vault write $VAULT_URI/pg_pnp_db_pass "value=$NEW_PWD"
          PG_PNP_DB_PASS=$NEW_PWD
        else
          logger "vault write $VAULT_URI/pg_sub_pass 'value=$NEW_PWD'"
          vault write $VAULT_URI/pg_sub_pass "value=$NEW_PWD"
          PG_SUB_PASS=$NEW_PWD
        fi
      else
        logger "ERROR Setting new password did not work rolling back to old password"
        NEW_PWD=$PG_OSSPG_PWD
        PGPASSWORD=$PGP_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=verify-full" '\x' -c "ALTER USER $DB_USER WITH PASSWORD '$PG_OSSPG_PWD';" >>$logFileName 2>&1
      fi
  fi
}


function  updVaultDBValues() {
  DB_NAME=$1 DB_USER=$2 tblName=$3 target_env=$4

  paramIsNotNull $DB_NAME "updVaultDBValues: Database name can't be empty"
  paramIsNotNull $DB_USER "updVaultDBValues: User name can't be empty"
  paramIsNotNull $tblName "updVaultDBValues: Table name can't be empty"
  paramIsNotNull $target_env "updVaultDBValues: Target environment can't be empty"

  rotateDBUserPwd $DB_NAME $DB_USER $tblName

  logger "Checking database name and user for updates in Vault"
  if [[ $DB_NAME =~ "pnp" ]] ; then
      if [[ $target_env != "prd" ]] ; then
        # Set the new uer name in Vault
        logger "Updating $VAULT_URI/pg_pnp_db_user with $DB_USER for database $DB_NAME at $target_env environment"
        vault write $VAULT_URI/pg_pnp_db_user "value=$DB_USER"
        logger "vault write $VAULT_URI/pg_pnp_db_user `value=$DB_USER`"
        if [[ $target_env = "dev" ]] ; then
          # Set the new database name in Vualt
          logger "Updating $VAULT_URI/pg_pnp_db with $DB_NAME for database $DB_NAME at $target_env environment"
          logger "vault write $VAULT_URI/pg_pnp_db `value=$DB_NAME`"
          vault write $VAULT_URI/pg_pnp_db "value=$DB_NAME"
        fi
      fi
  else
    if [[ $target_env = "stg" ]] ; then
      #subscriptionapi
      # Set the new uer name in Vault
      logger "Updating $VAULT_URI/pg_db_user with $DB_USER for database $DB_NAME at $target_env environment"
      logger "vault write $VAULT_URI/pg_db_user `value=$DB_USER`"
      vault write $VAULT_URI/pg_db_user "value=$DB_USER"
    fi
  fi
}

function checkLoadBalancer {
  DB_NAME=$1 DB_USER=$2 USER_PWD=$3 tblName=$4

  paramIsNotNull $DB_NAME "updVaultDBValues: Database name can't be empty"
  paramIsNotNull $DB_USER "updVaultDBValues: User name can't be empty"
  paramIsNotNull $USER_PWD "updVaultDBValues: User name can't be empty"
  paramIsNotNull $tblName "updVaultDBValues: Table name can't be empty"

  logger "Checking connection using load balancer host=$PG_LB_HOST port=$PG_LB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$LB_SSL_MODE $USER_PWD"
  #if (( "$(PGPASSWORD=$USER_PWD PGGSSENCMODE=disable PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_LB_HOST port=$PG_LB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$LB_SSL_MODE" -tAc "SELECT count(*) FROM $tblName" )" > 0 ))
  if (( "$(PGPASSWORD=$USER_PWD PGSSLROOTCERT=$PG_CERT_PATH psql "host=$PG_LB_HOST port=$PG_LB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$LB_SSL_MODE" -tAc "SELECT count(*) FROM $tblName" )" > 0 ))
    then
    logger "Verification passed for $DB_NAME.$tableName connecting with user $DB_USER using load balancer"
  else
    logger "Table $DB_NAME.$tableName is empty or unable to connect with user $DB_USER using load balancer"
    exit 1
  fi


}

function pgHostUpd2Vault {
  # Using ICD connection values instead Load Balancer
  logger "Final updates to Vault"
  # Sets the connection values using LB CIS Range
  logger "Updating $VAULT_URI/pg_host with $PG_ICD_HOST"
  vault write $VAULT_URI/pg_host "value=$PG_ICD_HOST"
  logger "vault write $VAULT_URI/pg_host `value=$PG_ICD_HOST`"
  logger "Updating $VAULT_URI/pg_port with $PG_ICD_DB_PORT"
  vault write $VAULT_URI/pg_port "value=$PG_ICD_DB_PORT"  #value for port
  logger "vault write $VAULT_URI/pg_port `value=$PG_ICD_DB_PORT`"
  logger "Updating $VAULT_URI/pg_ssl_mode with $SSL_MODE"
  vault write $VAULT_URI/pg_ssl_mode "value=$SSL_MODE"  #value for port
  logger "vault write $VAULT_URI/pg_ssl_mode `value=$SSL_MODE`"
}

function  checkTblsCounts() {

  ICD_CERT=$PG_CERT_PATH
  PG_ICD_HOST=$PG_ICD_HOST
  PG_ICD_DB_PORT=$PG_ICD_DB_PORT
  DB_NAME=$DB_NAME_PNP
  DB_USER=$DB_USER_PNP
  SSL_MODE=verify-full
  OSSPWD=Doctor4bluemix
  OSS_HOST=$DB_SERVER
  OSS_KONG=$DB_NAME_KONG
  KONG_USER=$DB_USER_KONG
  OSS_PNP_PWD=$PG_PNP_DB_PASS

  case "$target_env" in
  "dev")
      OSS_PNP=$DB_NAME_PNP_DEV
      OSS_USR=$DB_USER_PNP_DEV
      DB_NAME_SUB_OSS=$DB_USER_SUB_DEV
      DB_USER_SUB_OSS=$DB_USER_SUB_DEV
      ;;
  "stg" )
      OSS_PNP=$DB_NAME_PNP_STG
      OSS_USR=$DB_USER_PNP_STG
      DB_NAME_SUB_OSS=$DB_NAME_SUB_STG
      DB_USER_SUB_OSS=$DB_USER_SUB_STG
      ;;
  "prd" )
      OSS_PNP=$DB_NAME_PNP_PRD
      OSS_USR=$DB_USER_PNP_PRD
      DB_NAME_SUB_OSS=$DB_NAME_SUB_PRD
      DB_USER_SUB_OSS=$DB_USER_SUB_RRD
      ;;
  *)
    logger "Invalid environment use dev/stg/prd for development/staging/production environment"
    exit 1
    ;;
  esac


  if [ "$target_env" != "prd" ]; then
     kongCate=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_KONG user=$KONG_USER " -tAc "SELECT count(*) FROM category;" )
     kongCata=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_KONG user=$KONG_USER " -tAc "SELECT count(*) FROM catalog;" )
     kongSub=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_KONG user=$KONG_USER " -tAc "SELECT count(*) FROM subscription;" )
  fi
  pnpCase=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM case_table;" )
  pnpDis=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM display_names_table;" )
  pnpIndJun=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM incident_junction_table;" )
  pnpInd=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM incident_table;" )
  pnpManJun=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM maintenance_junction_table;" )
  pnpMan=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM maintenance_table;" )
  pnpNotDes=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM notification_description_table;" )
  pnpNot=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM notification_table;" )
  pnpRes=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM resource_table;" )
  pnpSub=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM subscription_table;" )
  pnpTagJun=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM tag_junction_table;" )
  pnpTag=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM tag_table;" )
  pnpVisJun=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM visibility_junction_table;" )
  pnpVis=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM visibility_table;" )
  pnpWatJun=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM watch_junction_table;" )
  pnpWat=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$OSS_PNP user=$OSS_USR " -tAc "SELECT count(*) FROM watch_table;" )

  if [ "$target_env" != "prd" ]; then
     pnpDevCate=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM category;" )
     pnpDevCata=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM catalog;" )
     pnpDevSub=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscription;" )
  fi
  pnpDevCase=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM case_table;" )
  ppnpDevDis=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM display_names_table;" )
  pnpDevIndJun=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM incident_junction_table;" )
  pnpDevInd=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM incident_table;" )
  pnpDevManJun=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM maintenance_junction_table;" )
  pnpDevMan=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM maintenance_table;" )
  pnpDevNotDes=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM notification_description_table;" )
  pnpDevNot=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM notification_table;" )
  pnpDevRes=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM resource_table;" )
  pnpDevSub=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscription_table;" )
  pnpDevTagJun=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM tag_junction_table;" )
  pnpDevTag=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM tag_table;" )
  pnpDevVisJun=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM visibility_junction_table;" )
  pnpDevVis=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM visibility_table;" )
  pnpDevWatJun=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM watch_junction_table;" )
  pnpDevWat=$(PGPASSWORD=$OSS_PNP_PWD PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM watch_table;" )

  echo -e "\t\t\t\t Checking PnP tables \n"
  echo -e "\t Table                          $OSS_KONG/$OSS_PNP \t PG_ICD_HOST "
  if [ "$target_env" != "prd" ]; then
     echo -e "\t category.......................\t$kongCate \t $pnpDevCate"
     echo -e "\t catalog........................\t$kongCata \t $pnpDevCata"
     echo -e "\t subscription...................\t$kongSub \t $pnpDevSub"
  fi
  echo -e "\t case_table.....................\t$pnpCase \t $pnpDevCase"
  echo -e "\t display_names_table............\t$pnpDis \t $ppnpDevDis"
  echo -e "\t incident_junction_table........\t$pnpIndJun \t $pnpDevIndJun"
  echo -e "\t incident_table.................\t$pnpInd \t $pnpDevInd"
  echo -e "\t maintenance_junction_table.....\t$pnpManJun \t $pnpDevManJun"
  echo -e "\t maintenance_table..............\t$pnpMan \t $pnpDevMan"
  echo -e "\t notification_description_table.\t$pnpNotDes \t $pnpDevNotDes"
  echo -e "\t notification_table.............\t$pnpNot \t $pnpDevNot"
  echo -e "\t resource_table.................\t$pnpRes \t $pnpDevRes"
  echo -e "\t subscription_table.............\t$pnpSub \t $pnpDevSub"
  echo -e "\t tag_junction_table.............\t$pnpTagJun \t $pnpDevTagJun"
  echo -e "\t tag_table......................\t$pnpTag \t $pnpDevTag"
  echo -e "\t visibility_junction_table......\t$pnpVisJun \t $pnpDevVisJun"
  echo -e "\t visibility_table...............\t$pnpVis \t $pnpDevVis"
  echo -e "\t watch_junction_table...........\t$pnpWatJun \t $pnpDevWatJun"
  echo -e "\t watch_table....................\t$pnpWat \t $pnpDevWat"
  echo -e "\n\n"

  subapi=$(PGPASSWORD=$OSSPWD psql "host=$OSS_HOST dbname=$DB_NAME_SUB_OSS user=$DB_USER_SUB_OSS " -tAc "SELECT count(*) FROM subscriptionapi;" )
  subapiICD=$(PGPASSWORD=$PG_SUB_PASS PGSSLROOTCERT=$ICD_CERT psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_SUB user=$DB_USER_SUB sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscriptionapi;" )

  echo -e "\t\t\t\t Checking Subscription DB tables \n"
  echo -e "\t Table                          $DB_NAME_SUB_OSS \t PG_ICD_HOST "
  echo -e "\t subscriptionapi................\t$subapi \t $subapiICD"

}

main() {
  parse_args "$@"

  chekDependencies
  logger ">>>>>> STARTING ${me} FOR TARGET ENVIRONMENT ($target_env) "
  importDB2ICD
  logger "<<<<<< TARGET ENVIRONMENT ($target_env) COMPLETED"
  checkTblsCounts

}
[[ $1 = --source-only ]] || main "$@"
