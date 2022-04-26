[![Build Status](https://wcp-cto-sre-jenkins.swg-devops.com/buildStatus/icon?job=Pipeline/api-pnp-data-encryption/master)](https://wcp-cto-sre-jenkins.swg-devops.com/job/Pipeline/job/api-pnp-data-encryption/job/master/)

# pnp-data-encryption

Package used to encrypt and decrypt text.  
The encryption used is AES_256_GCM, with a nonce and a derived key per message.

- Download

`go get github.ibm.com/cloud-sre/pnp-data-encryption`


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


## How to use

- Set MASTER_KEY environment variable

`export MASTER_KEY=123456`

- Encrypt a string

```go
encryptedMsg, err = encryption.Encrypt("this is my string msg to be encrypted")
	if err != nil {
		log.Println(err)
    }
log.Println("Encrypted msg:", encryptedMsg)
```

- Decrypt a string

```go
decryptedMsg, err := encryption.Decrypt(encryptedMsg)
	if err != nil {
		log.Println(err)
	}

	log.Printf("Original msg: %s", decryptedMsg)
```
