package rest

import (
	"dblayer"
	"log"
	"models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"github.com/stripe/stripe-go/customer"
)

type HandlerInterface interface {
	GetProducts(c *gin.Context)
	GetPromos(c *gin.Context)
	AddUser(c *gin.Context)
	SignIn(c *gin.Context)
	SignOut(c *gin.Context)
	GetOrders(c *gin.Context)
	Charge(c *gin.Context)
}

type Handler struct {
	db dblayer.DBLayer
}

func NewHandler() (*Handler, error) {
	//This creates a new pointer to the Handler object
	con := "root:cobra123@tcp(127.0.0.1:3306)/GoMusic"
	mysqlDb, err := dblayer.NewORM("mysql", con)

	if err != nil {
		return nil, err
	}
	return &Handler{db: mysqlDb}, nil
}

func (h *Handler) GetProducts(c *gin.Context) {

	if h.db == nil {
		return
	}

	products, err := h.db.GetAllProducts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, products)
}

func (h *Handler) GetPromos(c *gin.Context) {

	if h.db == nil {
		return
	}

	promos, err := h.db.GetPromos()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, promos)
}

func (h *Handler) SignIn(c *gin.Context) {

	if h.db == nil {
		return
	}

	var customer models.Customer
	err := c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customer, err = h.db.SignInUser(customer.FirstName, customer.LastName)
	if err != nil {
		if err == dblayer.ErrINVALIDPASSWORD {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, customer)
}

func (h *Handler) AddUser(c *gin.Context) {
	if h.db == nil {
		return
	}
	var customer models.Customer
	err := c.ShouldBindJSON(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	customer, err = h.db.AddUser(customer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, customer)
}

func (h *Handler) SignOut(c *gin.Context) {

	if h.db == nil {
		return
	}

	p := c.Param("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.db.SignOutUserById(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

}

func (h *Handler) GetOrders(c *gin.Context) {

	if h.db == nil {
		return
	}

	p := c.Param("id")
	id, err := strconv.Atoi(p)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	orders, err := h.db.GetCustomerOrdersById(int(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)

}

func (h *Handler) Charge(c *gin.Context) {

	if h.db == nil {
		return
	}

	request := struct {
		models.Order
		Remember    bool   `json:"rememberCard"`
		UseExisting bool   `json:"useExisting"`
		Token       string `json:"token"`
	}{}

	err := c.ShouldBindJSON(&request)
	if err != nil {
		c.JSON(http.StatusBadRequest, &request)
		return
	}

	stripe.Key = "sk_test_4eC39HqLyjWDarjtT1zdp7dc"

	chargeP := &stripe.ChargeParams{
		//the price we obtained from the incoming request:
		Amount: stripe.Int64(int64(request.Price)),
		//the currency:
		Currency: stripe.String("usd"),
		//the description:
		Description: stripe.String("GoMusic charge..."),
	}

	stripeCustomerID := ""

	if request.UseExisting {
		//use existing
		log.Println("Getting credit card id...")
		//This is a new method which retrieve the stripe customer id from the database
		stripeCustomerID, err = h.db.GetCreditCardCID(request.CustomerID)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		cp := &stripe.CustomerParams{}
		cp.SetSource(request.Token)
		customer, err := customer.New(cp)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		stripeCustomerID = customer.ID
	}

	if request.Remember {
		//save the stripe customer id, and link it to the actual customer id in our database
		err = h.db.SaveCreditCardForCustomer(request.CustomerID, stripeCustomerID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	/*
		we should check if the customer already ordered the same item or not but for simplicity, let's assume it's a new order
	*/

	//Assign the stipe customer id to the *stripe.ChargeParams object:
	chargeP.Customer = stripe.String(stripeCustomerID)
	//Charge the credit card
	_, err = charge.New(chargeP)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.db.AddOrder(request.Order)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
}
