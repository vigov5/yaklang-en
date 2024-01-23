package vulinbox

import (
	_ "embed"
	"encoding/json"
	"github.com/yaklang/yaklang/common/log"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"
)

//go:embed html/mall/vul_mall_register.html
var mallRegisterPage []byte

//go:embed html/mall/vul_mall_login.html
var mallLoginPage []byte

//go:embed html/mall/vul_mall_userProfile.html
var mallUserProfilePage []byte

func (s *VulinServer) mallUserRoute() {
	// var router = s.router
	// malloginGroup := router.PathPrefix("/mall").Name("shopping mall").Subrouter()
	mallloginRoutes := []*VulInfo{
		//Login function
		{
			DefaultQuery: "",
			Path:         "/user/login",
			// Title:        "Mall login",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					// Return to the login page
					writer.Header().Set("Content-Type", "text/html")
					writer.Write(mallLoginPage)
					return
				}

				// Parse the JSON data in the request body
				var loginRequest struct {
					Username string `json:"username"`
					Password string `json:"password"`
				}

				err := json.NewDecoder(request.Body).Decode(&loginRequest)
				if err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte("Invalid request"))
					return
				}

				username := loginRequest.Username
				password := loginRequest.Password

				// Execute user login logic here to verify whether the user name and password are correct
				// Check whether there is a match in the database User information
				if username == "" || password == "" {
					writer.WriteHeader(400)
					writer.Write([]byte("username or password cannot be empty"))
					return
				}
				// sql injection, universal password
				users, err := s.database.GetUserByUnsafe(username, password)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte("internal error: " + err.Error()))
					return
				}
				user := users[0]

				// Assuming the verification is passed, return the login success message
				response := struct {
					Id      uint   `json:"id"`
					Success bool   `json:"success"`
					Message string `json:"message"`
				}{
					Id:      user.ID,
					Success: true,
					Message: "Login successful",
				}
				var sessionID = "fixedSessionID"
				// session, err := user.CreateSession(s.database)
				if err != nil {
					return
				}
				http.SetCookie(writer, &http.Cookie{
					Name:    "sessionID",
					Value:   sessionID,
					Path:    "/",
					Expires: time.Now().Add(15 * time.Minute),

					//HttpOnly: true,
				})
				writer.Header().Set("Content-Type", "application/json")
				err = json.NewEncoder(writer).Encode(response)
				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				writer.WriteHeader(http.StatusOK)
				return
			},
			RiskDetected: true,
		},
		//Registration function
		{
			DefaultQuery: "",
			Path:         "/user/register",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				if request.Method == http.MethodGet {
					// Return to the login page
					writer.Header().Set("Content-Type", "text/html")
					writer.Write(mallRegisterPage)
					return
				} else if request.Method == http.MethodPost {
					// Parse the JSON data in the request body
					user := &VulinUser{
						Role: "user",
					}
					err := json.NewDecoder(request.Body).Decode(user)
					if err != nil {
						writer.Write([]byte(err.Error()))
						writer.WriteHeader(http.StatusBadRequest)
						return
					}

					remake := strings.ToLower(user.Remake)
					filterRemake := strings.ReplaceAll(remake, "<", "")
					filterRemake = strings.ReplaceAll(filterRemake, ">", "")
					filterRemake = strings.ReplaceAll(filterRemake, "script", "")
					user.Remake = filterRemake

					// Execute user registration logic here and store user information into the database
					err = s.database.CreateUser(user)
					if err != nil {
						writer.Write([]byte(err.Error()))
						writer.WriteHeader(http.StatusInternalServerError)
						return
					}

					// Assuming the verification is passed, return the login success message
					responseData, err := json.Marshal(user)
					if err != nil {
						writer.Write([]byte(err.Error()))
						writer.WriteHeader(http.StatusInternalServerError)
						return
					}
					response := struct {
						Id      uint   `json:"id"`
						Success bool   `json:"success"`
						Message string `json:"message"`
						Data    string `json:"data"`
					}{
						Id:      user.ID,
						Success: true,
						Message: "Register successful",
						Data:    string(responseData),
					}
					writer.Header().Set("Content-Type", "application/json")
					err = json.NewEncoder(writer).Encode(response)
					if err != nil {
						writer.Write([]byte(err.Error()))
						writer.WriteHeader(http.StatusInternalServerError)
						return
					}
					writer.WriteHeader(http.StatusOK)
					return
				}
			},
		},
		//User information
		{
			DefaultQuery: "",
			Path:         "/user/profile",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				realUser, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}

				// Get the user by id Information
				var a = request.URL.Query().Get("id")
				i, err := strconv.ParseInt(a, 10, 64)
				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(500)
					return
				}
				userInfo, err := s.database.GetUserById(int(i))
				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(500)
					return
				}

				// Horizontal override
				if realUser.Role != "admin" && realUser.Role != userInfo.Role {
					writer.Write([]byte("Not Enough Permissions"))
					writer.WriteHeader(http.StatusBadRequest)
					return
				}

				// Return to user personal page
				writer.Header().Set("Content-Type", "text/html")
				tmpl, err := template.New("userProfile").Parse(string(mallUserProfilePage))

				//Get the number of items in the shopping cart
				Cartsum, err := s.database.GetUserCartCount(int(userInfo.ID))
				//The personal information page displays the number of items in the shopping cart
				type UserProfile struct {
					ID       int
					Username string
					Cartsum  int
				}
				profile := UserProfile{
					// ID:       int(userInfo.ID),
					Username: userInfo.Username,
					Cartsum:  int(Cartsum),
				}

				err = tmpl.Execute(writer, profile)

				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					log.Println("Execution template fails:", err)
					writer.Write([]byte("Internal error, cannot render user profile"))
					return
				}
			},
		},
		//Log out
		{
			DefaultQuery: "",
			Path:         "/user/logout",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				cookie, err := request.Cookie("_cookie")
				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(http.StatusBadRequest)
					return
				}
				uuid := cookie.Value
				err = s.database.DeleteSession(uuid)
				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(http.StatusInternalServerError)
					return
				}
				writer.WriteHeader(http.StatusOK)
				return
			},
		},
	}

	for _, v := range mallloginRoutes {
		addRouteWithVulInfo(MallGroup, v)
	}

}
