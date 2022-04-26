package main

import (
	"log"
	"time"

	"github.ibm.com/cloud-sre/pnp-notifications-adapter/initadapter"
)

func main() {
	restarter(1, "main", startup)
}

func restarter(panics int, id string, f func()) {
	FCT := "Restarter: "
	defer func() {
		if err := recover(); err != nil {
			log.Println(FCT+"PANIC: Not able to recover :", id)
			log.Println(FCT, err)
			if panics > 10 {
				panic("ERROR: Too many panics.  Do something.")
			} else {
				time.Sleep(time.Duration(panics) * time.Second)
				go restarter(panics+1, id, f)
			}
		}
	}()
	f()
}

func startup() {
	_, err := initadapter.Initialize()
	if err != nil {
		log.Println(err)
		return
	}

	// gavila remove dead code
	//err = runMain(creds)
	//if err != nil {
	//	log.Println(err)
	//	return
	//}
}

var doneChan = make(chan bool)

// gavila remove dead code
//func runMain(creds *api.SourceConfig) error {
//	counter := 0
//	if creds == nil {
//		return errors.New("credentials are required")
//	}
//
//	nrMon, err := exmon.CreateMonitor() // Create new relic monitor
//	if err != nil {
//		return err
//	}
//
//	// Initial load so we don't wait for the first tick
//	err = loadNotifications(ctxt.Context{LogID: fmt.Sprintf("initial%d", counter), NRMon: nrMon}, creds)
//	if err != nil {
//		return err
//	}
//
//	loopFrequency := utils.GetEnvMinutes(LoopFrequencyEnv, 60)
//
//	log.Println("notifications.runMain(): Starting ticker frequency=" + loopFrequency.String())
//	ticker := time.NewTicker(loopFrequency)
//	loadChan := ticker.C
//
//	// System termination condition
//	sysChan := make(chan os.Signal, 1)
//	signal.Notify(sysChan, os.Interrupt, os.Kill)
//
//	// Custom Termination condition
//	//doneChan := make(chan bool)
//	go closeCondition(doneChan)
//
//	for {
//		select {
//		case <-loadChan:
//			log.Println("notifications.runMain(): Start load of notifications. Frequency=" + loopFrequency.String())
//			counter++
//			loadNotifications(ctxt.Context{LogID: fmt.Sprintf("ticker%d", counter), NRMon: nrMon}, creds)
//		case <-sysChan:
//			ticker.Stop()
//			cleanup()
//			return nil
//		case <-doneChan:
//			ticker.Stop()
//			cleanup()
//			return nil
//		}
//	}
//}

func terminate() {
	doneChan <- true
}

// Used to provide a termination condition if needed. Set doneChan to true
func closeCondition(doneChan chan bool) {
}

// Used to provide any orderly cleanup
func cleanup() {
	log.Println("Orderly exit.")
}

// loadNotifications is the launch point to load up the notifications
//func loadNotifications(ctx ctxt.Context, creds *api.SourceConfig) error {
//	METHOD := "loadNotifications"
//
//	cNotifications, err := api.GetNotifications(ctx, creds)
//
//	if err != nil {
//		return fmt.Errorf("ERROR (%s): Could not load notifications from cloudant. [%s]", METHOD, err.Error())
//	}
//
//	if cNotifications == nil || len(cNotifications.Items) == 0 {
//		return fmt.Errorf("no valid notifications found in cloudant")
//	}
//
//	log.Println("Got a list of notifications of length", len(cNotifications.Items))
//
//	pgDB, err := postgres.Open(os.Getenv(postgres.PostgresHost), os.Getenv(postgres.PostgresPort), os.Getenv(postgres.PostgresDB), os.Getenv(postgres.PostgresUser), os.Getenv(postgres.PostgresPass), os.Getenv(postgres.PostgresSSLMode))
//	if err != nil {
//		return fmt.Errorf("ERROR (%s): Could not open the postgres database connection. [%s]", METHOD, err.Error())
//	}
//	defer pgDB.Close()
//
//	pgNotifications, err := pgDB.GetAllNotifications(ctx, "", true)
//	if err != nil {
//		return fmt.Errorf("ERROR (%s): Could not load notifications from postgres. [%s]", METHOD, err.Error())
//	}
//
//	conn, err := mq.OpenConnection([]string{os.Getenv(mq.MQURL), os.Getenv(mq.MQURL2)}, mq.RoutingKey, mq.ExchangeName, mq.ExchangeType)
//	if err != nil {
//		return fmt.Errorf("ERROR (%s): Could not open message queue connection. [%s]", METHOD, err.Error())
//	}
//	// defer conn.Close()
//	if pgNotifications == nil || len(pgNotifications.Items) == 0 {
//		err = conn.SendNotifications(ctx, convert.CnToPGniList(cNotifications), datastore.BulkLoad)
//		if err != nil {
//			return fmt.Errorf("ERROR (%s): Could not bulk load notifications. [%s]", METHOD, err.Error())
//		}
//	} else {
//		log.Printf("INFO (%s): Performing comparison of %d existing items to %d new items", METHOD, len(pgNotifications.Items), len(cNotifications.Items))
//		err = conn.SendCompareAndUpdateNotifications(ctx, convert.CnToPGniList(pgNotifications), convert.CnToPGniList(cNotifications))
//		log.Printf("INFO (%s): Compare complete", METHOD)
//		if err != nil {
//			return fmt.Errorf("ERROR (%s): Could not compare load notifications. [%s]", METHOD, err.Error())
//		}
//	}
//	return nil
//}

/*
// Alternate main routine used for testing. This main deletes all the notifications. Uncomment this main and comment out the other one.
func main() {
	METHOD := "main"
	//postgres.SetupDBFunctions()

	_, err := initadapter.Initialize()
	if err != nil {
		log.Println(err)
		return
	}
	ctx := ctxt.Context{LogID: "SpecialMainForDelete"}

	for i := 0; i < 10; i++ {
		pgDB, err := postgres.Open(os.Getenv(postgres.PostgresHost), os.Getenv(postgres.PostgresPort), os.Getenv(postgres.PostgresDB), os.Getenv(postgres.PostgresUser), os.Getenv(postgres.PostgresPass), os.Getenv(postgres.PostgresSSLMode))
		if err != nil {
			fmt.Printf("ERROR (%s): Could not open the postgres database connection. [%s]", METHOD, err.Error())
			return
		}

		err = pgDB.DeleteAllNotifications(ctx)
		if err != nil {
			log.Println(err)
		} else {
			log.Println("Delete complete successfully")
		}
		pgDB.Close()

		time.Sleep(time.Second * 5)
	}

	ticker := time.NewTicker(time.Minute * 1)
	// System termination condition
	sysChan := make(chan os.Signal, 1)
	signal.Notify(sysChan, os.Interrupt, os.Kill)

	for {
		select {
		case <-sysChan:
			ticker.Stop()
			return
		}
	}
}
*/
