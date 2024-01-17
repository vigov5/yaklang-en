package vulinbox

import (
	_ "embed"
	"encoding/json"
	"fmt"

	// "fmt"
	"net/http"
	"strconv"
	"text/template"
)

//go:embed html/mall/vul_mall_cart.html
var mallCartPage []byte

func (s *VulinServer) mallCartRoute() {
	// var router = s.router
	// mallcartGroup := router.PathPrefix("/mall").Name("Mall").Subrouter()
	mallcartRoutes := []*VulInfo{

		//Shopping cart information
		{
			DefaultQuery: "",
			Path:         "/user/cart",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				//Verify whether you are logged in
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}
				// var Username = realuser.Role
				type TemplateData struct {
					CartInfo []UserCart
					Username string
				}
				// Get user shopping cart information by id
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
				cartInfo, err := s.database.GetCart(int(i))

				for i := range cartInfo {
					cartInfo[i].TotalPrice = cartInfo[i].ProductPrice * float64(cartInfo[i].ProductQuantity)
					// fmt.Println(userCart.ProductPrice)
				}
				data := TemplateData{
					Username: userInfo.Username,
					CartInfo: cartInfo,
				}

				if err != nil {
					writer.Write([]byte(err.Error()))
					writer.WriteHeader(500)
					return
				}
				// Print return value
				fmt.Printf("cartInfo: %+v\n", cartInfo)

				//If the shopping cart quantity is not 0, return the shopping cart page information
				writer.Header().Set("Content-Type", "text/html")
				tmpl, err := template.New("cart").Parse(string(mallCartPage))

				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					writer.Write([]byte("Internal error, cannot render user cartInfo1"))
					return
				}
				err = tmpl.Execute(writer, data)
				if err != nil {
					writer.WriteHeader(http.StatusInternalServerError)
					writer.Write([]byte("Internal error, cannot render user cartInfo2"))
					return
				}

			},
		},

		// Add to shopping cart
		{
			DefaultQuery: "",
			Path:         "/cart/add",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}

				// Parse the JSON data passed by the front end
				var cart UserCart
				err = json.NewDecoder(request.Body).Decode(&cart)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusBadRequest)
					return
				}

				// Call a function to add the product to the shopping cart
				err = s.database.AddCart(cart.UserID, cart)
				if err != nil {
					http.Error(writer, err.Error(), http.StatusInternalServerError)
					return
				}

				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("Add to shopping cart successfully"))

			},
		},
		//Get the number of items in the shopping cart
		{
			DefaultQuery: "",
			Path:         "/cart/count",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}
				//Parse the JSON data passed by the front end
				var RequestBody struct {
					UserID int `json:"userID"`
				}
				// var body RequestBody
				var ID UserCart

				//Get the userID passed by the front end
				err = json.NewDecoder(request.Body).Decode(&RequestBody)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte(err.Error()))
					return
				}
				ID.UserID = RequestBody.UserID
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte("UserID must be a number"))
					return
				}

				//Call the function to get the number of shopping cart items
				cartSum, err := s.database.GetUserCartCount(ID.UserID)
				if err != nil {
					return
				}
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(strconv.Itoa(int(cartSum))))

			},
		},

		// RiskDetected: true,
		//Check whether the shopping cart exists
		{
			DefaultQuery: "",
			Path:         "/cart/check",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}
				//Parse the JSON data passed by the front end
				var RequestBody struct {
					UserID      int    `json:"userID"`
					ProductName string `json:"productName"`
				}
				var ID UserCart

				//Get the userID and productName passed by the front end
				err = json.NewDecoder(request.Body).Decode(&RequestBody)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte(err.Error()))
					return
				}
				ID.UserID = RequestBody.UserID
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte("UserID must be a number"))
					return
				}
				ID.ProductName = RequestBody.ProductName

				//Call the function to check whether the product exists in the shopping cart
				cartExists, err := s.database.CheckCart(ID.UserID, ID.ProductName)
				if err != nil {
					return
				}
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte(strconv.FormatBool(cartExists)))

			},
		},

		//Shopping cart product plus one
		{
			DefaultQuery: "",
			Path:         "/cart/addOne",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}
				//Parse the JSON data passed by the front end
				var RequestBody struct {
					UserID      string `json:"userID"`
					ProductName string `json:"productName"`
				}
				var ID UserCart

				//Get the userID and productName passed by the front end
				err = json.NewDecoder(request.Body).Decode(&RequestBody)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte(err.Error()))
					return
				}
				ID.UserID, err = strconv.Atoi(RequestBody.UserID)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte("UserID must be a number"))
					return
				}
				ID.ProductName = RequestBody.ProductName

				//Call the function to add one shopping cart item
				err = s.database.AddCartQuantity(ID.UserID, ID.ProductName)
				if err != nil {
					return
				}
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("Shopping cart product plus one success"))

			},
		},
		//Shopping cart product minus one
		{
			DefaultQuery: "",
			Path:         "/cart/subOne",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}
				//Parse the JSON data passed by the front end
				var RequestBody struct {
					UserID      string `json:"userID"`
					ProductName string `json:"productName"`
				}
				var ID UserCart

				//Get the userID and productName passed by the front end
				err = json.NewDecoder(request.Body).Decode(&RequestBody)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte(err.Error()))
					return
				}
				ID.UserID, err = strconv.Atoi(RequestBody.UserID)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte("UserID must be a number"))
					return
				}
				ID.ProductName = RequestBody.ProductName

				//Call the function to subtract one shopping cart item
				err = s.database.SubCartQuantity(ID.UserID, ID.ProductName)
				if err != nil {
					return
				}
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("Decrease the shopping cart product by one Success"))

			},
		},
		//Delete the shopping cart item
		{

			DefaultQuery: "",
			Path:         "/cart/delete",
			Handler: func(writer http.ResponseWriter, request *http.Request) {
				_, err := s.database.mallAuthenticate(writer, request)
				if err != nil {
					return
				}
				//Parse the JSON data passed by the front end
				var RequestBody struct {
					UserID      string `json:"userID"`
					ProductName string `json:"productName"`
				}
				var ID UserCart

				//Get the userID and productName passed by the front end
				err = json.NewDecoder(request.Body).Decode(&RequestBody)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte(err.Error()))
					return
				}
				ID.UserID, err = strconv.Atoi(RequestBody.UserID)
				if err != nil {
					writer.WriteHeader(500)
					writer.Write([]byte("UserID must be a number"))
					return
				}
				ID.ProductName = RequestBody.ProductName

				//Call the function to delete the shopping cart item
				err = s.database.DeleteCartByName(ID.UserID, ID.ProductName)
				if err != nil {
					return
				}
				writer.WriteHeader(http.StatusOK)
				writer.Write([]byte("Delete The shopping cart product is successful"))

			},
		},
	}

	for _, v := range mallcartRoutes {
		addRouteWithVulInfo(MallGroup, v)
	}

}
