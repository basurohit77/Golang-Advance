#!/usr/bin/env bash
#==============================================================================================#
# Name: opensource-scan-golang.sh                                                              #                                                   #                                                     #
# Version='3.0.0-20220225'                                                                     #
#==============================================================================================#
# The script requires one argument: the full path of the target repository
# If the list of libraries in the repository is large, the script may take some time to run.
# Exit code 9 is reserved for CI/CD pipeline
# Eg: ./open-source-scan-golang.sh -p <Path of the project>
# Eg: ./open-source-scan-golang.sh -p <Path of the project> -d (debug mode optional)

function usage {
  cat << EOF
Usage: open-source-scan-golang.sh [-p | --target-repo-path] <local target repo path> [-d | --debug]
EOF
   exit 1
}

function recreateFile {
  rm -f     "${1}"
  touch     "${1}"
  chmod 766 "${1}"
}

# Loop through arguments and process them
for arg in "$@"
do
    case $arg in
        -p|--target-repo-path)
        target_repo_path="$2"
        shift # Remove --initialize from processing
        shift # Remove --initialize from processing
        ;;
      -d|--debug)
        debug=1
        shift # Remove --initialize from processing
        shift # Remove --initialize from processing
        ;;
    esac
done

if [ "${target_repo_path}" == "" ]; then
    echo '[ERROR] Missing option -p <path_to_repo>' >&2
    usage
    exit 2
fi

startTime="$(LANG='en_US' date)"
echo "Script started at ${startTime} ..."

## GO TO TARGET DIRECTORY
rootDir="$(pwd -P)"
echo "${target_repo_path}"
cd ${target_repo_path}
if [ ${?} -ne 0 ]; then
  echo "Package path of $target_repo_path not correct... exiting"
  echo "Error: unable to access local path to target repository: ${target_repo_path}"  >> resultgo.txt
  exit 1
fi

repo_URL="$(git config --get remote.origin.url)"
orgName="$(basename $(dirname ${repo_URL}))"
pkgName="$(basename -s'.git' ${repo_URL})"
echo "Working organization: ${orgName},  Target package: ${pkgName}"

go mod tidy >/dev/null 2>&1 || {
  echo 'Error: "go mod tidy" failed.'
  exit 4
}

echo "Installing go-licenses ..."
go install github.com/google/go-licenses@latest >/dev/null 2>&1
if [ ${?} -ne 0 ]; then
  echo 'Error: failed to install "github.com/google/go-licenses".'
  exit 5
fi

gitStatus="$(git status)"
if ( grep 'go\.mod\|go\.sum' <<<"${gitStatus}" >/dev/null 2>&1 ); then
  echo "Error: Repository has changed, Git status: "
  echo "${gitStatus}"
  exit 6
fi

## Section in Main to check that declarative-deployment-tools exists in env PATH
yq --version
if [ ${?} -ne 0 ]; then
  echo 'Error: failed to retrive yq version. Installation of decelerative tools not done".'
  exit 7
fi

echo "Creating dependency list ..."
dependenciesList=$(go list -m all 2>/dev/null)
if [ ${?} -ne 0 ]; then
  echo 'Error: "go list" failed.'
  exit 8
fi

# Initializing Global Array
packageArray=()
rpackageArray=()
versionArray=()
rversionArray=()
pedigreeLicenseNameArray=()
rpedigreeLicenseNameArray=()
pedigreeLicenseArray=()
rpedigreeLicenseArray=()
githubLicenseArray=()
rgithubLicenseArray=()


# Initializing Global String
versionFile="version.csv"
tempFile="tempfile.txt"                #File used for sorting
depListFile="depListFile.txt"          #File containing target package's depName and TagNumber
apovDepListFile="apovDepListFile.txt"  #File containing already avilable depName and TagNumber in oss-open-source-approvals
#sapovDepListFile="sapovDepListFile.txt" #File with sorted apovDepListFile
pedigreeListFile="pedigreeListFile.txt" #File with target package's depName and TagNumber with Pedigree status
gitdir="oss-open-source-approvals"
gitapprovego="approved-golang-packages.md" #ApprovegoList.md file from oss-open-source-approvals
gitneedsreviewgo="golist-needs-review.md"  #NeedReviewList.md file from oss-open-source-approvals
approvedgolist="ApprovedGoList.txt"        #Working approve golist ahter filter
needsreviewgolist="NeedsReviewGolist.txt"  #Working needReviewList after filter
finalpedigreereport="FinalPedigreeReport.txt" #A short report
pedigreereport="PedigreeReport.txt"           #A detail report
needsreviewlen_start=0
needsreviewlen_end=0
Loop=1
Flag=true #Flag used for lock process
num_processes=800 #Max subprocess that can spawn

# Create Global File
recreateFile ${versionFile}
recreateFile ${tempFile}
recreateFile ${depListFile}
recreateFile ${apovDepListFile}
#recreateFile ${sapovDepListFile}
recreateFile ${pedigreeListFile}
recreateFile ${approvedgolist}
recreateFile ${needsreviewgolist}
recreateFile ${finalpedigreereport}
recreateFile ${pedigreereport}

## "1.1A Section in Main to find the Package Name and the Version Number (including null)"
function parseDependencies {
  # A function for parsing the dependency names and tag
  # from the provided list of dependencies and store in global arrays
  local depCSV
  # Call SED to set ${depCSV} to be ${1}:
  # 1. without the first line
  # 2. without any line that start with "github.ibm.com"
  # 3. replace the same dependencies that go.mod replaces
  # 4. remove the hostname prefix "github.com" from module names
  # 5. convert to single space seperated line for multiline variable 'depCSV'
  # Piping the output of the SED command
  # 1. sort the variable to unique values
  # 2. put the variable to file
  
    sed                                  \
    -e 1d                                \
    -e '/^github\.ibm\.com/d'            \
    -e 's#^.*[[:space:]]=>[[:space:]]##' \
    -e 's ^github\.com/  '               \
    -e 's#[[:space:]]v# #'               \
    <<<"${1}" |sort --unique             \
    > "${depListFile}"
}

echo "Parse the dependencies name and tag ..."
parseDependencies "${dependenciesList}"
######################################### 1.1A END #####################################

## "1.2A Section in Main to create a temp file for approved list and needs-review list"
presentDir="$(pwd -P)"
cd ..
git clone git@github.ibm.com:cloud-sre/${gitdir}.git
if [ ${?} -ne 0 ]; then
  echo "Git clone path of cloud-sre/${gitdir} not found... exiting" >> resultgo.txt
  exit 1
fi
cd ${gitdir} 2>/dev/null
cat "${gitapprovego}" > "${presentDir}"/"${approvedgolist}"
cat "${gitneedsreviewgo}" > "${presentDir}"/"${needsreviewgolist}"
cd ${presentDir} 2>/dev/null
sort --unique -o ${approvedgolist} ${approvedgolist}
sort --unique -o ${needsreviewgolist} ${needsreviewgolist}
######################################### 1.2A END #####################################

## "1.3A Section in Main to check packages within needs-review list and put them in tempfile3"
function job_Check__NeedReviewList() {  
  #Param1 is NeedReview List's dependency Name
  #Param2 is NeedReview List's Tag number
  while read -r depName depTag; do 
    if ( [[ ${depName} == ${1} ]] && [[ ${depTag} == ${2} ]] ); then
      while [ "$Loop" -eq 1 ]; do
        if ( [[ "$Flag" == "true" ]] ); then
          Flag=false
          echo "${depName} ${depTag}" >> "${apovDepListFile}"
          break
        fi
      done
    fi
  done < ./${depListFile}
}
export -f job_Check__NeedReviewList
Loop=1
Flag=true
while read -r npackName nverName restline; do
  ((i=i%num_processes)); ((i++==0)) && wait
  job_Check__NeedReviewList ${npackName} ${nverName} &
done < ./${needsreviewgolist}
wait
Loop=1
Flag=true
echo "Removing of needs-review list from go-list done"
# i=0
# while read -r packName verName; do
#    printf "%s, %-40s %-35s\n" "$i" "${packName}" "${verName}"
#    i=$((i+1))
# done < ./${apovDepListFile}
#exit
### TCD ###
######################################### 1.3A END #####################################

## "1.3B Section in Main to check packages within approve list and append them in tempfile3"
function job_Check__ApproList() {  
  #Param1 is dependency Name
  #Param2 is Tag number
  while read -r adepName adepTag restline; do
    if ( [[ ${adepName} == ${1} ]] && [[ ${adepTag} == ${2} ]] ); then
      while [ "$Loop" -eq 1 ]; do
        if ( [[ "$Flag" == "true" ]] ); then
          Flag=false
          echo "${adepName} ${adepTag}" >> "${apovDepListFile}"
          Flag=true
          break
        fi
      done
   fi
  done < ./${approvedgolist}
}
export -f job_Check__ApproList
Loop=1
Flag=true
while read -r depName tagName; do
  ((i=i%num_processes)); ((i++==0)) && wait
  job_Check__ApproList ${depName} ${tagName} &
done < ./${depListFile}

wait
Loop=1
Flag=true
echo "Removing of approve list from go-list done"
# i=0
# while read -r packName verName; do
#    printf "%s, %-40s %-35s\n" "$i" "${packName}" "${verName}"
#    i=$((i+1))
# done < ./${apovDepListFile}
# exit
### TCD ###
######################################### 1.3B END #####################################

## "1.4A Section in Main to create file for which License need to be reviewed"
sort --unique -o ${apovDepListFile} ${apovDepListFile}
grep -Fvx -f ${apovDepListFile} ${depListFile} > ${tempFile} && mv ${tempFile} ${depListFile}
echo "Reorganising the File need to be reviewed done"
# i=0
# while read -r apackName averName; do
#   printf "%s, %-40s %-35s %-25s %-40s\n" "$i" "${apackName}" "${averName}"
#   i=$((i+1))
# done < ./${depListFile}
# exit
### TCD ###
######################################### 1.4A END #####################################
############################ File creation for review done #############################
## 2.1A Section in Main to find Pedigree License
echo "Pedigree License Start"
function job_Pedigree_Lic_Check() {  
    jppackName=$1
    jpverName=$2
    pedigreeReviewed=""
    licenseName=""
    needReview=""
    pedigreeLicenseFound=""
    pedigree_yaml=$(curl -X GET "https://pedigree-service.wdc1a.cirrus.ibm.com/api/pedigree/check?name=${jppackName}&version=${jpverName}" -H "accept: application/json" -H "Content-Type: multipart/form-data" 2>/dev/null | jq '.')
    needReview=$(echo $pedigree_yaml | yq r - 'needReview')
    if ( [[ "$needReview" == "true" ]] ); then
      pedigree_yaml=""
      pedigree_yaml=$(curl -X GET "https://pedigree-service.wdc1a.cirrus.ibm.com/api/pedigree/check?name=${jppackName}&version=v${jpverName}" -H "accept: application/json" -H "Content-Type: multipart/form-data" 2>/dev/null | jq '.')
    fi  
    pedigreeReviewed=$(echo $pedigree_yaml | yq r - 'pedigreeReviewed')
    licenseName=$(echo $pedigree_yaml | yq r - 'displayLicense')
    if [[ "$licenseName" != "null" ]]; then
        y="do nothing"
    else
      licenseName="Ped Lic Not Found"
    fi

    needReview=$(echo $pedigree_yaml | yq r - 'needReview')
    if [[ "$needReview" != "null" ]]; then
      if ( [[ "$needReview" == "true" ]] ); then
        pedigreeLicenseFound="Needs-Review"
      else
        pedigreeLicenseFound="Review-Not-Require"
      fi
    else
      pedigreeLicenseFound="Review-Sec-Not-Found"
    fi
    while [ "$Loop" -eq 1 ]; do
      if ( [[ "$Flag" == "true" ]] ); then
        Flag=false
        echo "${jppackName}#${jpverName}#${licenseName}#${pedigreeLicenseFound}" >> ${pedigreeListFile}
        Flag=true
        break
      fi
     done 
}
export -f job_Pedigree_Lic_Check
Loop=1
Flag=true
while read -r depName tagName; do
  ((i=i%num_processes)); ((i++==0)) && wait
  job_Pedigree_Lic_Check ${depName} ${tagName} &
done < ./${depListFile}
wait
Loop=1
Flag=true

echo "Pedigree License check done"
sort --unique -o ${pedigreeListFile} ${pedigreeListFile} 
i=0
while IFS="#" read -r packName verName licName plicFound; do
   printf "%s, %-40s %-35s %-25s %-40s\n" "$i" "${packName}" "${verName}" "${licName}" "${plicFound}"
   i=$((i+1))
done < ./${pedigreeListFile}
#exit
### TCD ###
######################################### 2.1A END #####################################

## "3.1A Git License checkf for all selected packages"
function job_Git_Lic_Check() {
  jgpackageName=$1
  jgverName=$2
  jgplicenseName=$3
  jgplicFound=$4

  initialUrl="github.com/"
  Url=${initialUrl}${jgpackageName}
  if ( [[ "${jgpackageName}" == *".c"* ]] || [[ "${jgpackageName}" == *".o"* ]] || [[ "${jgpackageName}" == *".i"* ]] ); then
    Url=${jgpackageName}
  fi
  githubLicenseName=""
  githubLicense=""
  githubLicenseName=$(go-licenses csv "$Url" 2>/dev/null | grep "${jgpackageName}," | cut -d ',' -f 3)
  if ( [[ "$githubLicenseName" == "" ]] || [[ $githubLicenseName == null ]] ); then
    githubLicense="GIT-Lic-Not-Found"
  else
    githubLicense="${githubLicenseName}"
  fi
  echo "${jgpackageName}#${jgverName}#${jgplicenseName}#${jgplicFound}#${githubLicense}" >> "${versionFile}"  
}

while IFS="#" read -r packName verName licName licFound; do
   job_Git_Lic_Check "${packName}" "${verName}" "${licName}" "${licFound}" 
done < ./${pedigreeListFile}
echo "Git License check Done"
# i=0
# while IFS="#" read -r packName verName licName licFound gitLicName; do
#    printf "%s, %-40s %-35s %-25s %-40s %-40s\n" "$i" "${packName}" "${verName}" "${licName}" "${licFound}" "${gitLicName}"
#    i=$((i+1))
# done < ./${versionFile}
## TCD ###
######################################### 3.1A END #####################################

## "4.1A Section in Main to build array of packages with result"
packageArray=( $(cut -d"#" -f1 ${versionFile}) )
versionArray=( $(cut -d"#" -f2 ${versionFile}) )
while IFS="#" read -r packName verName licName licFound gitLicName; do
  pedigreeLicenseNameArray+=("${licName}")
  pedigreeLicenseArray+=("${licFound}")
  githubLicenseArray+=("${gitLicName}")
done < ./${versionFile}
# for i in "${!packageArray[@]}"; do
#   printf "%-40s %-35s %-40s %-40s %-18s %-26s\n" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}"
# done  
# exit
### TCD ### 
######################################### 4.1A END #####################################

## 5. Section in Main to find and eleminate the package which are Pedigree approved and put in Aproved list
## 5.1A Section in Main to Open GIT To Update ApprovedList
datestamp=`TZ=America/New_York date`
presentDir=`pwd`
cd ..
cd ${gitdir}
sed -i'' -e '$ d' ${gitapprovego}
sed -i'' -e '$ d' ${gitapprovego}
rm -rf ${gitapprovego}-e

#printf "%-40s %-35s %-25s %-40s %-40s %-18s %-10s\n" "Package" "Version" "License Review" "License in Pedigree " "License in Github" "Approved By" "Date" >> approved-golang-packages.md
#echo "-------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------" >> approved-golang-packages.md
## 5.2A Section in Main to loop to work inside GIT  #############   Review Not Require
for i in "${!packageArray[@]}"; do
  if [[ "${pedigreeLicenseArray[i]}" == "Review Not Require" ]]; then
     printf "%-40s %-35s %-40s %-40s %-18s %-26s\n" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}" "Automatic Pedigree" "${datestamp}" >> ${gitapprovego}
     rpackageArray+=("${packageArray[i]}"); rversionArray+=("${versionArray[i]}"); rpedigreeLicenseArray+=("${pedigreeLicenseArray[i]}"); rpedigreeLicenseNameArray+=("${pedigreeLicenseNameArray[i]}"); rgithubLicenseArray+=("${githubLicenseArray[i]}")
     unset packageArray[i]; unset versionArray[i]; unset pedigreeLicenseArray[i]; unset pedigreeLicenseNameArray[i]; unset githubLicenseArray[i]
  fi
done
## 5.2B Section in Main to loop to work inside GIT  #############   GIT LIC in BSD, MIT, APACHE, Mozilla
for i in "${!packageArray[@]}"; do
  if ( [[ "${githubLicenseArray[i]}" =~ "BSD" ]] || [[ "${githubLicenseArray[i]}" =~ "MIT" ]] || [[ "${githubLicenseArray[i]}" =~ "Apache" ]] || [[ "${githubLicenseArray[i]}" =~ "CC0" ]] || [[ "${githubLicenseArray[i]}" =~ "Mozilla Public License" ]] || [[ "${githubLicenseArray[i]}" =~ "MPL" ]] ); then
     printf "%-40s %-35s %-40s %-40s %-18s %-26s\n" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}" "Automatic IBM-OSI" "${datestamp}" >> ${gitapprovego}
     rpackageArray+=("${packageArray[i]}"); rversionArray+=("${versionArray[i]}"); rpedigreeLicenseArray+=("${pedigreeLicenseArray[i]}"); rpedigreeLicenseNameArray+=("${pedigreeLicenseNameArray[i]}"); rgithubLicenseArray+=("${githubLicenseArray[i]}")
     unset packageArray[i]; unset versionArray[i]; unset pedigreeLicenseArray[i]; unset pedigreeLicenseNameArray[i]; unset githubLicenseArray[i]
  fi
done
## 5.3 Section in Main to Loop Inside GIT Ends  ###############
printf "EOF" >> ${gitapprovego}
printf "\n"  >> ${gitapprovego}
printf "\`\`\`" >> ${gitapprovego}
cd ${presentDir}

#echo "After updating approve list"
#for i in "${!packageArray[@]}"; do
#  printf "%s, %-40s %-35s %-25s %-40s %-40s\n" "$i" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}"
#done
#exit
### TCD ###
############# Eleminate Pedigree Aproved Packages and put in Aproved list End ##############
## 6. Section in Main to update the golist-needs-review, which need to be taken approved later

## 6.1 Section in Main to find the essential Pacakage need to be rectified and approved and creation of temp file to put it in GIT golist.md  ##################
echo "
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
                                         Essential Packages in ApprovedGoList
----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
"
printf "%-40s %-35s %-40s %-45s %-12s\n" "Package" "Version" "License in Pedigree " "License in Github" "Package Name"
echo "-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------"
for i in "${!rpackageArray[@]}"; do
  printf "%-40s %-35s %-40s %-45s %-12s\n" "${rpackageArray[i]}" "${rversionArray[i]}" "${rpedigreeLicenseNameArray[i]}" "${rgithubLicenseArray[i]}" "${orgName}/${pkgName}"
done
echo "------------                     -----------------                           -----------------                             --------------                           ---------------           ------------------
"

echo "
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
                                          Essential Packages that require Review and Approval
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
"
printf "%-40s %-35s %-40s %-45s %-12s\n" "Package" "Version" "License in Pedigree " "License in Github" "Package Name"
echo "------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------"

modflag=0

while IFS=' ' read -r modpackName modverName; do
  if ( [[ "$modpackName" == "require" ]] ); then
    modflag=1
  elif [ "$modflag" == "1" ]; then
    if ( [[ "$modpackName" == ")" ]] ); then
      modflag=0
    else
      for i in "${!packageArray[@]}"; do
        if ( [[ "$modpackName" =~ "${packageArray[i]}" ]] ); then
          printf "%-40s %-35s %-40s %-45s %-12s\n" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}" "${orgName}/${pkgName}"
          printf "%-40s %-35s %-40s %-45s %-12s\n" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}" "${orgName}/${pkgName}" >> "${finalpedigreereport}"
        fi
      done
    fi
  fi
done < ./go.mod

printf "EOF" >> "${finalpedigreereport}"
printf "\n"  >> "${finalpedigreereport}"
printf "\`\`\`" >> "${finalpedigreereport}"
echo "------------                     -----------------                           -----------------                             --------------                           ---------------            ---------------------"

### 6.2 Section in Main to Work on Git Repository ########################
presentDir=`pwd`
cd ..
cd ${gitdir} 2>/dev/null

#To check length before commit
sed -i'' -e '$ d' ${gitneedsreviewgo}
sed -i'' -e '$ d' ${gitneedsreviewgo}
rm -rf ${gitneedsreviewgo}-e
needsreviewlen_start=`cat ./${gitneedsreviewgo} | wc -l`

cat "${presentDir}"/"${finalpedigreereport}" >> ${gitneedsreviewgo}

#To check length after commit
sed -i'' -e '$ d' ${gitneedsreviewgo}
sed -i'' -e '$ d' ${gitneedsreviewgo}
rm -rf ${gitneedsreviewgo}-e
needsreviewlen_end=`cat ./${gitneedsreviewgo} | wc -l`

cd ..
rm -rf ${gitdir} 2>/dev/null
cd ${presentDir}


## 7. Section in Main to Print the complete list to a File
printf "%-40s %-35s %-25s %-40s %-40s\n" "Package" "Version" "License Review" "License in Pedigree " "License in Github" > "${pedigreereport}"
echo "------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------" >> "${pedigreereport}"

for i in "${!packageArray[@]}"; do
  printf "%-40s %-35s %-25s %-40s %-40s\n" "${packageArray[i]}" "${versionArray[i]}" "${pedigreeLicenseArray[i]}" "${pedigreeLicenseNameArray[i]}" "${githubLicenseArray[i]}" >> "${pedigreereport}"
done

echo "
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
                                          Pedigree Report Analysis done for Golang package: $target_repo_path
-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------
"
#####################   Clean Up Temporary Files ###################
if ( [[ "$debug" != "1" ]] ); then
  cd $target_repo_path
  cp ${finalpedigreereport} /tmp/open-source-report-golang.txt
  rm -f ${versionFile}
  rm -f ${tempFile}
  rm -f ${depListFile}
  rm -f ${apovDepListFile}
  #rm -f ${sapovDepListFile}
  rm -f ${pedigreeListFile}
  rm -f ${approvedgolist}
  rm -f ${needsreviewgolist}
  rm -f ${finalpedigreereport}
  rm -f ${pedigreereport}
fi

endTime="$(LANG='en_US' date)"
echo "Script ended at ${endTime} ..."

if [[ $needsreviewlen_end -gt $needsreviewlen_start ]]; then
  newFileNumber=`expr $needsreviewlen_end - $needsreviewlen_start`
  echo "ERROR: Found $newFileNumber libraries that did not pass the Open Source Scan."
  exit 9
fi