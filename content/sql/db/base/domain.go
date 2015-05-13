package base

func init() {
	sql["get_domain_https_support"] = getDomainHTTPSSupport
	sql["create_domain_https_support"] = createDomainHTTPSupport
	sql["update_domain_https_support"] = updateDomainHTTPSupport
}

const (
	getDomainHTTPSSupport   = `SELECT https FROM domain_https_support WHERE domain = $1`
	createDomainHTTPSupport = `
INSERT INTO domain_https_support(domain, https)
	SELECT $1, $2 EXCEPT
	SELECT domain, https FROM domain_https_support WHERE domain = $1`
	updateDomainHTTPSupport = `UPDATE domain_https_support SET https = $1 WHERE domain = $2`
	deleteDomainHTTPSupport = `DELETE FROM domain_https_support WHERE domain = $1`
)
