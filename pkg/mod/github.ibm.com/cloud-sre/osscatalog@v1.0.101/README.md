# osscatalog. 

## Overview

*osscatalog* is a common library and a set of tools for manipulating OSS records in the IBM Cloud Global Catalog.

The Global Catalog (https://globalcatalog.cloud.ibm.com) is the central repository for information about services and other entities that are part of IBM Cloud.

Every service, runtime, internal component, subcomponent, etc. in IBM Cloud is associated with a **OSS record**, that captures all the OSS-related information for that item (ownership, compliance, operational and support parameters, etc.). It is uniquely identified by its **CRN service-name** (even for things that are not, strictly-speaking, services).

In most cases, each item might also have a **Main record** in Global Catalog, which controls how the item (service) is shown to Clients on the *Client-facing Catalog* (https://cloud.ibm.com/catalog) and how it interacts with platform systems like IAM (for access management), Resource Controller (for provisioning), and billing.

The OSS record and the Main record are both stored in the Global Catalog, but they are distinct records. This allows them to be owned and managed by different owners and to have different visibility.
- Each Main record (if it exists) has the CRN service-name as its **name** attribute (per the [CRN spec](https://github.ibm.com/ibmcloud/builders-guide/blob/master/specifications/crn/CRN.md)) and is of a particular **kind** such as `service`, `runtime`, `template`, `platform_service`, `iaas`, etc.
- Each OSS record has a **name** of the form `oss.<CRN-service-name>` and **kind** `oss`
- There are also OSS records of **kind** `oss_segment` and `oss_tribe` for capturing information about Segments and Tribes

The following tools are provided (see sections below for more details on using each tool):
- **osscatviewer**: a simple viewer for OSS records in the Global Catalog. You could in principle use the main UI of the Global Catalog to view OSS records, but you won't see any of the custom attributes specific to these OSS records, so that's not very useful. **oscatviewer** displays a custom view specifically designed to show all the pertinent OSS information. 
  - You can access **osscatviewer** through a browser, at https://osscatviewer.stage1.mybluemix.net/
  - **osscatviewer** also has a special "edit mode" that allows some privileged users to edit some aspects of the OSS record directly  to control the merge process (see below)
- **osscatimporter**: a command line tool that creates or updates all OSS records. It does this by reading information from multiple registries (ServiceNow, Scorecard, the Main record in Global Catalog itself, ClearingHouse, etc.) and combining/merging the information from each registry to form a single OSS record for each item. It automatically resolves any conflicts or discrepancies encountered during the merge, but records a *validation issue* for each.
- **osscatpublisher**: a command line tool that copies OSS records from the Staging Global Catalog (https://globalcatalog.test.cloud.ibm.com) to Production (https://globalcatalog.cloud.ibm.com). For safety, and because the merge operation can sometimes be problematic (and is still undergoing constant enhancements), **osscatimporter** always writes or updates the OSS records only in the Staging Catalog. Once the results of a merge have been verified, you can use **osscatpublisher** to copy these records to the Production Catalog.

The intention is that **osscatimporter** and **osscatpublisher** should only be temporary. At some point in the future, onboarding of new services will be performed through a single centralized portal (e.g. RMC), which will write the OSS records directly to Global Catalog. Most of the current registries (e.g. ServiceNow, Scorecard) will become *consumers* of the OSS information centralized in Global Catalog instead of being independent onboarding portals and sources of various fragments of OSS information, as they are today.


## Structure of a OSS record

Each OSS record associated with a service, runtime, component, etc. contains 4 notable sections / elements:
- The **main body** (type [`OSSMetaData`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossrecord/ossrecord.go)): contains all the OSS information associated with this item, for use by various consumers that require this information (e.g. support flows, change management, incident management, monitoring flows, etc.)
- The **OSS merge control information** (type [`OSSMergeControl`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossmergecontrol/ossmergecontrol.go)) contains a handful of attributes that are used to control how this OSS record is being generated, by merging information extracted from multiple sources (ServiceNow, Scorecard, the Main record in Global Catalog itself, ClearingHouse, etc.). These attributes specify for example how multiple entries with similar names should be merged together or kept separate; the ability to override some values obtained from the original entries during the merge; and the *OSS Merge Control Tags* (see below)
- The **OSS validation information** (type [`OSSValidation`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossvalidation/validations.go)) contains a log of the most recent merge operation that created or updated this OSS record: a trace-back to the source entries in other registries from which this record was merged, and a list of *Validations Issues* that reflect every anomaly or discrepancy that was detected during the merge
- The **OSS Tags** (type [`osstags.TagSet`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/osstags/osstags.go)) are a set of pre-defined but extensible text tags that augment the information associated with a OSS record without having to define a new attribute in the main body for each tag. These tags fall in two categories:
   - **Merge Control Tags** modify how a OSS record is represented or merged. For example, the `type_subcomponent` tag specifies that this item should be considered a subcomponent, even though it may be represented as a service in some of the source registries for the merge (which may not natively support a "subcomponent" type); the `notready` tag specifies that this item has not yet completed onboarding in all the necessary registries, and thus various merge validation errors should be ignored. These Merge Control Tags can be specified in the `OSSTags` attribute of the [`OSSMergeControl`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossmergecontrol/ossmergecontrol.go) section, and they are automatically copied to the main body after a merge, and carried over from one merge of the same record to the next.
   - **Merge Status Tags** reflect some status of the OSS record as a result of a merge (for example `oss_status_green` if there are no serious validation issues, or `pnp_enabled` if this item is suitable for access by the PnP system). These Merge Status Tags exist only in the `GeneralInfo.OSSTags` attribute of the main body ([`OSSMetaData`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossrecord/ossrecord.go_) and they are regenerated automatically during each merge. It is illegal to specify any Merge Status Tags in the [`OSSMergeControl`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossmergecontrol/ossmergecontrol.go) section of the OSS record
  
There are also OSS records that represent Segments and Tribes (linked back to the OSS records for the service/runtimes/components/etc. that belong to each Segment/Tribe). At this time, these OSS records only contain a main body section (type [`OSSSegment`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossrecord/osssegment.go) and [`OSSTribe`](https://github.ibm.com/cloud-sre/osscatalog/blob/master/ossrecord/osstribe.go)). They have no merge control and no validation information. These OSS records are also created/updated by the **osscatimporter** tool, but there is no real merge involved. They are basically copied verbatim from Scorecard, which is the only external registry that contains pertinent information for Segments and Tribes at this time.

Finally, note that the OSS records stored in the Staging Catalog contain all the sections above. The OSS records copied into the Production Catalog by the **osscatpublisher** tool only contain the main body section.

### Description of key attributes in the main body of a OSS record

  - `ReferenceResourceName`: the CRN service-name that uniquely identifies this entry
  - `ReferenceDisplayName`: a user-friendly name for this entry (may not be unique, and may change over time)
  - `ReferenceCatalogID`: the Global Catalog ID of the *Main record* in Global Catalog that corresponds to this OSS record -- if there is one
  - `GeneralInfo.EntryType`: the type of this entry (`SERVICE`, `RUNTIME`, `IAAS`, `PLATFORM_COMPONENT`, `SUBCOMPONENT`, etc.)
  - `GeneralInfo.OperationalStatus`: the status of this entry (`GA`, `BETA`, `EXPERIMENTAL` `THIRDPARTY`, `NOTREADY`, etc.)
  - `GeneralInfo.ParentResourceName`: the CRN service-name of some other OSS record that can be considered a *parent* of this entry, for example between a service and multiple subcomponents of that service (most often it is empty)
  - `GeneralInfo.OSSTags`: see above
  
There are many other attributes in various subsections (Ownership, Support, Operations, Compliance, StatusPage, etc.) -- see the source code and source documentation for details.

## Viewing OSS records with osscatviewer

## Creating/updating OSS records with osscatimporter

## Publishing OSS records to Production with osscatpublisher

## Managing the PnPEnabled flag and the status page meta-data

