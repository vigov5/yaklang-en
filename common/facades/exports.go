package facades

var FacadesExports = map[string]interface{}{
	"NewFacadeServer": NewFacadeServer,
	"Serve":           Serve,

	// Use parameters
	"javaClassName":     SetJavaClassName,
	"javaCodeBase":      SetJavaCodeBase,
	"objectClass":       SetObjectClass,
	"javaFactory":       SetjavaFactory,
	"httpResource":      SetHttpResource,
	"ldapResourceAddr":  SetLdapResourceAddr,
	"rmiResourceAddr":   SetRmiResourceAddr,
	"evilClassResource": SetRmiResourceAddr,
}
