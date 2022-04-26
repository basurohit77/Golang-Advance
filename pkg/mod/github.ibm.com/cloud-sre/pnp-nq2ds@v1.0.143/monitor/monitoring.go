package monitor

import (
    "os"
)


const (
       SrvPrfx                      =  "pnp-nq2ds-"
       TxnNRdbConnection            =  "nq2ds-db-reconnection"
       TxnNRprocIncident            =  SrvPrfx+"process-incident"
       TxnNRprocIncPublic           =  TxnNRprocIncident+"-public"
       TxnNRprocMantSN              =  SrvPrfx+"process-maintenance"
       TxnNRdecodeAndMap            =  SrvPrfx+"decodeAndMap-maintenance"
       TxnNRprocNotification        =  SrvPrfx+"process-notification"
       TxnNRsendNoteToDB            =  SrvPrfx+"sendNoteToDB"
       TxnNRprocResource            =  SrvPrfx+"process-resource"
       TxnNRinsertUpdateResource    =  SrvPrfx+"insertUpdateResource"
       TxnNRprocStatus              =  SrvPrfx+"process-status"
       TxnNRpostIncident            =  SrvPrfx+"post-incident"
       TxnNRpostMaintenance         =  SrvPrfx+"post-maintenance"
       TxnNRpostNotification        =  SrvPrfx+"post-notification"
       TxnNRpostNotifSubConsumer    =  SrvPrfx+"post-notification-sub"
       TxnNRpostResource            =  SrvPrfx+"post-resource"
)

var (
 NRenvironment = os.Getenv("KUBE_APP_DEPLOYED_ENV")
 NRregion = os.Getenv("KUBE_CLUSTER_REGION")
 ENVIRONMENT = os.Getenv("K8S_ENV")
 REGION = os.Getenv("K8S_REGION")
 ZONE = os.Getenv("K8S_ZONE")
)
