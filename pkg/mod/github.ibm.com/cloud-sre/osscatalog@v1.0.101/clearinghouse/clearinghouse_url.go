package clearinghouse

const (
	// chDeliverablePageURL is the URL pattern to link to the page for one Deliverable record in ClearingHouse
	chDeliverablePageURL = "https://clearinghousev2.raleigh.ibm.com/CHNewCHRDM/CCHMServlet#&nature=wlhNDE&deliverableId=%s"

	chSearchURL = "https://w3.api.ibm.com/portfoliomgmt/run/ch-product-release/product_releases?name=%s&exact-match=false&contains=true&starts-with=false"

	chByIDURL = "https://w3.api.ibm.com/portfoliomgmt/run/ch-product-release/product_releases/%s?expand=dependency_providers,dependency_originators"

	chDependencyPageURL = "https://clearinghousev2.raleigh.ibm.com/CHNewCHRDM/CCHMServlet#&nature=wlhDepForm&dependencyId=%s&dependencyType=Usage"

	chTokenName = "clearinghouse"
)

/*
To renew the ClearingHouse token (every 180 days):
- Go in a browser to https://w3.api.ibm.com/portfoliomgmt/run/ch/oauth-end/oauth2/authorize?response_type=token&client_id=<YOUR-CLIENT-ID>&redirect_uri=https://localhost&scope=view_chdata
  Where <YOUR-CLIENT-ID> is the ClearingHouse client ID from the keyfile osscat.key (not the token)
- This will redirect to a dead page in the browser at http://localhost. But in the redirect URL, you will find the new token.
- Copy the new ClearingHouse token into the keyfile osscat.key

See original instructions from ClearingHouse team at https://w3-connections.ibm.com/wikis/home?lang=en#!/wiki/W7ee29825e2f5_4679_af9d_3f1e757f58d0/page/CH%20REST%20APIs
*/
