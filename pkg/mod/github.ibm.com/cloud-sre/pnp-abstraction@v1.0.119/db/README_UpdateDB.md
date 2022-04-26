# Plug-n-Play Datastore

Plug-n-Play uses Postgres Database for data storage. Please see /pnp-abstraction/images/DatabaseSchemaForPostgres.png for database schema.

## What to do if database table needs modification?
  1. See https://ibm.ent.box.com/notes/320112657054 on how to access PnP Postgres database.

  2. If you need to add a new column, run ALTER command in development manually, this will retain the data in the table. For example,
  `ALTER TABLE resource_table ADD COLUMN IF NOT EXISTS record_hash varchar(64);`
  If you want to remove ` NOT NULL` constraint of a column, run ALTER TABLE, for example,
  `ALTER TABLE watch_table ALTER COLUMN crn_full DROP NOT NULL;`

  3. If the newly added column needs an index for fast searching, then run CREATE INDEX command; otherwise go to next step. For example,
  `CREATE INDEX IF NOT EXISTS resource_catalog_parent_id_index on resource_table(catalog_parent_id);`

  4. Add the new column name in pnp-abstraction/db/constants.go 

  5. Update the structures in pnp-abstraction/datastore, modify insert.go, patch.go and read.go if needed. If the newly added column will have impact to record hash of a  maintenance or resource, then you might need to update computeMaintenanceRecordHash or computeResourceRecordHash.

  6. Update /pnp-abstraction/images/DatabaseSchemaForPostgres.xml using https://www.draw.io/
  Make a screenshot of the updated diagram and replace DatabaseSchemaForPostgres.png

  7. Add the new column to the GetCreatexxxTableStatment of pnp-db/dbcreate/create.go. We need to update pnp-db, because the test cases run in Jenkins will create Postgres ddatabase in Docker based on the functions in pnp-db.

  8. If you have created a new index in step 3, then add the new index in pnp-db/dbcreate/constants.go, and update CreatexxxTable and DropxxxTable in pnp-db/dbcreate/create.go.

  9. Test your changes in development environment to ensure all your changes are working fine.

  10. Add new test cases in pnp-db, **please READ** pnp-db/README.md before you test run the test cases. Please note, the test cases in pnp-db are not running mocked data, they are actually inserting/updating/deleting to the real Postgres database, and exercising the db functions in pnp-abstraction. If you deploy pnp-db to development environment, then it is test run in development environment. So, it is recommended to create a new set of tables for testing. **Never deploy pnp-db to staging and production.**

  10. Repeat step 1-3 for staging environment.

  11. If all your unit tests passed, check in and merge your changes in pnp-abstraction.

  12. Check in and merge your changes in pnp-db, make sure `EnsureTablesExist(database)` and  `RunTests(database)` are commented out in main.go

  13. Once you merge your pull request for pnp-abstraction, it is deployed to development and staging environment. If they both works fine, repeat step 1-3 in production environment. **Make sure ALTER command is done before the pull requests are deployed to production.**
