package vulinbox

import (
	"github.com/jinzhu/gorm"
	// uuid "github.com/satori/go.uuid"
)

type UserCart struct {
	gorm.Model
	UserID          int     `gorm:"column:UserID"`          //User ID
	ProductName     string  `gorm:"column:ProductName"`     //Product name
	Description     string  `gorm:"column:Description"`     //Product description
	ProductPrice    float64 `gorm:"column:ProductPrice"`    //Product price
	ProductQuantity int     `gorm:"column:ProductQuantity"` //Product quantity
	TotalPrice      float64 `gorm:"column:TotalPrice"`      //Total price of the product
}

// Add to shopping cart
func (s *dbm) AddCart(UserID int, cart UserCart) (err error) {
	cart.UserID = UserID
	cart.ProductQuantity = 1 // Set the default product quantity to 1
	if err := s.db.Create(&cart).Error; err != nil {
		return err
	}
	return nil
}

// Add one to the number of products in the shopping cart
func (s *dbm) AddCartQuantity(UserID int, ProductName string) (err error) {
	var v UserCart
	v.UserID = UserID
	v.ProductName = ProductName
	if err := s.db.Model(&v).Where("UserID = ? AND ProductName = ?", v.UserID, v.ProductName).Update("ProductQuantity", gorm.Expr("ProductQuantity + ?", 1)).Error; err != nil {
		return err
	}
	return nil
}

// Shopping cart product The quantity is reduced by one. When the ProductQuantity quantity is reduced to 0, delete the product record
func (s *dbm) SubCartQuantity(UserID int, ProductName string) (err error) {
	var v UserCart
	v.UserID = UserID
	v.ProductName = ProductName
	if err := s.db.Model(&v).Where("UserID = ? AND ProductName = ?", v.UserID, v.ProductName).Update("ProductQuantity", gorm.Expr("ProductQuantity - ?", 1)).Error; err != nil {
		return err
	}
	//Query the current product quantity
	if err := s.db.Where("UserID = ? AND ProductName = ?", v.UserID, v.ProductName).First(&v).Error; err != nil {
		return err
	}
	//If the current product quantity is 0 , delete the product record

	if v.ProductQuantity == 0 {
		if err := s.db.Where("UserID = ? AND ProductName = ?", v.UserID, v.ProductName).Delete(&v).Error; err != nil {
			return err
		}
	}
	return nil
}

// Get the shopping cart
func (s *dbm) GetCart(UserID int) (userCart []UserCart, err error) {
	var v UserCart
	v.UserID = UserID
	if err := s.db.Where("UserID = ?", v.UserID).Find(&userCart).Error; err != nil {
		return nil, err
	}
	return userCart, nil
}

// Get the total number of products in the shopping cart
func (s *dbm) GetUserCartCount(UserID int) (count int64, err error) {
	var total struct {
		Total int64
	}
	if err := s.db.Raw("SELECT SUM(ProductQuantity) AS total FROM user_carts WHERE UserID = ?", UserID).Scan(&total).Error; err != nil {
		return 0, err
	}
	return total.Total, nil
}

// Delete the shopping cart through user_id and product name
func (s *dbm) DeleteCartByName(UserID int, ProductName string) (err error) {
	var v UserCart
	v.UserID = UserID
	v.ProductName = ProductName
	if err := s.db.Where("UserID = ? AND ProductName = ?", v.UserID, v.ProductName).Delete(&v).Error; err != nil {
		return err
	}
	return nil
}

// Check whether the shopping cart exists
func (s *dbm) CheckCart(UserID int, ProductName string) (bool, error) {
	var v UserCart
	v.UserID = UserID
	v.ProductName = ProductName
	err := s.db.Where("UserID = ? AND ProductName = ?", v.UserID, v.ProductName).First(&v).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// No matching record is found, return false
			return false, nil
		}
		// An error occurred during the query, return error
		return false, err
	}
	// Find a matching record, return true
	return true, nil
}
