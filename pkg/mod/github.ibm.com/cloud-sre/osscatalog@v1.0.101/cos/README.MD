export endpoint=s3.us-east.cloud-object-storage.appdomain.cloud
export bucket_name="rmc-data"
export api_key=xxxx # pragma: whitelist secret
export remoteFileName="a.txt" 

access_token=`curl -X "POST" "https://iam.cloud.ibm.com/identity/token" \
     -H 'Accept: application/json' \
     -H 'Content-Type: application/x-www-form-urlencoded' \
     --data-urlencode "apikey=${api_key}" \
     --data-urlencode "response_type=cloud_iam" \
     --data-urlencode "grant_type=urn:ibm:params:oauth:grant-type:apikey" | jq -r .access_token`

curl -X "PUT" "https://$endpoint/$bucket_name/$remoteFileName" \
 -H "Authorization: bearer $access_token" \
 -H "Content-Type: text/plain" \
 -d @./parseinput.go