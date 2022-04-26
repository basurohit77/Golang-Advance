[![Build Status](https://wcp-cto-sre-jenkins.swg-devops.com/buildStatus/icon?job=Pipeline/api-pnp-abstraction/master)](https://wcp-cto-sre-jenkins.swg-devops.com/job/Pipeline/job/api-pnp-abstraction/job/master/)

# Plug-n-Play Abstraction

Plug-n-Play Abstraction is the top layer of the entire Plug-n-Play project and encompasses all the related APIs.  Within GitHub start with this project and then link to other sub-projects.

Plug-n-Play is a set of APIs that expose functions needed by other teams outside of the OneCloud Operations Platform.  Plug-n-Play is designed to provide functions with a set of general APIs that hide the underlying implementations to allow for easier integration.


## Unit tests, coverage and scanning

- Run `make test` to test.
- To view unit test coverage run `go tool cover -html=coverage.out` after running `make test`.
- You should aim to get at least 80% coverage for each package.
- Run `make scan` to run a security scan.
- Unit tests and scan should be successful before submitting a pull request to the master branch.
- You can find results of unit tests in unittest.out.

## CI/CD

- The jenkins job will run unit tests and a security scan on pull requests and pulls.
- Upon successful merge to master, the list of jobs in config-jenkins.yaml will be triggered to rebuild the dependant components.


## Terms

- *IBM OneCloud:* A common set of offering, architecture, operations, security, and compliance standards and process to enable IBM, across all brands, to realize the mission of One Cloud in support of One IBM


- *IBM OneCloud Service:* A standard service currently offered through the IBM Cloud or Watson platform and which is part of the IBM Cloud business unit

- *OneCloud Operations Platform:* More commonly referred to as the Operational Support System ("OSS") whose mission is to ensure that all Watson and Cloud platform services adopt the OSS platform to provide a consistent experience for our customers by utilizing a common integrated Service Management toolset.

- *Non-cloud service team:* Is referring to any team outside the IBM Watson & Cloud business unit that has a need for an integrated data source of operational data coming from the IBM Watson & Cloud services and offerings.  Managed service providers such as GTS, GBS, SaaS, etc. are a subset of these teams.

- *GTS, "Global Technology Services"*: who provide managed services across a portfolio of products and services including Watson & Cloud offering. Often then are managing IBM Cloud services and accounts on behalf of external clients and use our Cloud client portals to manage those services.  They also will have their own offerings that rely on IBM Cloud IaaS ans PaaS to deliver those offerings.

- *GBS, Global Business Services:*" similar to GTS but includes more of a consulting practice and project based implementation services for application development.  They also provide managed services on top of the IBM cloud offerings and have their own offerings that run on IBM cloud infrastructure and services.


## Objectives of Plug-n-Play:
- Provide a more cohesive, integrated model for enabling client reporting of IBM OneCloud service issues, and for receiving information on the state of the service they are using. 
- Provide essential data from IBM OneCloud IaaS & PaaS needed by end users and non-cloud Service teams who have their own management environments for their applications and deployments in IBM OneCloud. 
  - i.e. notifications, service status itself (up/down), etc. 
- Provide isolation between the OneCloud Operations Platform tools used to manage IBM OneCloud services and tools being used by other organizations, in order to protect those end user portals from changes in the IBM OneCloud Operations Platform tooling providers.


## Approach for implementing Plug-n-Play:
- Provide an API allowing business units other than IBM OneCloud (GTS, GBS, SaaS, etc.) to integrate with our OneCloud Operations Platform and enable visibility into the cloud operating environments. 

## Overall assumptions for the Plug-n-Play API:   
- IBM non-cloud Service teams will maintain primary and sole responsibility for client communications using the data pulled from OneCloud Operations Platform.  
- Data required by external business units will be pulled from the OneCloud Operations Platform thru the Plug-n-Play APIs 
- Only minimum required data needed for identification, authentication, and authorization of a user will be pushed into the OneCloud Operations Platform
- No Private Information which can identify an external end user client (contact#, email, name, etc) can be passed from an end user client portal thru the Plug-n-Play API
- Each time an IBM non-cloud Service team brings in a new client with cloud service, or expands existing cloud services, they will need to perform some type of “onboarding” to ensure all the appropriate entitlement data is passing through the Plug-n-Play API appropriately. 
- There is a requirement to have the identity of the reporter, unique account identifier, and assets identifier.
- Neither Entitlement or external access to TIP will be in scope of the Minimum Viable Product (“MVP”)
- We are building this Plug-n-Play API assuming it will be eventually offered to external clients in later releases 

## MVP Requirements

Use Cases for inclusion in Minimum Viable Product (“MVP”)and  Initial Design Consideration:

### 1) Case Management
-  The capability to view, update/append comments, and accept resolution of a case prior to closing for any service deployed on the OneCloud platform via their own service portal 
- The capability to view, update/append comments, and accept resolution for a case originating from the OneCloud client portal and/or ServiceNow Portal (i.e. opened somewhere else besides the API Abstraction layer) 
     - IaaS cases which are not linked accounts and originate in IMS will not be available for MVP 
- Support for copying SalesForce cases and receiving case updates for support team using SalesForce portal (i.e. SaaS and Hybrid Cloud teams)
     - This is enabled once the Salesforce to ServiceNow bridge is completed and Salesforce cases are contained in ServiceNow
- Regardless of the portal used by the non-cloud Service team, the capability will be sych'd across the various portals supported 

#### Scope Definition:
For MVP we will provide only cases that originate from ServiceNow. So this means in the short term we will only be able to provide PaaS cases. In the future when the new function is available, we will also be able to provide IaaS cases for linked accounts only. We will not be able to provide cases for IaaS only customers.  
Current examples of MVP scope are:
  - PaaS and Watson services onboarded into ServiceNow. (This includes services who have not yet boarded into TIP but are managed in ServiceNow)
  - Linked accounts requesting PaaS support that originates in the IMS system and then bridged from IMS to ServiceNow for support handling 
  - Cases that are opened in an IBM Cloud portal or console and/or directly in ServiceNow
  - Salesforce cases that enter into ServiceNow thorough the Salesforce-ServiceNow Bridge once that is completed
  - Net new OneCloud data center clients when implemented whose cases will originate in ServiceNow 

A more detailed explanation of why scope is defined that way can be found here: [PnP Case: Issue # 5](https://github.ibm.com/cloud-sre/pnp-case/issues/5)

#### Sample Process Flow:
- The IBM non-cloud Service team support person receives ticket in their IPC system, triages ticket & determines if there is a need to open a case with IBM Cloud, then initiates the opening of the case through automation that leverages the API abstraction layer and opens a case in the IBM OneCloud ServiceNow system.   
- A check will be made to ensure the service being reported adheres to the scope definition and is eligible for case creation within the ServiceNow system (i.e. PaaS tickets only in MVP).  If not then a message will be given to the requestor to open up the case outside the API Abstraction layer and directly in the IaaS portal (console.bluemix.net) 
- Once the case initiation request is verified then a Ticket copy and synchronization will occur and support team assignment will be made. The original ticket will remain the responsibility of the non-cloud Service team and stay in their systems queue to ensure ownership does not pass to an IBM Cloud support team
- Ticket handling, problem resolution, & communication will progress using the standard OneCloud support processes and procedures. 
- All communication back to the case originator will be done through updates in the case record, and made available programmatically through the API abstraction layer (See Case Notifications Use Case).

#### Assumptions and Dependencies
- For MVP the initial entitlement will be performed in the non-cloud Service team system, including any cloud premium support services.  In later releases we'll possibly use the BSS API's as they are developed and perform entitlement as non-cloud service teams access the API Abstraction layer.    
- As tickets get copied into ServiceNow from the Abstraction layer there will be no routing directly to a Tier 2 or Tier 3 team.  The first responders for any given service team (i.e. DSET) will manage the ticket and determine if additional tiered escalation is needed.
- There will be a Data Element contained with in the code that will identify  the team where it needs to be directed
- Phase 1 will include a pull down list provided through the Abstraction Layer listing all cloud services which the third parties would select a subset depending on what is being reported.
- There is a requirement on the non-cloud Service team for the specific Cloud Service being reported to be passed along with all other data.
- Assumption is that non-cloud Service team will have the ability to view the services subscribed by their client and identify that in the original request.

### 2) Case Notifications
- the ability to receive updates on the support ticket within the policy defined in the service description via the non-cloud Service team portal

#### Scope Definition:
For MVP we will provide only cases that originate from ServiceNow today and in the future. So this means in the short term we will only be able to provide PaaS cases for any services that is managed in Servicenow today. In the future when the new function is available, we will also be able to provide IaaS cases for linked accounts only. We will not be able to provide cases for IaaS only customers. 

Current examples are:
  - PaaS and Watson services onboarded into ServiceNow. (This includes services who have not yet boarded into TIP but are managed in ServiceNow)
  - Linked accounts requesting support that originate in the IMS system and bridge cases from IMS to ServiceNow
  - net new OneCloud data center clients 

A more detailed explanation of why scope is defined that way can be found here: (https://github.ibm.com/cloud-sre/pnp-case/issues/5)

#### Sample Process Flow:
- Ticket handling, problem resolution, & communication will progress using the standard OneCloud support processes and procedures. 
- As updates are made and posted to the case by IBM OneCloud support teams, the data fields as defined in the integration design will be made available to the subscribed non-cloud service team through the API abstraction layer
- The requested data will be pulled by the non-cloud service team member and merged into their ticket system as they deem appropriate and per their own account/client specific processes.
- Communication back to the client and/or originator will remain the responsibility of the originating non-cloud service team
- The non-cloud service team will close the ticket through the Plug & Play API and within their own system upon completion
- If an unplanned event for outage qualifies for an RCA to be generated by the cloud support teams, the standard IBM Cloud RCA request process will be followed 
- Standard cloud escalation processes apply to API generated cases in the same way as tickets manually entered into IBM CLoud portal
- Expect some correlation of tickets across accounts when a CIE affects multiple clients

#### Assumptions and Dependencies
- The data made available through the API abstraction layer will not contain specific IBM cloud support names or other sensitive personal information 
- Existing support processes will be maintained or improved to ensure appropriate data is entered into data fields that will pass to the non-cloud service team via the API abstraction layer

### 3) Service Status
- The ability to receive service status information for IBM CLoud services via the Abstraction layer API, which can then be integrated with other non-cloud Service team portals (i.e. green, red, yellow, region impacted, description of the status change, etc. ) by the service owners. 

- Service Status is the status assigned to a service or component based on availability. Availability for a specific service or component may be measured in different ways, but generally the availability means the ability for a service or component to perform its advertised capabilities for clients at a given moment in time.

#### Scope Definition:
- The service status indicator of Red or Green will be set based on the implications of all planned and unplanned service events for IaaS, & PaaS services onboarded into OSS/TIP:
  - Unplanned disruptive events
  - Planned disruptive changes and maintenance events
  - Planned non-disruptive changes for PaaS
  - Planned non-disruptive changes for IaaS by end of 3Q18 (dependent on SN API development between IMS & SN)

- NOTE: IaaS planned non-disruptive events are not passed to ServiceNow or RTC today but as the SN-IMS API is developed they would also be included in the MVP (End of 3Q18 is estimated)

#### Sample Process Flow:
- Plug-n-Play participants would have the status indicator of either Red or Green available to them through the API abstraction player for all defined OneCloud services 
- The PnP participants would subsribe to any or all of the OneCloud services that are applicable to their job role, client contracts, etc.
- When an unplanned or planned event takes place which indicates a change in status to RED, a status update would be sent to all subscribers of that particular service
- The subscriber would then have the ability to use the Service Notification APIs to obtain more detailed infromation about the status change and to track the resolution progress.
- Based on the data provided and knowledge of the specific account and/or client the non-cloud service team will determine if and how it would affect them and communicate it apporpriately based on local process and procedures

#### Assumptions and Dependencies
- Service Status will be exposed as either up or down  (Red and Green only)
- Green status is a service that does not have any open Confirmed CIEs in ServiceNow
- Red status is a service that has at least one open Confirmed CIE in ServiceNow
- Yellow Status will not be provided but the data model will be designed to include degraded service at a later time

### 4) Service Notifications
- The ability to receive notifications of service status, via the Abstraction layer API and per the Cloud notification policy, on actions taken to restore a service following a CIE.
- Non-cloud Service teams can then use that data to integrate with other notifications and status updates within their own internal portals.

- Service Notifications are information that apply to services and components. Notifications can contain status as well as text providing further information about the status at a given moment in time. As an example, the information may, but are not required to, include the reason for the degredation or outage, current actions being taken, and expected resolution time. In this design, notifications can be received for status, incidents, and maintenance.

#### Scope Definition:
- Notifications will be made available through the API Abstraction layer for all planned and unplanned service outages for IaaS, & PaaS services onboarded into OSS/TIP:
  - Unplanned disruptive events
  - Planned disruptive changes and maintenance events
  - Planned non-disruptive changes for PaaS
  - Planned non-disruptive changes for IaaS by end of 3Q18 (dependent on SN API development between IMS & SN)

#### Sample Process Flow:
- The PnP participants would subsribe to any or all of the OneCLoud services that are applicable to their job role, client contracts, etc.
- When an unplanned event (incident) or planned event (disruptive change/maintenance) occurs a record is posted via the API abstraction layer
- All subscribers to the service(s) would be notified through the API abstraction layer
- non-cloud service teams would evaluate the service notification data and make a determination whether it affects their client and communicate outward as appropiate based on thier own specific processes and procedures
- As the outage progresses and services are restored, the subscribers will be notified through the API abstraction layer as updates are entered into the CIE ticket and continue through the life cycle of the event

#### Assumptions and Dependencies
- As Planned non-disruptive changes for IaaS are available ServiceNow via IMS-SN API being developed we will include them with all other changes being made available through the API Abstraction layer.

### 5) Reporting 
- ability for non-cloud service teams to extract data generated in Cases 1-4 and create reports which are client specific and/or service specific (i.e. support tickets opened this month, support tickets closed this month, # of outages this week / month, etc.)

#### Scope Definition:
- No specific API solution is being developed for reporting since most non-cloud service teams have client specific requirements which would require editing and merging of IBM OneCloud data with theor own systems of record.  

#### Assumptions and Dependencies
- The data provided from Use cases 1-4 will be use as input for developing reports by the non-cloud service teams within thier own account or client specific processes, procedures, and tools


### Open questions and Considerations:
1. Should we consider running the abstraction layer in the IBM cloud?   Yes but TBD.  The primary driver is availability and not location   
   - Willing to trade latency for availability 
2. Need to consider ITAR workload and be restrictive on who can see it.   May need ATAR audit for ITAR clients.
   - No mechanism today to lock down ServiceNow
   - ServiceNow cannot handle ITAR today so will need to evaluate
3. Same considerations are applicable for Fed Ramp accounts
4. 3rd bucket is EU accounts & need to be flagged for EU handling process for account
   - All personnel needs to be working rom EU and reporting to somewhere other than IBM USA
   - Cloud is okay from our side but we need to flag the accounts



## API Documents

[Swagger API](http://sretools.rtp.raleigh.ibm.com/pnp.html) documentation is available which defines each of the Plug-n-Play interfaces.

## Related Projects

The following are sub-projects which build the overall Plug-n-Play solution.

- [pnp-status](https://github.ibm.com/cloud-sre/pnp-status) : Provides Service Status APIs
- [pnp-case](https://github.ibm.com/cloud-sre/pnp-case) : Provides Case Management APIs
- [pnp-subscription](https://github.ibm.com/cloud-sre/pnp-subscription) : Provides the subscription interface which allows clients to receive notifications.

## Miscellaneous
- [Database Schema](https://github.ibm.com/cloud-sre/pnp-abstraction/blob/master/DatabaseSchema.png) : 
This is an [image created from pnp-status](https://github.ibm.com/cloud-sre/pnp-status/blob/master/images/DatabaseSchema.xml)
- [Update Database Schema](https://github.ibm.com/cloud-sre/pnp-abstraction/blob/master/datastore/README_UpdateDB.md)
- [PnP Overview](https://ibm.box.com/s/ndq6gbznwd8q6rwy6l6tw91mg0wi1mis) : General overview slides for PnP
- [PnP POC - contains the RabbitMQ mappings](https://ibm.ent.box.com/file/308539780761)
