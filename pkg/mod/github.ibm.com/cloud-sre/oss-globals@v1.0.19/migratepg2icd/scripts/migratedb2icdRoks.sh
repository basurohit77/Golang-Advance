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
#
# Update from the original to use to migrate from ICD OSS account to ROKS
# Date: Aug-2021
# By: Alejandro Torres Rojas
# This version will need the follow environment variables
#  ROKS_SERVICE_USER     Service Credental user defined at PostgrSQL ICD
#  ROKS_SERVICE_PWD      Service Credental password defined at PostgrSQL ICD
#  ROKS_ADMIN_PWD        Admin password create at PostgrSQL ICD
#
# export ROKS_SERVICE_USER=bm_cloud_123456789_abcdd
# export ROKS_SERVICE_PWD=abdrq343adsaf3423
# export ROKS_ADMIN_PWD=1231dqwe24231sdda


me=$(basename "$0")
# Current Production/Staging database configuration
MIGRATEDB_ROOT=$(dirname $(readlink -f $0))
echo "MIGRATEDB_ROOT $MIGRATEDB_ROOT"
### NEW ACCOUNT PG TARGET SERVICE ###

PG_ICD_USER="admin"
PG_ICD_DB="ibmclouddb"
PG_ICD_DB_ROLE="ibm-cloud-base-user"
PG_CERT_PATH_OLD="$MIGRATEDB_ROOT/101afa63-91d0-11e9-a88d-5a059676d90f"
PG_CERT_PATH_NEW="$MIGRATEDB_ROOT/101afa63-91d0-11e9-a88d-5a059676d90f"

############ oss account 2117538 - OSS OSS Platfo #############
# Databases for PostgreSQL-OSS Prod  old account 2117538
PG_ICD_HOST_PRD="49c476de-4f6c-4918-8e67-97fa44369a5c.btdl8mld0r95fevivv30.databases.appdomain.cloud"
PG_ICD_DB_PORT_PRD=30387
PG_ADMIN_PWD_PRD="ecd24404fe43636adfa7d85608b21b22e83c4757014f7cc4ada"
#Databases for PostgreSQL-OSS Staging old account 1580127
PG_ICD_HOST_STG="18e41e2d-39fb-4da6-81c4-de9694e44faf.bn2a0fgd0tu045vmv2i0.databases.appdomain.cloud"
PG_ICD_DB_PORT_STG=31813
PG_ADMIN_PWD_STG=""
#Databases for PostgreSQL-OSS Staging old account 1580127
PG_ICD_HOST_DEV="08c4b6ca-c8ff-4ce5-bfd2-59904e54f3ea.b8a5e798d2d04f2e860e54e5d042c915.databases.appdomain.cloud"
PG_ICD_DB_PORT_DEV=30245
PG_ADMIN_PWD_DEV=""

############ new account ROKS 2308814 - OSSTESTDEV #############
#Databases for PostgreSQL-OSS Prd New 2117538
PG_ICD_HOST_NEW_PRD=""
PG_ICD_DB_PORT_NEW_PRD=0000
PG_ADMIN_PWD_NEW_PRD=""
#Databases for PostgreSQL-OSS Staging New 2117538
PG_ICD_HOST_NEW_STG=""
PG_ICD_DB_PORT_NEW_STG=0000
PG_ADMIN_PWD_NEW_STG=""
#Databases for PostgreSQL-OSS Dev New 2117538
PG_ICD_HOST_NEW_DEV="8ad0d8aa-0458-41f3-8017-7b274bcdba07.8117147f814b4b2ea643610826cd2046.private.databases.appdomain.cloud"
PG_ICD_DB_PORT_NEW_DEV=32680
PG_ADMIN_PWD_NEW_DEV=""

SSL_MODE="verify-full"

#PnP production
DB_NAME_PNP_PRD=pnp_prod
DB_USER_PNP_PRD=apipnp
DB_PWD_PNP_PRD=""
#subscriptionapi production
DB_NAME_SUB_PRD=subscriptionapi
DB_USER_SUB_PRD=apitip
DB_PWD_SUB_PRD=""

#PnP staging
DB_NAME_PNP_STG=pnp
DB_USER_PNP_STG=apipnp_test
DB_PWD_PNP_STG=""
#subscriptionapi_test staging
DB_NAME_SUB_STG=subscriptionapi_test
DB_USER_SUB_STG=apitip_test
DB_PWD_SUB_STG=""

#PnP development
DB_NAME_PNP_DEV=pnp_dev
DB_USER_PNP_DEV=apipnp_dev
DB_PWD_PNP_DEV=""
#subscriptionapi_dev development
DB_NAME_SUB_DEV=subscriptionapi_dev
DB_USER_SUB_DEV=subscriptionapi_dev
DB_PWD_SUB_DEV=""


timestamp=`date "+%Y%m%d%H%M"`
PG_HOME=/usr/bin/
DATA_DIR=$MIGRATEDB_ROOT/data
LOG_DIR=$MIGRATEDB_ROOT/log
logFileName=$LOG_DIR/$(basename "$0" | cut -d. -f1)_$timestamp.log
#temporary files to dump pnp and subscription databases
pnp_db=$DB_NAME_PNP_PRD.$timestamp
pnp_stg_db=$DB_NAME_PNP_STG.$timestamp
pnp_dev_db=$DB_NAME_PNP_STG.$timestamp

sub_prd_db=$DB_NAME_SUB_PRD.$timestamp
sub_stg_db=$DB_USER_SUB_STG.$timestamp
sub_dev_db=$DB_USER_SUB_STG.$timestamp

SERVICE_CREDENTAIL_ROLE=ibm-cloud-base-user
SERVICE_CREDENTAIL_USER=ibm_cloud_99323fe1_bf26_4367_8633_490aecbea400
SERVICE_CREDENTAIL_PWD=""
SERVICE_CREDENTAIL_USER_2117538=ibm_cloud_42db41b6_1d10_484a_974b_62ea10e6810c
SERVICE_CREDENTAIL_PWD_2117538=""
ROKS_ACT=2308814
OSS_ACT=2117538




helpMessage="\
Usage:
       $me -e dev 
Migrate postgresql databases from osspg1 to ICD service.

  -h, --help               Show this help information.

  -e [Required]            Source environment for the database migration [dev/stg/prd]

"

 parse_args() {
     # Parse arg flags
     while : ; do
      if [[ $1 = "-e" && -n $2  ]]; then
         target_env=$2
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

  setEnvValues $target_env
  ############################################################
  # TODO CANNOT PING FROM THE JUMPBOX NEED TO CHECK WITH KUN
  ############################################################
  # echo "Checking Connectivity to $target_env PG ICD services oss account 2117538"
  # if ping -c1 -W1 $PG_ICD_HOST 2>/dev/null; then
  #    echo "Sucessfully connected to ICD $target_env PG server $PG_ICD_HOST_PRD ✓"
  # else
  #    echo "Unable to connect to ICD $target_env PG server: $PG_ICD_HOST ✗"
  #  exit 1
  # fi
  # echo "Checking Connectivity to $target_env PG ICD services ROKS account 2308814"
  # if ping -c1 -W1 $PG_ICD_HOST_NEW 2>/dev/null; then
  #    echo "Sucessfully connected to ICD $target_env ROKS PG server $PG_ICD_HOST_PRD ✓"
  # else
  #    echo "Unable to connect to ICD PG $target_env ROKS server: $PG_ICD_HOST ✗"
  #  exit 1
  # fi

  if [ -f "$PG_CERT_PATH_OLD" ]; then
    echo "$PG_CERT_PATH_OLD ✓"
  else
    echo "Missing $PG_CERT_PATH_OLD verify the location of the ICD certificate old account 1580127 ✗"
    exit 1
  fi

  if [ -f "$PG_CERT_PATH_NEW" ]; then
    echo "$PG_CERT_PATH_NEW ✓"
  else
    echo "Missing $PG_CERT_PATH_NEW verify the location of the ICD certificate old account 2117538 ✗"
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

}

function logger {

  for var in "$@"; do
        echo -e $(date)" ${var}"
        echo -e $(date)" ${var}" >> $logFileName
  done
}



function setDB2readOnly() {
  DB_NAME=$1 ADMIN_PWD=$2 PG_CERT=$3 PG_HOST=$4 PG_PORT=$5 target_env=$6

  #ALTER DATABASE foobar SET default_transaction_read_only = true;
  logger "Setting databases $DB_NAME for $target_env environment at $PG_HOST to READ ONLY mode"
  PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_PORT dbname=$PG_ICD_DB  user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "ALTER DATABASE $DB_NAME SET default_transaction_read_only = true;"
  logger "Setting databases $DB_NAME_SUB for $target_env environment at $PG_HOST to READ ONLY mode"
  PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_PORT dbname=$PG_ICD_DB  user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "ALTER DATABASE $DB_NAME_SUB SET default_transaction_read_only = true;"
}

function setDB2Open() {
  target_env=$1

  #ALTER DATABASE foobar SET default_transaction_read_only = true;

  logger "Setting databases for $target_env environment to OPEN mode"
  # open accces first in case it is on READ ONLY already
  #PGPASSWORD=$PG_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB  user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "ALTER DATABASE $DB_NAME_PNP SET default_transaction_read_only = false;"
  #PGPASSWORD=$PG_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB  user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "ALTER DATABASE $DB_NAME_SUB SET default_transaction_read_only = false;"
  # Set to READ ONLY mode now
  PGPASSWORD=$PG_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB  user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "ALTER DATABASE $DB_NAME_PNP SET default_transaction_read_only = false;"
  PGPASSWORD=$PG_ADMIN_PWD PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$PG_ICD_DB  user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "ALTER DATABASE $DB_NAME_SUB SET default_transaction_read_only = false;"
}


# Set the values of the global variables depending of the target environment DEV/STG/PRD
function  setEnvValues() {
  target_env=$1
  logger "Setting variables for $target_env environment"

  case "$target_env" in
  "dev")
      #OSS ICD Account 2117538
      PG_ICD_HOST=$PG_ICD_HOST_DEV      
      PG_ICD_DB_PORT=$PG_ICD_DB_PORT_DEV
      PG_ADMIN_PWD=$PG_ADMIN_PWD_DEV

      #New ICD ROKS Account 2308814
      PG_ICD_HOST_NEW=$PG_ICD_HOST_NEW_DEV
      PG_ICD_DB_PORT_NEW=$PG_ICD_DB_PORT_NEW_DEV
      PG_ADMIN_PWD_NEW=$PG_ADMIN_PWD_NEW_DEV
      # database nanmes for PnP and Subciption
      DB_NAME_PNP=$DB_NAME_PNP_DEV
      DB_USER_PNP=$DB_USER_PNP_DEV
      DB_PWD_PNP=$DB_PWD_PNP_DEV
      DB_NAME_SUB=$DB_NAME_SUB_DEV
      DB_USER_SUB=$DB_USER_SUB_DEV
      DB_PWD_SUB=$DB_PWD_SUB_DEV

      ;;
  "stg" )
      #OSS ICD Account 2117538
      PG_ICD_HOST=$PG_ICD_HOST_STG
      PG_ICD_DB_PORT=$PG_ICD_DB_PORT_STG
      PG_ADMIN_PWD=$PG_ADMIN_PWD_STG

      #New ICD ROKS Account 2308814
      PG_ICD_HOST_NEW=$PG_ICD_HOST_NEW_STG
      PG_ICD_DB_PORT_NEW=$PG_ICD_DB_PORT_NEW_STG
      PG_ADMIN_PWD_NEW=$PG_ADMIN_PWD_NEW_STG
      # database nanmes for PnP and Subciption
      DB_NAME_PNP=$DB_NAME_PNP_STG
      DB_USER_PNP=$DB_USER_PNP_STG
      DB_PWD_PNP=$DB_PWD_PNP_STG
      DB_NAME_SUB=$DB_NAME_SUB_STG
      DB_USER_SUB=$DB_USER_SUB_STG
      DB_PWD_SUB=$DB_PWD_SUB_STG
      ;;
  "prd" )
      #OSS ICD Account 2117538
      PG_ICD_HOST=$PG_ICD_HOST_PRD
      PG_ICD_DB_PORT=$PG_ICD_DB_PORT_PRD
      PG_ADMIN_PWD=$PG_ADMIN_PWD_PRD

      #New ICD ROKS Account 2308814
      PG_ICD_HOST_NEW=$PG_ICD_HOST_NEW_PRD
      PG_ICD_DB_PORT_NEW=$PG_ICD_DB_PORT_NEW_PRD
      PG_ADMIN_PWD_NEW=$PG_ADMIN_PWD_NEW_PRD
      # database nanmes for PnP and Subciption
      DB_NAME_PNP=$DB_NAME_PNP_PRD
      DB_USER_PNP=$DB_USER_PNP_PRD
      DB_PWD_PNP=$DB_PWD_PNP_PRD
      DB_NAME_SUB=$DB_NAME_SUB_PRD
      DB_USER_SUB=$DB_USER_SUB_PRD
      DB_PWD_SUB=$DB_PWD_SUB_PRD
      ;;
  *)
    logger "Invalid environment use dev/stg/prd for development/staging/production environment"
    exit 1
    ;;
  esac

}

# DROP the target database if exist to create a new one with the latest data from sourcce
function dropDB() {
  DB_NAME=$1 ADMIN_PWD=$2 PG_CERT=$3 PG_HOST=$4 PG_DB_PORT=$5

  logger "Trying to drop $DB_NAME database at $PG_HOST"
  if [ "$( PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" )" = '1' ]
  then
    logger "$DB_NAME database already exist, dropping $DB_NAME database to refresh data"
    logger "Setting database $DB_NAME to READ ONLY "
    #setDB2readOnly $DB_NAME $ADMIN_PWD $PG_CERT $PG_HOST $PG_DB_PORT $target_env
    logger "  Closing any open connections "
    logger "    Revoking connection to $DB_NAME"
    PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "REVOKE CONNECT ON DATABASE $DB_NAME FROM PUBLIC;" >>$logFileName 2>&1
    logger "    Killing any open connection to $DB_NAME before to drop the database"
    PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "SELECT pg_terminate_backend (pid) FROM pg_stat_activity WHERE datname ='$DB_NAME';" >>$logFileName 2>&1
    logger "    Dropping $DB_NAME database"
    PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "DROP DATABASE $DB_NAME;" >>$logFileName 2>&1
  else
    logger "Database $DB_NAME does not exist"
  fi
}



# Create a rol if it does not exist
function createRole  {
  DB_USER=$1 ADMIN_PWD=$2 PG_CERT=$3 PG_HOST=$4 PG_DB_PORT=$5

  logger "Checking role $DB_USER at $PG_HOST"
  if [ "$( PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" -tAc "SELECT 1 FROM pg_roles WHERE rolname='$DB_USER'" )" = '1' ]
  then
    logger "Role $DB_USER already exist"
  else
    logger "Creating $DB_USER role"
    PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "CREATE ROLE $DB_USER CREATEDB LOGIN;" >>$logFileName 2>&1
    logger "Role $DB_USER created ✓"
  fi
  # Grant role to admin to create a database using admin
  logger "Granting role $DB_USER to admin"
  PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "GRANT $DB_USER to admin;" >>$logFileName 2>&1
  #logger "Granting role $PG_ICD_DB_ROLE to $DB_USER"
  #PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "GRANT '$PG_ICD_DB_ROLE' to $DB_USER;" >>$logFileName 2>&1
}

# Crete a new database with the roles if they don't exist
function createDB() {
  DB_NAME=$1  DB_USER=$2 ADMIN_PWD=$3 PG_CERT=$4 PG_HOST=$5 PG_DB_PORT=$6

  # Checks if database exist, if does, drops it
  logger "Trying to create $DB_NAME database at $PG_HOST"
  createRole $DB_USER $ADMIN_PWD $PG_CERT $PG_HOST $PG_DB_PORT
  if [ "$( PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" -tAc "SELECT 1 FROM pg_database WHERE datname='$DB_NAME'" )" = '1' ]
  then
    logger "Database $DB_NAME already exist"
    dropDB $DB_NAME
  else
    logger "Creating $DB_NAME database"
    PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$PG_ICD_DB user=$PG_ICD_USER sslmode=$SSL_MODE" '\x' -c "CREATE DATABASE $DB_NAME WITH OWNER $DB_USER;" >>$logFileName 2>&1
    logger "Database $DB_NAME created ✓"
  fi
}


## import the data to the newly created database using the data dumped from the source database
function importDB() {
   DB_NAME=$1 ADMIN_PWD=$2 PG_CERT=$3 PG_HOST=$4 PG_DB_PORT=$5 db_dump=$6 target_env=$7

   logger " Importing $db_dump  into $DB_NAME database ✓"
   PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$DB_NAME user=$PG_ICD_USER sslmode=$SSL_MODE" < $db_dump >>$logFileName 2>&1
   logger " Database $DB_NAME is imported ✓"
}

# Dumps the source database in a SQL file
function dumpDB {
  DB_NAME=$1 DB_USER=$2 ADMIN_PWD=$3 PG_CERT=$4 PG_HOST=$5 PG_DB_PORT=$6 db_dump=$7

  logger "Dumping $DB_NAME database"
  logger "  Connecting to server:$PG_HOST database:$DB_NAME user:$DB_USER"
  logger "  port: $PG_DB_PORT cert: $PG_CERT adminUsr: $PG_ICD_USER"

  PGPASSWORD=$ADMIN_PWD PGSSLROOTCERT=$PG_CERT pg_dump "port=$PG_DB_PORT host=$PG_HOST user=$PG_ICD_USER dbname=$DB_NAME sslmode=$SSL_MODE"  | sed -E 's/(DROP|CREATE|COMMENT ON) EXTENSION/-- \1 EXTENSION/g' > $db_dump
  logger "  Completed dumping $DB_NAME at $db_dump ✓"

}


function checkTableinDB() {
  DB_NAME=$1 DB_USER=$2 USER_PWD=$3 PG_CERT=$4 PG_HOST=$5 PG_DB_PORT=$6 tableName=$7

  logger "Runnig SELECT count(*) FROM $tableName"
  logger
  if (( "$(PGPASSWORD=$USER_PWD PGSSLROOTCERT=$PG_CERT psql "host=$PG_HOST port=$PG_DB_PORT dbname=$DB_NAME user=$DB_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM $tableName" )" > 0 ))
    then
    logger "Verification passed for $DB_NAME.$tableName connecting with user $DB_USER"
  else
    logger "Table $DB_NAME.$tableName is empty or unable to connect with user $DB_USER"
    exit 1
  fi
}

# Checks the healthz of the newly created DB instance
function checkDB() {
  DB_NAME=$1 DB_USER=$2 PG_OSSPG_PWD=$3 PG_CERT=$4 PG_HOST=$5 PG_DB_PORT=$6

  logger "Checking $DB_NAME healthz"
  if [[ $DB_NAME =~ "pnp" ]]; then
    logger "Checking tables were transfered to $DB_NAME database at $PG_HOST:$PG_DB_PORT"
    checkTableinDB $DB_NAME $DB_USER $PG_OSSPG_PWD $PG_CERT $PG_HOST $PG_DB_PORT incident_table
    logger "Checking $DB_NAME tables completed ✓"
  else
    logger "Checking subscriptionapi table to $DB_NAME database at $PG_HOST:$PG_DB_PORT"
    checkTableinDB $DB_NAME $DB_USER $PG_OSSPG_PWD $PG_CERT $PG_HOST $PG_DB_PORT subscriptionapi
    logger "Checking $DB_NAME tables completed ✓"
  fi
}



function importDB2ICD() {
  target_env=$1
  #Set the source DB in readonly
  # REMOVING THIS SECTION TO AVOID SET THE DB IN READONLY MODE FOR NOW
  #setDB2readOnly $DB_NAME_PNP $PG_ADMIN_PWD $PG_CERT_PATH_OLD $PG_ICD_HOST $PG_ICD_DB_PORT $target_env
  logger "importDB2ICD  $DB_NAME_PNP and $DB_NAME_SUB for $target_env environment"
  # set teporary file for exporting data
  pnp_dump=$DATA_DIR/$target_env'_'$DB_NAME_PNP'_db'_$timestamp.sql
  dumpDB $DB_NAME_PNP $DB_USER_PNP $PG_ADMIN_PWD $PG_CERT_PATH_OLD $PG_ICD_HOST $PG_ICD_DB_PORT $pnp_dump
  logger "  Importing  $DB_NAME_PNP from $pnp_dump ✓"
  dropDB $DB_NAME_PNP $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW
  createDB $DB_NAME_PNP $DB_USER_PNP $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW
  importDB $DB_NAME_PNP $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW $pnp_dump $target_env
  checkDB $DB_NAME_PNP "admin" $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW
  grantCloudUser2PublicSchema $DB_NAME_PNP
  # starting subscriptionapi workflow
  #Set the source DB in readonly
  #setDB2readOnly $DB_NAME_SUB $PG_ADMIN_PWD $PG_CERT_PATH_OLD $PG_ICD_HOST $PG_ICD_DB_PORT $target_env
  logger "  Importing  $DB_NAME_SUB from $sub_dump ✓"
  # set teporary file for exporting data
  sub_dump=$DATA_DIR/$target_env'_'$DB_NAME_SUB'_db'_$timestamp.sql
  echo $DB_NAME_SUB $DB_USER_SUB $PG_ADMIN_PWD $PG_CERT_PATH_OLD $PG_ICD_HOST $PG_ICD_DB_PORT $sub_dump
  dumpDB $DB_NAME_SUB $DB_USER_SUB $PG_ADMIN_PWD $PG_CERT_PATH_OLD $PG_ICD_HOST $PG_ICD_DB_PORT $sub_dump
  dropDB $DB_NAME_SUB $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW
  createDB $DB_NAME_SUB $DB_USER_SUB $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW
  importDB $DB_NAME_SUB $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW $sub_dump $target_env
  checkDB $DB_NAME_SUB "admin" $PG_ADMIN_PWD_NEW $PG_CERT_PATH_NEW $PG_ICD_HOST_NEW $PG_ICD_DB_PORT_NEW
  #setDB2Open $target_env
   grantCloudUser2PublicSchema $DB_NAME_SUB
}

# Need to run this grant to allow service user to access pnp and subcription tables
function grantCloudUser2PublicSchema() {
  DB_NAME=$1
  logger "Granting ALL ON SCHEMA public access to $SERVICE_CREDENTAIL_ROLE for $DB_NAME"
  PGPASSWORD=$PG_ADMIN_PWD_NEW PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME user=admin sslmode=$SSL_MODE" '\x' -c 'GRANT ALL ON SCHEMA public to "ibm-cloud-base-user";' >>$logFileName 2>&1
  logger "Granting TABLES access to $SERVICE_CREDENTAIL_ROLE  for $DB_NAME"
  PGPASSWORD=$PG_ADMIN_PWD_NEW PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME user=admin sslmode=$SSL_MODE" '\x' -c 'GRANT all ON ALL TABLES IN SCHEMA public to "ibm-cloud-base-user";' >>$logFileName 2>&1
}





function  checkTblsCounts() {
  DB_USER_PNP=$SERVICE_CREDENTAIL_USER_2117538
  DB_PWD_PNP=$SERVICE_CREDENTAIL_PWD_2117538
  ### OSS Account
  logger "DB_NAME=$DB_NAME_PNP $DB_USER=$DB_USER_PNP PG_OSSPG_PWD=$DB_PWD_PNP PG_CERT=$PG_CERT_PATH_OLD PG_HOST=$PG_ICD_HOST PG_DB_PORT=$PG_ICD_DB_PORT "
  pnpCate=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM category;" )
  pnpCata=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM catalog;" )
  pnpSub=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscription;" )
  pnpCase=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM case_table;" )
  pnpDis=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM display_names_table;" )
  pnpIndJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM incident_junction_table;" )
  pnpInd=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM incident_table;" )
  pnpManJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM maintenance_junction_table;" )
  pnpMan=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM maintenance_table;" )
  pnpNotDes=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM notification_description_table;" )
  pnpNot=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM notification_table;" )
  pnpRes=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM resource_table;" )
  pnpSub=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscription_table;" )
  pnpTagJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM tag_junction_table;" )
  pnpTag=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM tag_table;" )
  pnpVisJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE"  -tAc "SELECT count(*) FROM visibility_junction_table;" )
  pnpVis=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE"  -tAc "SELECT count(*) FROM visibility_table;" )
  pnpWatJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE"  -tAc "SELECT count(*) FROM watch_junction_table;" )
  pnpWat=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE"  -tAc "SELECT count(*) FROM watch_table;" )

  DB_USER_PNP=$SERVICE_CREDENTAIL_USER
  DB_PWD_PNP=$SERVICE_CREDENTAIL_PWD
 
  ### ROKS Account
  pnpDevCate=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM category;" )
  pnpDevCata=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM catalog;" )
  pnpDevSub=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscription;" )
  pnpDevCase=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM case_table;" )
  pnpDevDis=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM display_names_table;" )
  pnpDevIndJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM incident_junction_table;" )
  pnpDevInd=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM incident_table;" )
  pnpDevManJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM maintenance_junction_table;" )
  pnpDevMan=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM maintenance_table;" )
  pnpDevNotDes=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM notification_description_table;" )
  pnpDevNot=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM notification_table;" )
  pnpDevRes=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM resource_table;" )
  pnpDevSub=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscription_table;" )
  pnpDevTagJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM tag_junction_table;" )
  pnpDevTag=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM tag_table;" )
  pnpDevVisJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM visibility_junction_table;" )
  pnpDevVis=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM visibility_table;" )
  pnpDevWatJun=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM watch_junction_table;" )
  pnpDevWat=$(PGPASSWORD=$DB_PWD_PNP PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_PNP user=$DB_USER_PNP sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM watch_table;" )

  echo -e "\t\t\t\t Checking PnP tables \n"
  echo -e "\t Table                          Act: $ROKS_ACT  \t Act:  $OSS_ACT "
  echo -e "\t category.......................\t$pnpCate \t $pnpDevCate"
  echo -e "\t catalog........................\t$pnpCata \t $pnpDevCata"
  echo -e "\t subscription...................\t$pnpSub \t $pnpDevSub"
  echo -e "\t case_table.....................\t$pnpCase \t $pnpDevCase"
  echo -e "\t display_names_table............\t$pnpDis \t $pnpDevDis"
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


  subapi=$(PGPASSWORD=$SERVICE_CREDENTAIL_PWD_2117538 PGSSLROOTCERT=$PG_CERT_PATH_OLD psql "host=$PG_ICD_HOST port=$PG_ICD_DB_PORT dbname=$DB_NAME_SUB user=$SERVICE_CREDENTAIL_USER_2117538 sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscriptionapi;" )
  subapiICD=$(PGPASSWORD=$SERVICE_CREDENTAIL_PWD PGSSLROOTCERT=$PG_CERT_PATH_NEW psql "host=$PG_ICD_HOST_NEW port=$PG_ICD_DB_PORT_NEW dbname=$DB_NAME_SUB user=$SERVICE_CREDENTAIL_USER sslmode=$SSL_MODE" -tAc "SELECT count(*) FROM subscriptionapi;" )

  echo -e "\t\t\t\t Checking Subscription DB tables \n"
  echo -e "\t Table                          Act: $ROKS_ACT  \t Act:  $OSS_ACT "
  echo -e "\t subscriptionapi................\t$subapi \t $subapiICD"

}

main() {
  parse_args "$@"

  chekDependencies
  logger ">>>>>> STARTING ${me} FOR TARGET ENVIRONMENT ($target_env) "
  importDB2ICD $target_env
  logger "<<<<<< TARGET ENVIRONMENT ($target_env) COMPLETED"
  logger "<<<<<< CHECKING COUNTERS FOR ($target_env)"
  checkTblsCounts
  logger "<<<<<< MIGRATION COMPLETED FOR TARGET ENVIRONMENT ($target_env)"
}
[[ $1 = --source-only ]] || main "$@"
