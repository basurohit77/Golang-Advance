[![Build Status](https://wcp-cto-sre-jenkins.swg-devops.com/buildStatus/icon?job=Pipeline/api-pnp-notifications-adapter/master)](https://wcp-cto-sre-jenkins.swg-devops.com/job/Pipeline/job/api-pnp-notifications-adapter/job/master/)

# PnP Notifications Adapter
NOTICE:  This repo is now used as a library only.

Adapter used to pull Security Notifications and Announcements

# Strategy - throw out the above.

1. Use the cloudant database that the old IBM Status page is using to pick up the security notifications and announcements.  Note there are other notifications like incidents, but those are old.
2. Using the notification category IDs from those records, use the osscatalog library to look at the oss records and associate that category ID with an actual CRN service name.  If the OSS record does not have this, we need to log an error.  Then we can try the cloudant database that is manually kept by the status page team to see if we can find the service name from there.
3. If I cannot find the service name through the steps in #2, log an error and throw away the notification.
4. If I do have a service name, then try to find it in the global catalog (https://resource-catalog.bluemix.net/api/v1?languages=%2A).  If I find it, then use the display name information from there.  If I don't find it, then try to use the service name from the ossrecord (if present) or pull it from the cloudant database list.
5. I will be forced to construct the affected regions based on the notification record.  It appears this is easy since these are region IDs that are uppercase versions of what can go in the CRN.


Key points

- Security Notifications kept forever.
- Only show last 90 days of others.
- If you cannot find the notification category ID, then don't provide the record

## Steps to build and push to artifactory

1. Clone the [declarative-deployment-tools](https://github.ibm.com/cloud-sre/declarative-deployment-tools) repository and add it to your system PATH
2. Close this repository to <SOME_PATH>/src/github.ibm.com/cloud-sre
3. cd to the <SOME_PATH>/src/github.ibm.com/cloud-sre/pnp-notifications-adapter folder
4. Use the gomake command in declarative-deployment-tools to invoke a target on the shared Makefile
    For example:
    - To get dependencies: gomake dep
    - To build: gomake
    - To test: gomake test

