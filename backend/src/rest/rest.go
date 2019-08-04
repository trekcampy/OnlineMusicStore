package rest

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func RunAPIWithHandler(address string, h HandlerInterface) error {
	//Get gin's default engine
	r := gin.Default()

	r.Use(MyCustomLogger1(), MyCustomLogger2())
	//get products
	r.GET("/products", h.GetProducts)
	//get promos
	r.GET("/promos", h.GetPromos)
	//post user sign in
	r.POST("/users/signin", h.SignIn)
	//add a user
	r.POST("/users", h.AddUser)
	//post user sign out
	r.POST("/user/:id/signout", h.SignOut)
	//get user orders
	r.GET("/user/:id/orders", h.GetOrders)
	//post purchase charge
	r.POST("/users/charge", h.Charge)
	//run the server
	return r.Run(address)
	//return r.RunTLS(address, "/tmp/cert.pem", "/tmp/key.pem")
}

func RunAPI(address string) error {
	h, err := NewHandler()
	if err != nil {
		return err
	}
	return RunAPIWithHandler(address, h)
}

func MyCustomLogger1() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("***************Prashant*********************")
		c.Next()
		fmt.Println("***************Prashant*********************")
	}
}

func MyCustomLogger2() gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("***************Desai*********************")
		c.Next()
		fmt.Println("***************Desai*********************")
	}
}
