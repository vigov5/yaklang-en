package vulinbox

import (
	"net/http"
)

func (s *dbm) Authenticate(writer http.ResponseWriter, request *http.Request) (*VulinUser, error) {
	session, err := request.Cookie("_cookie")
	if err != nil {
		//writer.Header().Set("Location", "/logic/user/login")
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`
<script>
alert("Please log in first");
window.location.href = '/logic/user/login?from=` + request.URL.Path + `';
</script>
`))
		return nil, err
	}

	auth := session.Value
	se, err := s.GetUserBySession(auth)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal error session " + err.Error()))
		return nil, err
	}

	// Perform the logic of obtaining user details here
	// Assume querying user information based on user name
	users, err := s.GetUserByUsernameUnsafe(se.Username)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal error, cannot retrieve user information"))
		return nil, err
	}
	user := users[0]

	return user, nil
}

// Authentication of shopping mall
func (s *dbm) mallAuthenticate(writer http.ResponseWriter, request *http.Request) (*VulinUser, error) {
	session, err := request.Cookie("sessionID")
	if err != nil {
		writer.WriteHeader(http.StatusOK)
		writer.Write([]byte(`
<script>
alert("Please log in first");
window.location.href = '/mall/user/login';
</script>
`))
		return nil, err
	}

	auth := session.Value
	se, err := s.GetUserBySession(auth)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal error session " + err.Error()))
		return nil, err
	}

	// Perform the logic of obtaining user details here
	// Assume querying user information based on user name
	users, err := s.GetUserByUsernameUnsafe(se.Username)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal error, cannot retrieve user information"))
		return nil, err
	}
	user := users[0]

	return user, nil
}

//func (s *dbm) IsAdmin(username string) bool {
//	var user VulinUser
//	if err := um.db.Where("username = ? AND role = ?", username, "admin").First(&user).Error; err != nil {
//		return false
//	}
//	return true
//}
