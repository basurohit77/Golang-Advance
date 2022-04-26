#!/bin/bash
set -o pipefail #abort if any command fails

me=$(basename "$0")
MIGRATEDB_ROOT=$(dirname $(readlink -f $0))
timestamp=`date "+%Y%m%d%H%M"`
LOG_DIR=$MIGRATEDB_ROOT/../log

#declare -a lstChart
declare -a lstChart

VAULT_URL=https://vserv-us.sos.ibm.com:8200

helpMessage="\
Usage: $me -e [development/staging/production]  -r [ussouth/useast/eugb] -l ArraysLstChart
       $me -e development -r ussouth -l "${lstChart_arr[@]}"
Migrate postgresql databases from osspg1 to ICD service.

  -h, --help               Show this help information.

  -e, --environment [Required]  Environment [development/staging/production]
  -r, --region [Required]       Region [ussouth/useast/eugb]
  -l, --listCharts [Optional]   List of charts example declare -a OSS_CHARTS_LST=(api-pnp-status api-pnp-case api-pnp-db-cleaner api-pnp-notifications-adapter)
                                export OSS_CHARTS_LST
                                If list is not passed will check and use any chart with changes.

  Requires the following env vars to be set:	OSS_PLATFORM_API_KEY: This key allows ibmcloud login to fetch kubeconfigs.
"

parse_args() {
    # Parse arg flags
    while : ; do
      if [[  ( $1 = "-e"  || $1 = "--environment" ) && -n $2 ]]; then
        environment=$2
        shift 2
      elif [[  ( $1 = "-r"  || $1 = "--region" ) && -n $2 ]]; then
          region=$2
          shift 2
      elif [[  ( $1 = "-l"  || $1 = "--listCharts" ) && -n $2 ]]; then
        shift 1
        lstChart=("${@}")
       shift 1
      elif [[ $1 = "-h" || $1 = "--help"  ]]; then
        echo "$helpMessage"
        exit 0
      else
        break
      fi
    done
}

function chekDependencies {

   if  [  -z $environment ]; then
     echo "Environment is mandatory, use -e [development/staging/production]"
     echo "use -h for more information"
     exit 1
   fi

   if [[ "$environment" != "development" ]] && [[ "$environment" != "staging" ]] && [[ "$environment" != "production" ]] ; then
     echo "Expecting environment as one of the follow: [development/staging/production]"
     exit 1
   fi

   if  [  -z $region ]; then
     echo "Environment is mandatory, use -r [ussouth/useast/eugb]"
     echo "use -h for more information"
     exit 1
   fi

   if [[ "$region" != "ussouth" ]] && [[ "$environment" != "useast" ]] && [[ "$environment" != "eugb" ]] ; then
     echo "Expecting region as one of the follow: [development/staging/production]"
     exit 1
   fi

   if [ ${#lstChart[@]} -le 0 ]; then
     echo 'A list of charts is mandatory, use -l "${arr[@]}"'
     echo "use -h for more information"
     exit 1
   else
     # set the list of charts as env variable
     export OSS_CHARTS_LST=${lstChart[@]}
     echo "Notice: Using OSS_CHARTS_LST=$OSS_CHARTS_LST"
   fi
   # Set VAULT_ADDR
   if [ -z ${VAULT_ADDR+x} ]; then  #check varible exist and it is set
     export VAULT_ADDR=$VAULT_URL
     echo "Notice: Using VAULT_ADDR=$VAULT_ADDR"
   else
     echo "Notice: Using VAULT_ADDR value from environment ($VAULT_ADDR)"
   fi

   if [ -z ${OSS_PLATFORM_API_KEY+x} ]; then  #check varible exist and it is set
     echo "Please set OSS_PLATFORM_API_KEY first, check help for more information"
   fi

   if [ ! -f metadata.json ]; then
    echo "Missing 'metadata.json' file metadata.json file must exist under oss-charts"
    exit 1
   fi

   logFileName=$LOG_DIR/$(basename "$0" | cut -d. -f1)_$environment_$timestamp.log
 }


function getClusterInfo() {

  bxRegionName=$(jq -r "try .environments.$environment.deployments.$region.bxRegionName" metadata.json)
  bxClusterName=$(jq -r "try .environments.$environment.deployments.$region.bxClusterName" metadata.json)
  bxClusterID=$(jq -r "try .environments.$environment.deployments.$region.bxClusterID" metadata.json)

}


function deployCharts() {
# log in to bluemix to fetch kubeconfig files for the clusters
ok='no' && for i in 1 2 3 4 5; do ibmcloud login --apikey $OSS_PLATFORM_API_KEY --no-region && ok='yes' && break|| sleep 60; done
[ "$ok" == "no" ] && echo >&2 "fail retries" && exit 1

getClusterInfo

ok='no' && for i in 1 2 3 4 5; do ibmcloud target -r $bxRegionName && ok='yes' && break|| sleep 60; done
[ "$ok" == "no" ] && echo >&2 "fail retries" && exit 1
ok='no' && for i in 1 2 3 4 5; do ibmcloud ks cluster config --cluster $bxClusterID && ok='yes' && break|| sleep 60; done
[ "$ok" == "no" ] && echo >&2 "fail retries" && exit 1

# find any chart with changes
for chartName  in ${OSS_CHARTS_LST[@]}; do
  squad=$(echo $chartName | cut -d- -f1)
	logger "Found changes to chart: $chartName for squad: $squad"

  if [ -d "$chartName" ]; then

      # now we're examining a specific chart, and seeing if we can deploy it to a specific cluster
      cdEnabled=$(kdep-merge-inherited-values ./$chartName/$region-$environment-values.yaml | jq -r .continuousDeployment.enabled)
      if [ ! "$cdEnabled" == "true" ]; then
        logger "Continuous deployment not enabled, skipping this chart."
        continue
      fi

      # CD enabled for chart and cluster, now we fetch the kubeconfig
      echo "#####################################################################################"
      echo "Deploying chart $chartName to region $bxRegionName ($bxClusterName/$bxClusterID)"

      # kdep will provision secrets and deploy via helm
			logger "Trying ./$chartName/$region-$environment-values.yaml for bxRegionName:$bxRegionName and bxClusterName/ID:$bxClusterName/$bxClusterID"
      kdep -d ./$chartName/$region-$environment-values.yaml
    # done
	else
		logger "Did not find a directory $chartName"
  fi
done
}


function logger() {
  for var in "$@"; do
        echo -e $(date)" ${var}"
        echo -e $(date)" ${var}" >> $logFileName
  done
}



main()
{
  parse_args "$@"
  chekDependencies
  logger "Starting re-deployment for  environment: $environment and region: $region"
  deployCharts
  logger "Complete re-deployment"
}
[[ $1 = --source-only ]] || main "$@"
